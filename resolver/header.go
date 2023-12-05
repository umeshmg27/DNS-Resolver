package resolver

import "fmt"

// Header - DNS query header conssits of the following Data
type Header struct {
	ID                    uint16
	Flags                 uint16
	QuestionCount         uint16
	AnswerRecordCount     uint16
	AuthorityRecordCount  uint16
	AdditionalRecordCount uint16
}

func (h *Header) Encode() []byte {
	// Convert the Header struct into bytes
	buf := make([]byte, 12) // DNS header is always 12 bytes

	// Manually encode each uint16 into big-endian bytes
	buf[0] = byte(h.ID >> 8)
	buf[1] = byte(h.ID & 0xFF)
	buf[2] = byte(h.Flags >> 8)
	buf[3] = byte(h.Flags & 0xFF)
	buf[4] = byte(h.QuestionCount >> 8)
	buf[5] = byte(h.QuestionCount & 0xFF)
	buf[6] = byte(h.AnswerRecordCount >> 8)
	buf[7] = byte(h.AnswerRecordCount & 0xFF)
	buf[8] = byte(h.AuthorityRecordCount >> 8)
	buf[9] = byte(h.AuthorityRecordCount & 0xFF)
	buf[10] = byte(h.AdditionalRecordCount >> 8)
	buf[11] = byte(h.AdditionalRecordCount & 0xFF)

	return buf
}

func DecodeHeader(data []byte) (*Header, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("header data is too short")
	}

	return &Header{
		ID:                    uint16(data[0])<<8 | uint16(data[1]),
		Flags:                 uint16(data[2])<<8 | uint16(data[3]),
		QuestionCount:         uint16(data[4])<<8 | uint16(data[5]),
		AnswerRecordCount:     uint16(data[6])<<8 | uint16(data[7]),
		AuthorityRecordCount:  uint16(data[8])<<8 | uint16(data[9]),
		AdditionalRecordCount: uint16(data[10])<<8 | uint16(data[11]),
	}, nil

}

func VerifyHeader(responseHeader *Header, reqId uint16) error {
	if responseHeader.ID != reqId {

		return fmt.Errorf("\n\n Response and request header doesn't match")
	}

	fmt.Printf("\n\n responseHeader.Flags %+v \n", responseHeader)

	switch responseHeader.Flags & 0b1111 {
	case 1:
		return fmt.Errorf("There was a format error with the Query")

	case 2:
		return fmt.Errorf("Sever failure - server was unable to process the query.")

	case 3:
		return fmt.Errorf("This domain name does not exist.")
	}

	if responseHeader.AnswerRecordCount+responseHeader.AuthorityRecordCount+responseHeader.AdditionalRecordCount == 0 {
		return fmt.Errorf("No records available in the DNS records")
	}
	return nil

}
