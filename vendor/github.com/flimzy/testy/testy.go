package testy

import "testing"

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
