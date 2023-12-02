package resolver

// Question - DNS query header conssits of the following Data
type Question struct {
	Name  string
	Type  uint16
	Class uint16
}

// DNSMessage - Complete message
type DNSMessage struct {
	Header     Header
	Questions  []Question
	Answers    []ResourceRecord
	Authority  []ResourceRecord
	Additional []ResourceRecord
}
