package main

type DNSHeader struct {
	ID      uint16 // packet identifier
	QR      uint16 // query response
	OPCODE  uint16 // operation code
	AA      uint16 // auth answer
	TC      uint16 // tom cruise(!)
	RD      uint16 // recursion desired
	RA      uint16 // recursion available
	Z       uint16 // reserved
	RCODE   uint16 // response code
	QDCOUNT uint16 // question count
	ANCOUNT uint16 // answer record count
	NSCOUNT uint16 // auth record count
	ARCOUNT uint16 // additional record count
}
