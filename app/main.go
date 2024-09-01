package main

import (
	"fmt"
	"net"
)

var _ = net.ListenUDP

func main() {
	address := "127.0.0.1:2053"
	network := "udp"
	udpAddr, err := net.ResolveUDPAddr(network, address)

	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}
	fmt.Printf("running on PORT %d", udpAddr.Port)

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer func(udpConn *net.UDPConn) {
		err := udpConn.Close()
		if err != nil {
			fmt.Println("Failed to close UDP connection:", err)
		}
	}(udpConn)

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		fmt.Println(size, source, err)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])

		fmt.Println("start=======>")
		fmt.Println(string(buf)[:20])
		fmt.Println("**")
		fmt.Println(receivedData)
		fmt.Println("**")
		fmt.Println("end=======>")

		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		header := DNSHeader{
			ID:      1234, // Example ID
			QR:      1,    // Set QR to 1 for response
			OPCODE:  0,    // Standard query
			AA:      0,    // Not authoritative
			TC:      0,    // Not truncated
			RD:      0,    // Recursion not desired
			RA:      0,    // Recursion available
			Z:       0,    // Reserved
			RCODE:   0,    // No error
			QDCOUNT: 1,    // One question in the query
			ANCOUNT: 1,    // One answer
			NSCOUNT: 0,    // No authority records
			ARCOUNT: 0,    // No additional records
		}
		response := header.toBytes()

		questionSection := encodeDomainName("codecrafters.io")
		questionSection = append(questionSection, 0x00, 0x01) // QTYPE A
		questionSection = append(questionSection, 0x00, 0x01) // QCLASS

		response = append(response, questionSection...)

		answerSection := encodeDomainName("codecrafters.io")
		answerSection = append(answerSection, 0x00, 0x01)             // TYPE A
		answerSection = append(answerSection, 0x00, 0x01)             // CLASS IN
		answerSection = append(answerSection, 0x00, 0x00, 0x00, 0x3C) // TTL (60 seconds)
		answerSection = append(answerSection, 0x00, 0x04)             // Data length (4 bytes for IPv4)
		answerSection = append(answerSection, 0x08, 0x08, 0x08, 0x08) // RDATA (8.8.8.8)

		response = append(response, answerSection...)

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
