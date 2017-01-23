package studyhandler

import "testing"

type iframeTest struct {
	Input    string
	Expected string
}

func TestAddIframeID(t *testing.T) {
	tests := []iframeTest{
		iframeTest{
			Input:    `<html><head></head><body><h1>Foo!</h1></body></html>`,
			Expected: `<html><head><meta name="iframeid" content="__IFRAME__"/></head><body><h1>Foo!</h1></body></html>`,
		},
		iframeTest{
			Input:    `Don't look now!`,
			Expected: `<html><head><meta name="iframeid" content="__IFRAME__"/></head><body>Don&#39;t look now!</body></html>`,
		},
	}
	for _, test := range tests {
		result, err := addIframeID(test.Input, "__IFRAME__")
		if err != nil {
			t.Errorf("Unexpected error: %s\n", err)
		}
		if result != test.Expected {
			t.Errorf("Iframe not injected as expected\n\tExpected: %s\n\t  Actual: %s\n", test.Expected, result)
		}
	}
}
