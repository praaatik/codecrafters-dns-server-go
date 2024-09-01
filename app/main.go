package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	// Command-line arguments
	var resolver string
	flag.StringVar(&resolver, "resolver", "", "DNS resolver address in the form <ip>:<port>")
	flag.Parse()

	if resolver == "" {
		fmt.Println("Usage: ./your_server --resolver <address>")
		os.Exit(1)
	}

	// Resolve the DNS resolver address
	resolverAddr, err := net.ResolveUDPAddr("udp", resolver)
	if err != nil {
		fmt.Printf("Failed to resolve resolver address: %v\n", err)
		return
	}

	// Start the DNS server
	address := "127.0.0.1:2053"
	network := "udp"
	udpAddr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}
	fmt.Printf("Running on PORT %d\n", udpAddr.Port)

	udpConn, err := net.ListenUDP(network, udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		handleQuery(buf[:size], resolverAddr, udpConn, source)
	}
}

