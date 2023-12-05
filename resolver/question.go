package resolver

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// Question - DNS query header conssits of the following Data
type Question struct {
	Name  string
	Type  uint16
	Class uint16
}

const (
	TypeA   = 1 // Type A: Host address
	ClassIN = 1 // Class IN: Internet
)

func (q *Question) EncodeQuestion() ([]byte, error) {
	var buffer []byte

	//convert domain name into DNS format
	parts := strings.Split(q.Name, ".")
	for _, part := range parts {
		if len(part) > 63 {
			return nil, fmt.Errorf("part of the domain is too long: %s", part)
		}
		buffer = append(buffer, byte(len(part)))
		buffer = append(buffer, part...)
	}
	buffer = append(buffer, 0) // Null byte to end the domain name

	// Append type and class in big-endian format
	buffer = append(buffer, byte(q.Type>>8), byte(q.Type&0xFF))
	buffer = append(buffer, byte(q.Class>>8), byte(q.Class&0xFF))

	return buffer, nil
}

func DecodeQuestion(buffer []byte, startPosition int) (*Question, int, error) {
	name, size, _ := DecodeDomainName(buffer, startPosition)
	// if err != nil {
	// 	return nil, 0, err
	// }
	offset := startPosition + size
	body := Question{
		Name:  name,
		Type:  binary.BigEndian.Uint16(buffer[offset : offset+2]),
		Class: binary.BigEndian.Uint16(buffer[offset+2 : offset+4]),
	}

	// Return size of body since it varies with domain name length.
	return &body, size + 4, nil
}
