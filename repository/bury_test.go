package repo

import (
	"testing"

	fb "github.com/FlashbackSRS/flashback-model"
)

type buryTest struct {
	Bury     fb.Interval
	Interval fb.Interval
	New      bool
	Expected fb.Interval
}

func TestBuryInterval(t *testing.T) {
	tests := []buryTest{
		buryTest{
			Bury:     10 * fb.Day,
			Interval: 20 * fb.Day,
			New:      false,
			Expected: 4 * fb.Day,
		},
		buryTest{
			Bury:     10 * fb.Day,
			Interval: 20 * fb.Day,
			New:      true,
			Expected: 7 * fb.Day,
		},
		buryTest{
			Bury:     10 * fb.Day,
			Interval: 1 * fb.Day,
			New:      false,
			Expected: 1 * fb.Day,
		},
	}
	for _, test := range tests {
		result := buryInterval(test.Bury, test.Interval, test.New)
		if result != test.Expected {
			t.Errorf("%s / %s / %t:\n\tExpected: %s\n\t  Actual: %s\n", test.Bury, test.Interval, test.New, test.Expected, result)
		}
	}
}
