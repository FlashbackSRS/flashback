package fb

import "testing"

type validator interface {
	Validate() error
}

type validationTest struct {
	name string
	v    validator
	err  string
}

func testValidation(t *testing.T, tests []validationTest) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.v.Validate()
			checkErr(t, test.err, err)
		})
	}
}
