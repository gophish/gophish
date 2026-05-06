package dialer

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"testing"
)

func TestDefaultDeny(t *testing.T) {
	control := restrictedControl([]*net.IPNet{})
	host := "169.254.169.254"
	expected := fmt.Errorf("upstream connection denied to internal host at %s", host)
	conn := new(syscall.RawConn)
	got := control("tcp4", fmt.Sprintf("%s:80", host), *conn)
	if !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("unexpected error dialing denylisted host. expected %v got %v", expected, got)
	}
}

func TestDefaultAllow(t *testing.T) {
	control := restrictedControl([]*net.IPNet{})
	host := "1.1.1.1"
	conn := new(syscall.RawConn)
	got := control("tcp4", fmt.Sprintf("%s:80", host), *conn)
	if got != nil {
		t.Fatalf("error dialing allowed host. got %v", got)
	}
}

func TestCustomAllow(t *testing.T) {
	host := "127.0.0.1"
	_, ipRange, _ := net.ParseCIDR(fmt.Sprintf("%s/32", host))
	allowed := []*net.IPNet{ipRange}
	control := restrictedControl(allowed)
	conn := new(syscall.RawConn)
	got := control("tcp4", fmt.Sprintf("%s:80", host), *conn)
	if got != nil {
		t.Fatalf("error dialing allowed host. got %v", got)
	}
}

func TestCustomDeny(t *testing.T) {
	host := "127.0.0.1"
	_, ipRange, _ := net.ParseCIDR(fmt.Sprintf("%s/32", host))
	allowed := []*net.IPNet{ipRange}
	control := restrictedControl(allowed)
	conn := new(syscall.RawConn)
	expected := fmt.Errorf("upstream connection denied to internal host at %s", host)
	got := control("tcp4", "192.168.1.2:80", *conn)
	if !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("unexpected error dialing denylisted host. expected %v got %v", expected, got)
	}
}

func TestSingleIP(t *testing.T) {
	orig := DefaultDialer.AllowedHosts()
	host := "127.0.0.1"
	DefaultDialer.SetAllowedHosts([]string{host})
	control := DefaultDialer.Dialer().Control
	conn := new(syscall.RawConn)
	expected := fmt.Errorf("upstream connection denied to internal host at %s", host)
	got := control("tcp4", "192.168.1.2:80", *conn)
	if !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("unexpected error dialing denylisted host. expected %v got %v", expected, got)
	}

	host = "::1"
	DefaultDialer.SetAllowedHosts([]string{host})
	control = DefaultDialer.Dialer().Control
	conn = new(syscall.RawConn)
	expected = fmt.Errorf("upstream connection denied to internal host at %s", host)
	got = control("tcp4", "192.168.1.2:80", *conn)
	if !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("unexpected error dialing denylisted host. expected %v got %v", expected, got)
	}

	// Test an allowed connection
	got = control("tcp4", fmt.Sprintf("[%s]:80", host), *conn)
	if got != nil {
		t.Fatalf("error dialing allowed host. got %v", got)
	}
	DefaultDialer.SetAllowedHosts(orig)
}

// TestAllowedHostsDoesNotBlockExternal verifies that configuring
// allowed_internal_hosts does not accidentally block legitimate external
// destinations. This is a regression test for issue #9423, where setting
// allowed_internal_hosts caused all outbound connections (including external
// IPv4 addresses such as smtp.google.com) to be denied.
func TestAllowedHostsDoesNotBlockExternal(t *testing.T) {
	orig := DefaultDialer.AllowedHosts()
	defer DefaultDialer.SetAllowedHosts(orig)

	DefaultDialer.allowedHosts = nil
	if err := DefaultDialer.SetAllowedHosts([]string{"203.0.113.1/32"}); err != nil {
		t.Fatalf("error setting allowed hosts: %v", err)
	}
	control := DefaultDialer.Dialer().Control
	conn := new(syscall.RawConn)

	// External IPv4 address (smtp.google.com from issue #9423) should be
	// permitted even when allowed_internal_hosts is configured.
	got := control("tcp4", "142.251.127.108:587", *conn)
	if got != nil {
		t.Fatalf("external IPv4 host should not be denied, got %v", got)
	}

	// External IPv6 address (Google) should also be permitted.
	got = control("tcp6", "[2607:f8b0:4002:c06::1b]:443", *conn)
	if got != nil {
		t.Fatalf("external IPv6 host should not be denied, got %v", got)
	}

	// The explicitly allowed host should be reachable even though it falls
	// inside the TEST-NET-3 internal range.
	got = control("tcp4", "203.0.113.1:80", *conn)
	if got != nil {
		t.Fatalf("explicitly allowed internal host should not be denied, got %v", got)
	}

	// A different internal host that was not allowed should still be denied.
	got = control("tcp4", "127.0.0.1:80", *conn)
	if got == nil || !strings.Contains(got.Error(), "upstream connection denied") {
		t.Fatalf("disallowed internal host should be denied, got %v", got)
	}
}
