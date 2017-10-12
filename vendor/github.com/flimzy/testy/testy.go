package testy

import (
	"net/http"
	"regexp"
	"testing"
)

// Error compares actual.Error() against expected, and triggers an error if
// they do not match. If actual is non-nil, t.SkipNow() is called as well.
func Error(t *testing.T, expected string, actual error) {
	var err string
	if actual != nil {
		err = actual.Error()
	}
	if expected != err {
		t.Errorf("Unexpected error: %s", err)
	}
	if actual != nil {
		t.SkipNow()
	}
}

type statusCoder interface {
	StatusCode() int
}

// StatusCode returns the HTTP status code embedded in the error, or 500 if
// there is no specific status code.
func StatusCode(err error) int {
	if err == nil {
		return 0
	}
	if coder, ok := err.(statusCoder); ok {
		return coder.StatusCode()
	}
	return http.StatusInternalServerError
}

// StatusError compares actual.Error() and the embeded HTTP status code against
// expected, and triggers an error if they do not match. If actual is non-nil,
// t.SkipNow() is called as well.
func StatusError(t *testing.T, expected string, status int, actual error) {
	var err string
	var actualStatus int
	if actual != nil {
		err = actual.Error()
		actualStatus = StatusCode(actual)
	}
	if expected != err {
		t.Errorf("Unexpected error: %s", err)
	}
	if status != actualStatus {
		t.Errorf("Unexpected status code: %d", actualStatus)
	}
	if actual != nil {
		t.SkipNow()
	}
}

// ErrorRE compares actual.Error() against expected, which is treated as a
// regular expression, and triggers an error if they do not match. If actual is
// non-nil, t.SkipNow() is called as well.
func ErrorRE(t *testing.T, expected string, actual error) {
	var err string
	if actual != nil {
		err = actual.Error()
	}
	if !regexp.MustCompile(expected).MatchString(err) {
		t.Errorf("Unexpected error: %s", err)
	}
	if actual != nil {
		t.SkipNow()
	}
}
