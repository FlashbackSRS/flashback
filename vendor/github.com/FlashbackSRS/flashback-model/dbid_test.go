package fb

import (
	"testing"
)

func TestValidateDBID(t *testing.T) {
	type Test struct {
		name string
		id   string
		err  string
	}
	tests := []Test{
		{
			name: "bogus id",
			id:   "really really bogus",
			err:  "invalid DBID format",
		},
		{
			name: "unsupported type",
			id:   "foo-chicken",
			err:  "unsupported DBID type 'foo'",
		},
		{
			name: "invalid base64",
			id:   "bundle- really bad stuff",
			err:  "invalid DBID encoding",
		},
		{
			name: "valid",
			id:   "bundle-orsxg5bamrrgszak",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDBID(test.id)
			checkErr(t, test.err, err)
		})
	}
}

func TestEncodeDBID(t *testing.T) {
	expected := "foo-orsxg5banfsa"
	result := EncodeDBID("foo", []byte("test id"))
	if result != expected {
		t.Errorf("Unexpected result: %s", result)
	}
}
