package middleware

import (
	"context"
	"math"
	"net/http"
	"net/netip"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter limits how often each client may call the API, with one token
// bucket per client: a client may send burst requests at once and then
// perSecond more each second. A client that runs out is answered with 429 and
// a Retry-After header until its bucket refills; nobody is banned.
//
// Clients are identified by clientIP, so behind a reverse proxy the proxy's
// network must be listed as trusted. The buckets live in this process's
// memory: each instance of the service limits on its own.
type RateLimiter struct {
	perSecond      rate.Limit
	burst          int
	trustedProxies []netip.Prefix
	// now is the clock; tests replace it to control time.
	now func() time.Time

	mu      sync.Mutex
	clients map[netip.Addr]*clientBucket
}

type clientBucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter returns a RateLimiter that allows each client perSecond
// requests per second with bursts of up to burst requests.
func NewRateLimiter(perSecond float64, burst int, trustedProxies []netip.Prefix) *RateLimiter {
	return &RateLimiter{
		perSecond:      rate.Limit(perSecond),
		burst:          burst,
		trustedProxies: trustedProxies,
		now:            time.Now,
		clients:        make(map[netip.Addr]*clientBucket),
	}
}

// Middleware returns the middleware that enforces the limit, answering with
// 429 through writeError when a client has no tokens left.
func (rl *RateLimiter) Middleware(writeError ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(clientKey(clientIP(r, rl.trustedProxies))) {
				w.Header().Set(headerRetryAfter, strconv.Itoa(rl.retryAfterSeconds()))
				writeError(w, http.StatusTooManyRequests, codeRateLimited, messageRateLimited)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// allow takes one token from key's bucket, creating the bucket on first use.
func (rl *RateLimiter) allow(key netip.Addr) bool {
	now := rl.now()

	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket, found := rl.clients[key]
	if !found {
		bucket = &clientBucket{limiter: rate.NewLimiter(rl.perSecond, rl.burst)}
		rl.clients[key] = bucket
	}
	bucket.lastSeen = now
	return bucket.limiter.AllowN(now, 1)
}

// retryAfterSeconds is how long an empty bucket needs to earn one token,
// rounded up to whole seconds as Retry-After requires.
func (rl *RateLimiter) retryAfterSeconds() int {
	return int(math.Ceil(1 / float64(rl.perSecond)))
}

// Cleanup removes the buckets of clients that have been idle for longer than
// idleAfter, every interval, until ctx is canceled. A forgotten client simply
// starts again with a full bucket, which it would have earned by then anyway.
// Without it the map would grow with every client ever seen.
func (rl *RateLimiter) Cleanup(ctx context.Context, interval, idleAfter time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.removeIdle(idleAfter)
		}
	}
}

func (rl *RateLimiter) removeIdle(idleAfter time.Duration) {
	cutoff := rl.now().Add(-idleAfter)

	rl.mu.Lock()
	defer rl.mu.Unlock()
	for key, bucket := range rl.clients {
		if bucket.lastSeen.Before(cutoff) {
			delete(rl.clients, key)
		}
	}
}
