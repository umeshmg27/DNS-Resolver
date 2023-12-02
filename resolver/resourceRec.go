package resolver

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

// ResourceRecord - DNS query header conssits of the following Data
type ResourceRecord struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	rdLEngth uint16
	Data     []byte // This can be an IP address for A records, a hostname for CNAME, etc.
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
	ip := net.ParseIP(string(rr.Data))
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

// Decode Resource
func decodeResource(buffer []byte, startPosition int) (*Resource, int, error) {
	// Could either be a pointer, inlined name or combination.
	name, size := decodeDomainName(buffer, startPosition)
	offset := startPosition + size

	qType := binary.BigEndian.Uint16(buffer[offset : offset+2])
	qClass := binary.BigEndian.Uint16(buffer[offset+2 : 4+offset])
	ttl := binary.BigEndian.Uint32(buffer[offset+4 : offset+8])
	rdLength := binary.BigEndian.Uint16(buffer[8+offset : 10+offset])

	rData := []byte{}
	if qType == 2 && qClass == 1 {
		rData = []byte(decodeNSrData(buffer, buffer[10+offset:10+offset+int(rdLength)]))
	} else {
		rData = buffer[10+offset : 10+uint16(offset)+rdLength]
	}
	resource := ResourceRecord{

		name,
		qType,
		qClass,
		ttl,
		rdLength,
		rData,
	}

	// Return length of the section so that caller can update buffer position.
	endPosition := offset + 10 + int(rdLength)
	return &resource, endPosition - startPosition, nil
}
