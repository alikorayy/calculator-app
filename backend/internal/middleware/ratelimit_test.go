package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

const (
	testPerSecond = 1
	testBurst     = 3
)

// fakeClock is a manually advanced clock for the rate limiter.
type fakeClock struct{ current time.Time }

func (c *fakeClock) now() time.Time          { return c.current }
func (c *fakeClock) advance(d time.Duration) { c.current = c.current.Add(d) }
func newFakeClock() *fakeClock               { return &fakeClock{current: time.Unix(0, 0)} }
func (rl *RateLimiter) trackedClients() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.clients)
}
func okHandler(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

// newTestLimiter returns a limiter on a fake clock that trusts 172.16.0.0/12.
func newTestLimiter(clock *fakeClock) *RateLimiter {
	rl := NewRateLimiter(testPerSecond, testBurst, []netip.Prefix{netip.MustParsePrefix("172.16.0.0/12")})
	rl.now = clock.now
	return rl
}

// send makes one request from remoteAddr (with an optional X-Forwarded-For) and returns the response.
func send(rl *RateLimiter, remoteAddr, forwardedFor string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		req.Header.Set(headerForwardedFor, forwardedFor)
	}
	rec := httptest.NewRecorder()
	rl.Middleware(plainErrorWriter)(http.HandlerFunc(okHandler)).ServeHTTP(rec, req)
	return rec
}

func TestRateLimiter_AllowsABurstThenAnswers429(t *testing.T) {
	clock := newFakeClock()
	rl := newTestLimiter(clock)

	for i := range testBurst {
		if rec := send(rl, "85.1.1.1:5000", ""); rec.Code != http.StatusOK {
			t.Fatalf("request %d within the burst: status %d, want 200", i+1, rec.Code)
		}
	}

	rec := send(rl, "85.1.1.1:5000", "")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request after the burst: status %d, want 429", rec.Code)
	}
	if got := rec.Header().Get(headerRetryAfter); got != "1" {
		t.Errorf("Retry-After = %q, want %q", got, "1")
	}
	if want := codeRateLimited + ": " + messageRateLimited; rec.Body.String() != want {
		t.Errorf("body = %q, want %q", rec.Body.String(), want)
	}
}

func TestRateLimiter_TokensRefillOverTime(t *testing.T) {
	clock := newFakeClock()
	rl := newTestLimiter(clock)
	for range testBurst {
		send(rl, "85.1.1.1:5000", "")
	}

	clock.advance(time.Second / testPerSecond)

	if rec := send(rl, "85.1.1.1:5000", ""); rec.Code != http.StatusOK {
		t.Errorf("after one refill interval: status %d, want 200", rec.Code)
	}
	if rec := send(rl, "85.1.1.1:5000", ""); rec.Code != http.StatusTooManyRequests {
		t.Errorf("only one token was refilled: status %d, want 429", rec.Code)
	}
}

func TestRateLimiter_ClientsHaveSeparateBuckets(t *testing.T) {
	tests := []struct {
		name           string
		first, second  [2]string // remote address, X-Forwarded-For
		wantSameBucket bool
	}{
		{"different IPv4 clients", [2]string{"85.1.1.1:1", ""}, [2]string{"92.2.2.2:1", ""}, false},
		{"same IPv4 client, different ports", [2]string{"85.1.1.1:1", ""}, [2]string{"85.1.1.1:2", ""}, true},
		{"different clients behind the trusted proxy", [2]string{"172.18.0.3:1", "85.1.1.1"}, [2]string{"172.18.0.3:1", "92.2.2.2"}, false},
		{"one IPv6 /64 network is one client", [2]string{"[2001:db8::1]:1", ""}, [2]string{"[2001:db8::2]:1", ""}, true},
		{"different IPv6 /64 networks", [2]string{"[2001:db8:0:1::1]:1", ""}, [2]string{"[2001:db8:0:2::1]:1", ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := newTestLimiter(newFakeClock())
			for range testBurst {
				send(rl, tt.first[0], tt.first[1])
			}

			limited := send(rl, tt.second[0], tt.second[1]).Code == http.StatusTooManyRequests
			if limited != tt.wantSameBucket {
				t.Errorf("second client limited = %v, want %v", limited, tt.wantSameBucket)
			}
		})
	}
}

func TestRateLimiter_RetryAfterIsRoundedUpToWholeSeconds(t *testing.T) {
	tests := []struct {
		perSecond float64
		want      int
	}{
		{10, 1},
		{1, 1},
		{0.5, 2},
		{0.3, 4},
	}

	for _, tt := range tests {
		rl := NewRateLimiter(tt.perSecond, 1, nil)
		if got := rl.retryAfterSeconds(); got != tt.want {
			t.Errorf("perSecond %v: retryAfterSeconds = %d, want %d", tt.perSecond, got, tt.want)
		}
	}
}

func TestRateLimiter_RemoveIdleForgetsOnlyIdleClients(t *testing.T) {
	clock := newFakeClock()
	rl := newTestLimiter(clock)
	send(rl, "85.1.1.1:1", "")
	clock.advance(time.Minute)
	send(rl, "92.2.2.2:1", "")

	rl.removeIdle(30 * time.Second)

	if got := rl.trackedClients(); got != 1 {
		t.Errorf("tracked clients = %d, want 1 (only the recently seen one)", got)
	}
}

func TestRateLimiter_CleanupRunsUntilCancelled(t *testing.T) {
	rl := NewRateLimiter(testPerSecond, testBurst, nil)
	rl.allow(netip.MustParseAddr("85.1.1.1"))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		rl.Cleanup(ctx, time.Millisecond, 0)
		close(done)
	}()

	deadline := time.Now().Add(time.Second)
	for rl.trackedClients() != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := rl.trackedClients(); got != 0 {
		t.Errorf("tracked clients = %d after cleanup, want 0", got)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Cleanup did not return after the context was canceled")
	}
}
