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

// DecodeDomainName - Decodes the domain nmae from the given buffer
func DecodeDomainName(buffer []byte, offset int) (string, int, error) {
	var s strings.Builder
	idx := offset
	seen := make(map[int]bool)

	for {
		if idx >= len(buffer) {
			return "", 0, fmt.Errorf("buffer too short")
		}

		length := int(buffer[idx])
		if length == 192 { // Pointer
			if seen[idx] {
				return "", 0, fmt.Errorf("circular reference detected")
			}
			seen[idx] = true

			suffix, _, err := DecodeDomainName(buffer, int(buffer[idx+1]))
			if err != nil {
				return "", 0, err
			}
			s.WriteString(suffix)
			idx += 2
			break
		} else {
			if idx+1+length > len(buffer) {
				return "", 0, fmt.Errorf("buffer too short for expected length")
			}

			name := buffer[idx+1 : idx+1+length]
			idx += 1 + length

			s.Write(name) // Write the name part to the builder

			// Check for the end of the string or add a dot
			if buffer[idx] == 0x00 {
				idx++ // Move past the null byte
				break
			} else {
				s.WriteByte('.') // Add a dot for the next part of the domain
			}
		}
	}
	return s.String(), idx - offset, nil
}

// DecodeNameServerData - Decodes the Name server data in the given buffer
func DecodeNameServerData(buffer, rdata []byte) (string, error) {
	var s strings.Builder
	idx := 0

	for {
		if idx >= len(rdata) {
			return "", fmt.Errorf("rdata too short")
		}

		length := int(rdata[idx])
		if length == 192 { // Pointer
			suffix, _, err := DecodeDomainName(buffer, int(rdata[idx+1]))
			if err != nil {
				return "", fmt.Errorf("failed to decode domain name: %w", err)
			}
			s.WriteString(suffix)
			idx += 2
			break
		} else {
			if idx+1+length > len(rdata) {
				return "", fmt.Errorf("rdata too short for expected length")
			}

			name := rdata[idx+1 : idx+1+length]
			idx += 1 + length

			s.WriteString(string(name)) // Write the name part to the builder

			if rdata[idx] == 0x00 {
				idx++ // Move past the null byte
				break
			} else {
				s.WriteByte('.') // Add a dot for the next part of the domain
			}
		}
	}
	return s.String(), nil
}

// DecodeResource - Decodes the resource Data and return Decoded ResourceRecord
func DecodeResource(buffer []byte, startPosition int) (*ResourceRecord, int, error) {
	// Decode the domain name, handling pointers and inline names.
	name, size, err := DecodeDomainName(buffer, startPosition)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode domain name: %w", err)
	}
	offset := startPosition + size

	// Ensure buffer has enough data for the fixed-length fields.
	if offset+10 > len(buffer) {
		return nil, 0, fmt.Errorf("buffer too short for resource record header")
	}

	// Extracting the resource record fields.
	qType := binary.BigEndian.Uint16(buffer[offset : offset+2])
	qClass := binary.BigEndian.Uint16(buffer[offset+2 : 4+offset])
	ttl := binary.BigEndian.Uint32(buffer[offset+4 : offset+8])
	rdLength := binary.BigEndian.Uint16(buffer[8+offset : 10+offset])

	// Check if the buffer contains enough data for rData.
	if offset+10+int(rdLength) > len(buffer) {
		return nil, 0, fmt.Errorf("buffer too short for rData")
	}

	// Decode rData based on qType and qClass.
	var rData []byte
	if qType == 2 && qClass == 1 {
		decodedRData, err := DecodeNameServerData(buffer, buffer[10+offset:10+offset+int(rdLength)])
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decode name server data: %w", err)
		}
		rData = []byte(decodedRData)
	} else {
		rData = buffer[10+offset : 10+uint16(offset)+rdLength]
	}

	// Creating the ResourceRecord struct.
	resource := ResourceRecord{
		Name:     name,
		Type:     qType,
		Class:    qClass,
		TTL:      ttl,
		rdLEngth: rdLength,
		Data:     rData,
	}

	// Calculate the end position of the resource record in the buffer.
	endPosition := offset + 10 + int(rdLength)
	return &resource, endPosition - startPosition, nil
}
