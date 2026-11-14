package middleware

import (
	"net/http"
	"net/netip"
	"strings"
)

// clientIP returns the address of the client that sent r.
//
// The TCP peer address is used unless the peer is a trusted proxy. Only then is
// X-Forwarded-For read, from right to left: each proxy appends the address it
// received the request from, so the rightmost entry that is not itself a
// trusted proxy is the client. Entries further left were written by the client
// and could be forged, so they are never used. An unreadable entry stops the
// walk and the last trustworthy address is returned.
func clientIP(r *http.Request, trustedProxies []netip.Prefix) netip.Addr {
	client := peerAddr(r)
	if !isTrusted(client, trustedProxies) {
		return client
	}

	entries := strings.Split(strings.Join(r.Header.Values(headerForwardedFor), ","), ",")
	for i := len(entries) - 1; i >= 0; i-- {
		addr, err := netip.ParseAddr(strings.TrimSpace(entries[i]))
		if err != nil {
			return client
		}
		client = addr.Unmap()
		if !isTrusted(client, trustedProxies) {
			return client
		}
	}
	return client
}

// peerAddr returns the address of the TCP peer, or the zero Addr if
// r.RemoteAddr cannot be parsed.
func peerAddr(r *http.Request) netip.Addr {
	addrPort, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}
	}
	return addrPort.Addr().Unmap()
}

func isTrusted(addr netip.Addr, trustedProxies []netip.Prefix) bool {
	for _, prefix := range trustedProxies {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// clientKey is the rate-limiting key for addr. An IPv4 address is a client on
// its own. IPv6 hands every home or device a whole /64 network, so limiting
// single IPv6 addresses would let one client rotate through billions of them;
// the /64 network is used instead.
func clientKey(addr netip.Addr) netip.Addr {
	if !addr.Is6() {
		return addr
	}
	// Prefix only fails for an invalid address or prefix length; neither is possible here.
	prefix, _ := addr.Prefix(ipv6ClientPrefixBits)
	return prefix.Addr()
}
