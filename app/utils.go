package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

func (header *DNSHeader) toBytes() []byte {
	buffer := make([]byte, 12) // DNS header = 12 bytes

	binary.BigEndian.PutUint16(buffer[0:2], header.ID) // encode the header.ID
	binary.BigEndian.PutUint16(buffer[2:4], (header.QR<<15)|(header.OPCODE<<11)|(header.AA<<10)|(header.TC<<9)|(header.RD<<8)|(header.RA<<7)|(header.Z<<4)|(header.RCODE))
	binary.BigEndian.PutUint16(buffer[4:6], header.QDCOUNT)
	binary.BigEndian.PutUint16(buffer[6:8], header.ANCOUNT)
	binary.BigEndian.PutUint16(buffer[8:10], header.NSCOUNT)
	binary.BigEndian.PutUint16(buffer[10:12], header.ARCOUNT)

	return buffer
}

func encodeDomainName(domain string) []byte {
	encoded := []byte{}
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		encoded = append(encoded, byte(len(label)))
		encoded = append(encoded, label...)
	}

	// Null byte to end the domain name
	encoded = append(encoded, 0x00)

	return encoded
}

func parseDNSHeader(buf []byte) DNSHeader {
	id := binary.BigEndian.Uint16(buf[0:2])
	flags := binary.BigEndian.Uint16(buf[2:4])
	opcode := (flags >> 11) & 0x0F
	rd := (flags >> 8) & 0x01

	qdcount := binary.BigEndian.Uint16(buf[4:6])
	nscount := binary.BigEndian.Uint16(buf[8:10])
	arcount := binary.BigEndian.Uint16(buf[10:12])

	header := DNSHeader{
		ID:      id,
		QR:      1,      // Set QR to 1 for response
		OPCODE:  opcode, // Mimic OPCODE
		AA:      0,      // Not authoritative
		TC:      0,      // Not truncated
		RD:      rd,     // Mimic RD
		RA:      0,      // Recursion not available
		Z:       0,      // Reserved
		RCODE:   0,      // No error if standard query; else 4
		QDCOUNT: qdcount,
		ANCOUNT: 0, // Will be set dynamically
		NSCOUNT: nscount,
		ARCOUNT: arcount,
	}

	if opcode != 0 { // If not a standard query, set RCODE to 4
		header.RCODE = 4
	}

	return header
}

func parseQuestions(buf []byte, offset int, count int) ([]DNSQuestion, int) {
	questions := []DNSQuestion{}

	for i := 0; i < count; i++ {
		qname, newOffset := parseDomainName(buf, offset)
		qtype := buf[newOffset : newOffset+2]
		qclass := buf[newOffset+2 : newOffset+4]
		offset = newOffset + 4

		questions = append(questions, DNSQuestion{
			Name:  qname,
			Type:  qtype,
			Class: qclass,
		})
	}

	return questions, offset
}

func parseDomainName(buf []byte, offset int) (string, int) {
	labels := []string{}
	for {
		length := int(buf[offset])

		// Check if the label is a pointer
		if length&0xC0 == 0xC0 {
			pointer := int(binary.BigEndian.Uint16(buf[offset:offset+2]) & 0x3FFF)
			label, _ := parseDomainName(buf, pointer)
			labels = append(labels, label)
			offset += 2
			break
		}

		// Zero length means end of the name
		if length == 0 {
			offset++
			break
		}

		offset++
		labels = append(labels, string(buf[offset:offset+length]))
		offset += length
	}
	return strings.Join(labels, "."), offset
}

func forwardDNSQuery(query []byte, resolverAddr *net.UDPAddr) ([]byte, error) {
	conn, err := net.DialUDP("udp", nil, resolverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial resolver: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(query)
	if err != nil {
		return nil, fmt.Errorf("failed to send query to resolver: %v", err)
	}

	response := make([]byte, 512)
	_, _, err = conn.ReadFromUDP(response)
	if err != nil {
		return nil, fmt.Errorf("failed to receive response from resolver: %v", err)
	}

	return response, nil
}

func handleQuery(query []byte, resolverAddr *net.UDPAddr, udpConn *net.UDPConn, source *net.UDPAddr) {
	// Parse the DNS header
	header := parseDNSHeader(query[:12])

	// Parse questions
	questions, offset := parseQuestions(query, 12, int(header.QDCOUNT))

	if len(questions) > 1 {
		// Forward each question separately
		var responses [][]byte
		for i := 0; i < len(questions); i++ {
			// Create a DNS query for each question
			queryPart := query[:12]
			queryPart = append(queryPart, encodeDomainName(questions[i].Name)...)
			queryPart = append(queryPart, questions[i].Type...)
			queryPart = append(queryPart, questions[i].Class...)

			// Append the rest of the query (if applicable)
			if offset < len(query) {
				queryPart = append(queryPart, query[offset:]...)
			}

			// Forward the query to the resolver
			response, err := forwardDNSQuery(queryPart, resolverAddr)
			if err != nil {
				fmt.Println("Failed to forward query:", err)
				continue
			}
			responses = append(responses, response)
		}

		// Combine responses
		var combinedResponse []byte
		for _, res := range responses {
			combinedResponse = append(combinedResponse, res[12:]...) // Skip the header (12 bytes)
		}

		// Include the original header
		combinedHeader := header
		combinedHeader.ANCOUNT = uint16(len(responses))
		combinedHeader.QDCOUNT = uint16(len(questions))
		combinedResponseHeader := combinedHeader.toBytes()
		combinedResponse = append(combinedResponseHeader, combinedResponse...)

		// Send the combined response back to the client
		_, err := udpConn.WriteToUDP(combinedResponse, source)
		if err != nil {
			fmt.Println("Failed to send combined response:", err)
		}
		return
	}

	// Forward the query to the resolver
	response, err := forwardDNSQuery(query[:offset], resolverAddr)
	if err != nil {
		fmt.Println("Failed to forward query:", err)
		return
	}

	// Send the resolver's response back to the client
	_, err = udpConn.WriteToUDP(response, source)
	if err != nil {
		fmt.Println("Failed to send response:", err)
	}
}
