package resolver

import (
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

func DecodeQuestion(data []byte) (*Question, error) {
	var (
		name strings.Builder
		i    int
		qlen = len(data)
	)

	// Decode the domain name
	for i < qlen {
		length := int(data[i])
		i++
		if length == 0 {
			break
		}
		if length > 63 || i+length > qlen {
			return nil, fmt.Errorf("invalid domain name label in question")
		}
		name.WriteString(string(data[i : i+length]))
		i += length
		if i < qlen {
			name.WriteString(".")
		}
	}

	if i+4 > qlen { // 2 bytes for Type and 2 bytes for Class
		return nil, fmt.Errorf("insufficient data for question type and class")
	}

	question := &Question{
		Name:  name.String(),
		Type:  uint16(data[i])<<8 | uint16(data[i+1]),
		Class: uint16(data[i+2])<<8 | uint16(data[i+3]),
	}

	return question, nil
}
