package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestClientIP(t *testing.T) {
	proxies := []netip.Prefix{netip.MustParsePrefix("172.16.0.0/12"), netip.MustParsePrefix("10.0.0.0/8")}
	tests := []struct {
		name         string
		remoteAddr   string
		forwardedFor []string
		trusted      []netip.Prefix
		wantClientIP string
	}{
		{"direct client", "85.1.1.1:5000", nil, proxies, "85.1.1.1"},
		{"forged header from an untrusted peer is ignored", "85.1.1.1:5000", []string{"1.2.3.4"}, proxies, "85.1.1.1"},
		{"no trusted proxies configured", "172.18.0.3:5000", []string{"85.1.1.1"}, nil, "172.18.0.3"},
		{"trusted proxy forwards the client", "172.18.0.3:5000", []string{"85.1.1.1"}, proxies, "85.1.1.1"},
		{"forged entry left of the proxy's entry", "172.18.0.3:5000", []string{"1.2.3.4, 85.1.1.1"}, proxies, "85.1.1.1"},
		{"chain of trusted proxies", "172.18.0.3:5000", []string{"85.1.1.1, 10.0.0.2"}, proxies, "85.1.1.1"},
		{"header split over several lines", "172.18.0.3:5000", []string{"1.2.3.4", "85.1.1.1"}, proxies, "85.1.1.1"},
		{"trusted proxy without a header", "172.18.0.3:5000", nil, proxies, "172.18.0.3"},
		{"every entry is a trusted proxy", "172.18.0.3:5000", []string{"10.0.0.9, 10.0.0.2"}, proxies, "10.0.0.9"},
		{"unreadable entry stops the walk", "172.18.0.3:5000", []string{"85.1.1.1, not-an-ip"}, proxies, "172.18.0.3"},
		{"IPv4-mapped IPv6 peer", "[::ffff:172.18.0.3]:5000", []string{"85.1.1.1"}, proxies, "85.1.1.1"},
		{"IPv6 client", "[2001:db8::1]:5000", nil, proxies, "2001:db8::1"},
		{"unparseable remote address", "not-an-address", nil, proxies, "invalid IP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for _, value := range tt.forwardedFor {
				req.Header.Add(headerForwardedFor, value)
			}

			if got := clientIP(req, tt.trusted).String(); got != tt.wantClientIP {
				t.Errorf("clientIP = %s, want %s", got, tt.wantClientIP)
			}
		})
	}
}

func TestClientKey(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"85.1.1.1", "85.1.1.1"},
		{"2001:db8:1:2:aaaa:bbbb:cccc:dddd", "2001:db8:1:2::"},
		{"2001:db8:1:2::1", "2001:db8:1:2::"},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			if got := clientKey(netip.MustParseAddr(tt.addr)).String(); got != tt.want {
				t.Errorf("clientKey(%s) = %s, want %s", tt.addr, got, tt.want)
			}
		})
	}
}
