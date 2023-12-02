package resolver

import (
	"fmt"
	"net"
	"strings"
)

// ResourceRecord - DNS query header conssits of the following Data
type ResourceRecord struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  string // This can be an IP address for A records, a hostname for CNAME, etc.
}

func (rr *ResourceRecord) Encode() ([]byte, error) {
	var buffer []byte

	// Encode Name (using the same manual method as for Question)
	nameBuffer, err := EncodeDomainName(rr.Name)
	if err != nil {
		return nil, err
	}
	buffer = append(buffer, nameBuffer...)

	// Encode Type and Class manually
	buffer = append(buffer, byte(rr.Type>>8), byte(rr.Type&0xFF))
	buffer = append(buffer, byte(rr.Class>>8), byte(rr.Class&0xFF))

	// Encode TTL manually
	buffer = append(buffer,
		byte(rr.TTL>>24),
		byte(rr.TTL>>16),
		byte(rr.TTL>>8),
		byte(rr.TTL&0xFF),
	)

	// Encode RData for A record (IPv4 address)
	ip := net.ParseIP(rr.Data)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", rr.Data)
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil, fmt.Errorf("not an IPv4 address: %s", rr.Data)
	}

	// RDLength for IPv4 is always 4
	buffer = append(buffer, 0, 4)
	buffer = append(buffer, ipv4...)

	return buffer, nil
}

// EncodeDomainName manually encodes a domain name
func EncodeDomainName(domain string) ([]byte, error) {
	var buffer []byte

	parts := strings.Split(domain, ".")
	for _, part := range parts {
		if len(part) > 63 {
			return nil, fmt.Errorf("part of the domain name is too long: %s", part)
		}
		buffer = append(buffer, byte(len(part)))
		buffer = append(buffer, part...)
	}
	buffer = append(buffer, 0) // Null byte to end the domain name

	return buffer, nil
}
