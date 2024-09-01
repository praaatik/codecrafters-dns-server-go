package main

import (
	"encoding/binary"
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
	//ancount := binary.BigEndian.Uint16(buf[6:8])
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
		ANCOUNT: 1, // One answer
		NSCOUNT: nscount,
		ARCOUNT: arcount,
	}

	if opcode != 0 { // If not a standard query, set RCODE to 4
		header.RCODE = 4
	}

	return header
}

func parseQuestionSection(buf []byte) (string, uint16, uint16) {
	// Parse QNAME (domain name)
	var qnameParts []string
	offset := 0
	for {
		length := int(buf[offset])
		if length == 0 {
			offset++
			break
		}
		offset++
		qnameParts = append(qnameParts, string(buf[offset:offset+length]))
		offset += length
	}
	qname := strings.Join(qnameParts, ".")

	// Parse QTYPE (2 bytes)
	qtype := binary.BigEndian.Uint16(buf[offset : offset+2])
	offset += 2

	// Parse QCLASS (2 bytes)
	qclass := binary.BigEndian.Uint16(buf[offset : offset+2])

	return qname, qtype, qclass
}
