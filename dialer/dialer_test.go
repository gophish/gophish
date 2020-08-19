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
