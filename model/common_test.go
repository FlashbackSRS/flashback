package model

import (
	"math"
	"testing"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
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

func floatCompare(x, y float64) bool {
	return math.Abs(x-y) < 0.01
}

func parseDue(t *testing.T, ds string) fb.Due {
	d, err := fb.ParseDue(ds)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
