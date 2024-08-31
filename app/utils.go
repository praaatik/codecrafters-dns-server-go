package main

import "encoding/binary"

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
