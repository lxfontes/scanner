package scanner

import (
	"fmt"
	"math"
	"net"
	"sync"
	"time"
)

type Request struct {
	Address net.Addr
	retCh   chan Response
}

type Response struct {
	Address net.Addr
	Open    bool
	Error   error
}

type Reporter interface {
	ScanResponse(response Response)
}

type ScannerConfig struct {
	// Should not be too high otherwise will exhaust local ports (time_wait)
	Concurrency int
	Timeout     time.Duration
	Reporter    Reporter
}

type Scanner struct {
	Config    ScannerConfig
	queue     chan Request
	quit      chan bool
	closeOnce sync.Once
}

const MinPort = 1
const MaxPort = math.MaxUint16
const queueFactor = 1.5

func (r Response) String() string {
	reachable := "not reacheable"
	if r.Open {
		reachable = "reacheable"
	}

	return fmt.Sprintf("%s is %s", r.Address, reachable)
}

func NewScanner(cfg ScannerConfig) *Scanner {
	sc := &Scanner{
		Config: cfg,
		queue:  make(chan Request, int(float64(cfg.Concurrency)*queueFactor)),
		quit:   make(chan bool),
	}

	for i := 0; i < sc.Config.Concurrency; i++ {
		go sc.loop()
	}

	return sc
}

func (sc *Scanner) loop() {
	for {
		select {
		case req := <-sc.queue:
			req.retCh <- sc.Scan(req)
		case <-sc.quit:
			return
		}
	}
}

func (sc *Scanner) Stop() {
	sc.closeOnce.Do(func() {
		close(sc.quit)
	})
}

func (sc *Scanner) Scan(request Request) Response {
	resp := Response{
		Address: request.Address,
	}

	c, err := net.DialTimeout(request.Address.Network(), request.Address.String(), sc.Config.Timeout)

	if err != nil {
		resp.Error = err
		resp.Open = false
	} else {
		c.Close()
		resp.Open = true
	}

	return resp
}

func (sc *Scanner) ScanIP(ip string) {
	sc.ScanIPPortRange(ip, 0, math.MaxUint16)
}

// Naive implementation
// This will likely get blocked by rate limiters ( eb/ip tables, fail2ban, etc)
func (sc *Scanner) ScanIPPortRange(ip string, portStart int, portEnd int) {

	if portEnd > math.MaxUint16 {
		portEnd = math.MaxUint16
	}

	if portStart > portEnd {
		portStart = portEnd
	}

	expectedResults := portEnd - portStart + 1

	respCh := make(chan Response)

	go func() {
		for port := portStart; port <= portEnd; port++ {
			req := Request{
				retCh: respCh,
				Address: &net.TCPAddr{
					IP:   net.ParseIP(ip),
					Port: port,
				},
			}

			sc.queue <- req
		}
	}()

	for i := 0; i < expectedResults; i++ {
		select {
		case resp := <-respCh:
			sc.Config.Reporter.ScanResponse(resp)
		case <-sc.quit:
			return
		}
	}
}

func (sc *Scanner) ScanCIDR(cidr string) error {
	return sc.ScanCIDRPortRange(cidr, 0, math.MaxUint16)
}

func (sc *Scanner) ScanCIDRPortRange(cidr string, portStart int, portEnd int) error {
	ip, net, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	for ip := ip.Mask(net.Mask); net.Contains(ip); incrementIP(ip) {
		sc.ScanIPPortRange(ip.String(), portStart, portEnd)
	}

	return nil
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
