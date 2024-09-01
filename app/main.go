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

		/*

			// first 12 bits are for the header
			requestHeader := parseDNSHeader(buf[:12])

			// remaining are for the question
			questionName, _, _ := parseQuestionSection(buf[12:])
			//fmt.Printf("Received Query - Name: %s, Type: %d, Class: %d\n", questionName, questionType, questionClass)

			// to the future reading me, values are taken from the challenge itself.
			//response header is the same as the request header
			responseHeader := requestHeader
			response := responseHeader.toBytes()

			questionSection := encodeDomainName(questionName)
			questionSection = append(questionSection, 0x00, 0x01) // QTYPE A
			questionSection = append(questionSection, 0x00, 0x01) // QCLASS

			response = append(response, questionSection...)

			answerSection := encodeDomainName(questionName)
			answerSection = append(answerSection, 0x00, 0x01)             // TYPE A
			answerSection = append(answerSection, 0x00, 0x01)             // CLASS IN
			answerSection = append(answerSection, 0x00, 0x00, 0x00, 0x3C) // TTL (60 seconds)
			answerSection = append(answerSection, 0x00, 0x04)             // Data length (4 bytes for IPv4)
			answerSection = append(answerSection, 0x08, 0x08, 0x08, 0x08) // RDATA (8.8.8.8)

			response = append(response, answerSection...)
		*/

		requestHeader := parseDNSHeader(buf[:12])
		questions, _ := parseQuestions(buf, 12, int(requestHeader.QDCOUNT))

		// Prepare the response
		responseHeader := requestHeader
		responseHeader.QR = 1 // Set QR to 1 for response
		responseHeader.ANCOUNT = requestHeader.QDCOUNT

		response := responseHeader.toBytes()

		// Append each question back to the response (uncompressed)
		for _, q := range questions {
			response = append(response, encodeDomainName(q.Name)...)
			response = append(response, q.Type...)
			response = append(response, q.Class...)
		}

		// Append each answer section to the response (using a fixed IP like 1.1.1.1)
		for _, q := range questions {
			response = append(response, encodeDomainName(q.Name)...) // Answer Name (uncompressed)
			response = append(response, 0x00, 0x01)                  // TYPE A
			response = append(response, 0x00, 0x01)                  // CLASS IN
			response = append(response, 0x00, 0x00, 0x00, 0x3C)      // TTL (60 seconds)
			response = append(response, 0x00, 0x04)                  // Data length (4 bytes for IPv4)
			response = append(response, 0x01, 0x01, 0x01, 0x01)      // RDATA (1.1.1.1)
		}

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
