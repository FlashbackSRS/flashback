package fb

import (
	"testing"
	"time"
)

func init() {
	now = func() time.Time {
		return parseTime("2017-01-01T00:00:00Z")
	}
}

func parseTimePtr(src string) *time.Time {
	t := parseTime(src)
	return &t
}

func parseTime(src string) time.Time {
	t, err := time.Parse(time.RFC3339, src)
	if err != nil {
		panic(err)
	}
	return t
}

func checkErr(t *testing.T, expected interface{}, err error) {
	var expectedMsg, errMsg string
	switch e := expected.(type) {
	case error:
		if e == err {
			return
		}
		if e != nil {
			expectedMsg = e.Error()
		}
	case string:
		expectedMsg = e
	case nil:
		// use empty string
	default:
		t.Fatalf("Unexpected type error type %T", expected)
	}
	if err != nil {
		errMsg = err.Error()
	}
	if expectedMsg != errMsg {
		t.Errorf("Unexpected error: %s", errMsg)
	}
}

func parseInterval(i string) Interval {
	iv, err := ParseInterval(i)
	if err != nil {
		panic(err)
	}
	return iv
}

func parseDue(d string) Due {
	du, err := ParseDue(d)
	if err != nil {
		panic(err)
	}
	return du
}
