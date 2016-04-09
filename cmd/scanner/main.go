package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lxfontes/scanner"
)

type Printer struct {
	verbose bool
}

func (p Printer) ScanResponse(resp scanner.Response) {
	portName, _ := scanner.LookupPort(resp.Address)

	if !p.verbose {
		if resp.Open {
			fmt.Println(resp.Address.String(), "[", portName, "] is open")
		}
		return
	}

	//verbose output
	if resp.Error != nil {
		// this should go to stderr actually
		fmt.Println("error:", resp.Address.String(), resp.Error)
	}

	if resp.Open {
		fmt.Println(resp.Address.String(), portName, "is open")
	} else {
		fmt.Println(resp.Address.String(), portName, "is closed")
	}
}

func main() {
	var (
		timeout     = flag.Duration("timeout", 1*time.Second, "TCP Timeout")
		concurrency = flag.Int("concurrency", 10, "Concurrent connections")
		portStart   = flag.Int("start", scanner.MinPort, "Port Range Start")
		portEnd     = flag.Int("end", scanner.MaxPort, "Port Range End")
		verbose     = flag.Bool("verbose", false, "Verbose Output")
	)

	flag.Usage = func() {
		fmt.Println(os.Args[0], "[options]", "[cidr] [cidr] [cidr]")
		fmt.Println("Example")
		fmt.Println("\t", os.Args[0], "127.0.0.1/32 192.168.88.0/24 2001:DB8::/48")
		fmt.Println()
		fmt.Println("Options")
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg := scanner.ScannerConfig{
		Timeout:     *timeout,
		Concurrency: *concurrency,
		Reporter:    Printer{verbose: *verbose},
	}

	if *portStart < 1 || *portEnd < *portStart || *portEnd > scanner.MaxPort {
		fmt.Println("Invalid port start/end combination")
		fmt.Println("Range", scanner.MinPort, "to", scanner.MaxPort)
		os.Exit(1)
	}

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	sc := scanner.NewScanner(cfg)
	for _, cidr := range flag.Args() {
		sc.ScanCIDRPortRange(cidr, *portStart, *portEnd)
	}
}
