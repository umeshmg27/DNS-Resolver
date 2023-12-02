package resolver

// ResourceRecord - DNS query header conssits of the following Data
type ResourceRecord struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  string // This can be an IP address for A records, a hostname for CNAME, etc.
}
