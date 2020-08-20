package dialer

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

// RestrictedDialer is used to create a net.Dialer which restricts outbound
// connections to only allowlisted IP ranges.
type RestrictedDialer struct {
	allowedHosts []*net.IPNet
}

// DefaultDialer is a global instance of a RestrictedDialer
var DefaultDialer = &RestrictedDialer{}

// SetAllowedHosts sets the list of allowed hosts or IP ranges for the default
// dialer.
func SetAllowedHosts(allowed []string) {
	DefaultDialer.SetAllowedHosts(allowed)
}

// AllowedHosts returns the configured hosts that are allowed for the dialer.
func (d *RestrictedDialer) AllowedHosts() []string {
	ranges := []string{}
	for _, ipRange := range d.allowedHosts {
		ranges = append(ranges, ipRange.String())
	}
	return ranges
}

// SetAllowedHosts sets the list of allowed hosts or IP ranges for the dialer.
func (d *RestrictedDialer) SetAllowedHosts(allowed []string) error {
	for _, ipRange := range allowed {
		// For flexibility, try to parse as an IP first since this will
		// undoubtedly cause issues. If it works, then just append the
		// appropriate subnet mask, then parse as CIDR
		if singleIP := net.ParseIP(ipRange); singleIP != nil {
			if singleIP.To4() != nil {
				ipRange += "/32"
			} else {
				ipRange += "/128"
			}
		}
		_, parsed, err := net.ParseCIDR(ipRange)
		if err != nil {
			return fmt.Errorf("provided ip range is not valid CIDR notation: %v", err)
		}
		d.allowedHosts = append(d.allowedHosts, parsed)
	}
	return nil
}

// Dialer returns a net.Dialer that restricts outbound connections to only the
// addresses allowed by the DefaultDialer.
func Dialer() *net.Dialer {
	return DefaultDialer.Dialer()
}

// Dialer returns a net.Dialer that restricts outbound connections to only the
// allowed addresses over TCP.
//
// By default, since Gophish anticipates connections originating to hosts on
// the local network, we only deny access to the link-local addresses at
// 169.254.0.0/16.
//
// If hosts are provided, then Gophish blocks access to all local addresses
// except the ones provided.
//
// This implementation is based on the blog post by Andrew Ayer at
// https://www.agwa.name/blog/post/preventing_server_side_request_forgery_in_golang
func (d *RestrictedDialer) Dialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   restrictedControl(d.allowedHosts),
	}
}

// defaultDeny represents the list of IP ranges that we want to block unless
// explicitly overriden.
var defaultDeny = []string{
	"169.254.0.0/16", // Link-local (used for VPS instance metadata)
}

// allInternal represents all internal hosts such that the only connections
// allowed are external ones.
var allInternal = []string{
	"0.0.0.0/8",
	"127.0.0.0/8",        // IPv4 loopback
	"10.0.0.0/8",         // RFC1918
	"100.64.0.0/10",      // CGNAT
	"172.16.0.0/12",      // RFC1918
	"169.254.0.0/16",     // RFC3927 link-local
	"192.88.99.0/24",     // IPv6 to IPv4 Relay
	"192.168.0.0/16",     // RFC1918
	"198.51.100.0/24",    // TEST-NET-2
	"203.0.113.0/24",     // TEST-NET-3
	"224.0.0.0/4",        // Multicast
	"240.0.0.0/4",        // Reserved
	"255.255.255.255/32", // Broadcast
	"::/0",               // Default route
	"::/128",             // Unspecified address
	"::1/128",            // IPv6 loopback
	"::ffff:0:0/96",      // IPv4 mapped addresses.
	"::ffff:0:0:0/96",    // IPv4 translated addresses.
	"fe80::/10",          // IPv6 link-local
	"fc00::/7",           // IPv6 unique local addr
}

type dialControl = func(network, address string, c syscall.RawConn) error

type restrictedDialer struct {
	*net.Dialer
	allowed []string
}

func restrictedControl(allowed []*net.IPNet) dialControl {
	return func(network string, address string, conn syscall.RawConn) error {
		if !(network == "tcp4" || network == "tcp6") {
			return fmt.Errorf("%s is not a safe network type", network)
		}

		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("%s is not a valid host/port pair: %s", address, err)
		}

		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("%s is not a valid IP address", host)
		}

		denyList := defaultDeny
		if len(allowed) > 0 {
			denyList = allInternal
		}

		for _, ipRange := range allowed {
			if ipRange.Contains(ip) {
				return nil
			}
		}

		for _, ipRange := range denyList {
			_, parsed, err := net.ParseCIDR(ipRange)
			if err != nil {
				return fmt.Errorf("error parsing denied range: %v", err)
			}
			if parsed.Contains(ip) {
				return fmt.Errorf("upstream connection denied to internal host")
			}
		}
		return nil
	}
}
