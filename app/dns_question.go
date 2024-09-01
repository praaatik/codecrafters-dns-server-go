package main

type DNSQuestion struct {
	Name  string
	Type  []byte
	Class []byte
}
