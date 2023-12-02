package resolver

import (
	"reflect"
	"testing"
)

func TestHeaderEncode(t *testing.T) {
	header := Header{
		ID:                    0xABCD,
		Flags:                 0x0100,
		QuestionCount:         1,
		AnswerRecordCount:     0,
		AuthorityRecordCount:  0,
		AdditionalRecordCount: 0,
	}

	encoded := header.Encode()
	expected := []byte{0xAB, 0xCD, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Header.Encode() = %v, want %v", encoded, expected)
	}
}

func TestHeaderDecode(t *testing.T) {
	data := []byte{0xAB, 0xCD, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	expected := Header{
		ID:                    0xABCD,
		Flags:                 0x0100,
		QuestionCount:         1,
		AnswerRecordCount:     0,
		AuthorityRecordCount:  0,
		AdditionalRecordCount: 0,
	}

	header, err := DecodeHeader(data)
	if err != nil {
		t.Fatalf("DecodeHeader() error = %v", err)
	}

	if !reflect.DeepEqual(*header, expected) {
		t.Errorf("DecodeHeader() = %v, want %v", *header, expected)
	}
}

func TestHeaderEncodeDecode(t *testing.T) {
	original := Header{
		ID:                    0xABCD,
		Flags:                 0x0100,
		QuestionCount:         1,
		AnswerRecordCount:     0,
		AuthorityRecordCount:  0,
		AdditionalRecordCount: 0,
	}

	encoded := original.Encode()
	decoded, err := DecodeHeader(encoded)
	if err != nil {
		t.Fatalf("DecodeHeader() error = %v", err)
	}

	if !reflect.DeepEqual(*decoded, original) {
		t.Errorf("DecodeHeader() = %v, want %v", *decoded, original)
	}
}
