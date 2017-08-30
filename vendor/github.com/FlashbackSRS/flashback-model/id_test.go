package fb

import (
	"testing"
)

func TestValidateDocID(t *testing.T) {
	type Test struct {
		name string
		id   string
		err  string
	}
	tests := []Test{
		{
			name: "bogus id",
			id:   "really really bogus",
			err:  "invalid DocID format",
		},
		{
			name: "unsupported type",
			id:   "foo-chicken",
			err:  "unsupported DocID type 'foo'",
		},
		{
			name: "invalid base64",
			id:   "deck- really bad stuff",
			err:  "invalid DocID encoding",
		},
		{
			name: "valid",
			id:   "deck-0123456789",
		},
		{
			name: "multiple dashes",
			id:   "deck--v4-v4",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDocID(test.id)
			checkErr(t, test.err, err)
		})
	}
}

func TestEncodeDocID(t *testing.T) {
	expected := "foo-dGVzdCBpZA"
	result := EncodeDocID("foo", []byte("test id"))
	if result != expected {
		t.Errorf("Unexpected result: %s", result)
	}
}
