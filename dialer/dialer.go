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

// SetAllowedHosts replaces the list of allowed hosts or IP ranges for the
// default dialer.
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

// SetAllowedHosts replaces the list of allowed hosts or IP ranges for the
// dialer. The provided list is only applied if every entry is valid.
func (d *RestrictedDialer) SetAllowedHosts(allowed []string) error {
	parsedHosts := []*net.IPNet{}
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
		parsedHosts = append(parsedHosts, parsed)
	}
	d.allowedHosts = parsedHosts
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
// By default, Gophish denies access to all internal addresses (loopback,
// RFC1918, link-local such as the cloud metadata endpoints, CGNAT, multicast,
// etc.). Hosts or IP ranges explicitly configured as allowed are exempted
// from this restriction via SetAllowedHosts.
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

// allInternal represents all internal IP ranges, which are always denied
// unless explicitly listed as allowed. IPv4-mapped IPv6 addresses are
// normalized to IPv4 before matching, so they are covered by the IPv4 ranges.
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
	"::/128",             // Unspecified address
	"::1/128",            // IPv6 loopback
	"fe80::/10",          // IPv6 link-local
	"fc00::/7",           // IPv6 unique local addr
	"2001::/32",          // Teredo (embeds IPv4)
	"2002::/16",          // 6to4 (embeds IPv4)
	"64:ff9b::/96",       // NAT64 well-known prefix (embeds IPv4)
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
		// Normalize IPv4-mapped IPv6 addresses so they are matched against
		// the IPv4 ranges rather than the IPv6 ones.
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}

		for _, ipRange := range allowed {
			if ipRange.Contains(ip) {
				return nil
			}
		}

		for _, ipRange := range allInternal {
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
