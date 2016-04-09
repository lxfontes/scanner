package scanner

import (
	"net"
	"testing"
	"time"
)

type dummyPrinter struct {
	t             *testing.T
	openCounter   int
	closedCounter int
}

func (d *dummyPrinter) ScanResponse(resp Response) {
	//d.t.Log(resp)

	if resp.Open {
		d.openCounter++
	} else {
		d.closedCounter++
	}
}

func ephemeralListener(ip net.IP) (*net.TCPListener, error) {
	// request an ephemeral port
	lAddr := &net.TCPAddr{
		IP:   ip,
		Port: 0,
	}

	l, err := net.ListenTCP("tcp", lAddr)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func TestServices(t *testing.T) {
	harness := map[int]string{
		80:    "http",
		110:   "pop3",
		143:   "imap",
		49150: "49150", // unknown ports should return string representation of int port
	}

	for port, service := range harness {
		srv := ServiceLookup(port)
		if srv != service {
			t.Fatal(port, "not resolving to", service)
		}
	}

	e, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:110")
	portName, portNumber := LookupPort(e)

	if portName != "pop3" {
		t.Fatal("port 110 didn't resolve to pop3")
	}

	if portNumber != 110 {
		t.Fatal("port number", portNumber, "is not", 110)
	}
}

func TestIpIncrement(t *testing.T) {
	harness := map[string]string{
		"127.0.0.1":       "127.0.0.2",
		"127.0.0.255":     "127.0.1.0",
		"127.0.255.255":   "127.1.0.0",
		"127.255.255.255": "128.0.0.0",
	}

	for ip, target := range harness {
		sIP := net.ParseIP(ip)
		targetIP := net.ParseIP(target)
		incrementIP(sIP)

		if !sIP.Equal(targetIP) {
			t.Fatal("Failed to increment ip", ip, "expected", target, "got", sIP)
		}
	}
}

func TestScannerOpen(t *testing.T) {
	l, err := ephemeralListener(net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		if c, err := l.Accept(); err != nil {
			t.Fatal(err)
		} else {
			c.Close()
		}
	}()

	dummy := dummyPrinter{t: t}
	cfg := ScannerConfig{
		Concurrency: 2,
		Timeout:     5 * time.Second,
		Reporter:    &dummy,
	}

	lPort := l.Addr().(*net.TCPAddr).Port
	sc := NewScanner(cfg)
	sc.ScanCIDRPortRange("127.0.0.1/32", lPort, lPort)
	if dummy.openCounter != 1 {
		t.Fatal("Port", lPort, "should be accessible", dummy.openCounter)
	}
}

func TestScannerClosed(t *testing.T) {
	l, err := ephemeralListener(net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	// simulated a closed port by not accepting a connection, OS will not reuse this port for a while (tw_recycle/tw_reuse)
	l.Close()

	dummy := dummyPrinter{t: t}
	cfg := ScannerConfig{
		Concurrency: 2,
		Timeout:     5 * time.Second,
		Reporter:    &dummy,
	}

	lPort := l.Addr().(*net.TCPAddr).Port
	sc := NewScanner(cfg)
	sc.ScanCIDRPortRange("127.0.0.1/32", lPort, lPort)
	if dummy.closedCounter != 1 {
		t.Fatal("Port", lPort, "should not be accessible", dummy.closedCounter, dummy.openCounter)
	}
}

func TestScannerIPv6(t *testing.T) {
	l, err := ephemeralListener(net.ParseIP("::1"))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		if c, err := l.Accept(); err != nil {
			t.Fatal(err)
		} else {
			c.Close()
		}
	}()

	dummy := dummyPrinter{t: t}
	cfg := ScannerConfig{
		Concurrency: 2,
		Timeout:     5 * time.Second,
		Reporter:    &dummy,
	}

	lPort := l.Addr().(*net.TCPAddr).Port
	sc := NewScanner(cfg)
	sc.ScanCIDRPortRange("::1/128", lPort, lPort)
	if dummy.openCounter != 1 {
		t.Fatal("Port", lPort, "should be accessible", dummy.openCounter)
	}
}
