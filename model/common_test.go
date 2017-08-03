package model

import (
	"testing"
	"time"
)

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

func parseTime(t *testing.T, str string) time.Time {
	ts, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}
