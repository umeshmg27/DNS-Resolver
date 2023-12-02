package resolver

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
	// Convert bytes into a Header struct
	return nil, nil
}
