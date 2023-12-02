package resolver

import (
	"reflect"
	"testing"
)

func TestQuestionEncode(t *testing.T) {
	question := Question{
		Name:  "www.example.com",
		Type:  TypeA,
		Class: ClassIN,
	}

	encoded, err := question.EncodeQuestion()
	if err != nil {
		t.Fatalf("Question.Encode() error = %v", err)
	}

	// This is the expected byte slice for the encoded question.
	// Adjust it according to the expected DNS format.
	expected := []byte{3, 'w', 'w', 'w', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0, 0, TypeA, 0, ClassIN}

	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Question.Encode() = %v, want %v", encoded, expected)
	}
}
