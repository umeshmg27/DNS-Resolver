package resolver

// import (
// 	"reflect"
// 	"testing"
// )

// func TestResourceRecordEncode(t *testing.T) {
// 	rr := ResourceRecord{
// 		Name:  "www.example.com",
// 		Type:  TypeA,
// 		Class: ClassIN,
// 		TTL:   300,
// 		Data:  "1.2.3.4",
// 	}

// 	encoded, err := rr.Encode()
// 	if err != nil {
// 		t.Fatalf("ResourceRecord.Encode() error = %v", err)
// 	}

// 	// This expected byte slice includes the encoded domain name, type, class, TTL, RDLength, and RData.
// 	// The exact byte slice will depend on how you've implemented the domain name encoding.
// 	expected := []byte{
// 		// ... bytes representing the encoded domain name "www.example.com"
// 		0x00, 0x01, // Type A
// 		0x00, 0x01, // Class IN
// 		0x00, 0x00, 0x01, 0x2c, // TTL (300 seconds)
// 		0x00, 0x04, // RDLength (4 bytes for IPv4)
// 		0x01, 0x02, 0x03, 0x04, // RData (1.2.3.4)
// 	}

// 	if !reflect.DeepEqual(encoded, expected) {
// 		t.Errorf("ResourceRecord.Encode() = %v, want %v", encoded, expected)
// 	}
// }
