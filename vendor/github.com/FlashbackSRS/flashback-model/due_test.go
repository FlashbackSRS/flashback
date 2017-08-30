package fb

import (
	"fmt"
	"testing"
	"time"

	"github.com/flimzy/diff"
)

func TestParseDue(t *testing.T) {
	if _, err := ParseDue("foobar"); err == nil {
		t.Errorf("Expected an error for invalid input to ParseDue")
	}

	expectedDue := parseTime("2017-01-01T00:00:00Z")
	result, err := ParseDue("2017-01-01")
	if err != nil {
		t.Errorf("Error parsing date-formatted Due value: %s", err)
	}
	if !expectedDue.Equal(result.Time()) {
		t.Errorf("Due = %s, expected %s\n", result, expectedDue)
	}

	expectedDue = parseTime("2017-01-01T12:30:45Z")
	result, err = ParseDue("2017-01-01 12:30:45")
	if err != nil {
		t.Errorf("Error parsing time-formatted Due value: %s", err)
	}
	if !expectedDue.Equal(result.Time()) {
		t.Errorf("Due = %s, expected %s\n", result, expectedDue)
	}
}

type StringerTest struct {
	Name     string
	I        fmt.Stringer
	Expected string
}

func TestStringer(t *testing.T) {
	tests := []StringerTest{
		{
			Name:     "Interval seconds",
			I:        Interval(100 * Second),
			Expected: "100s",
		},
		{
			Name:     "Interval seconds, plus nanoseconds",
			I:        Interval(10*Second + 1),
			Expected: "10s",
		},
		{
			Name:     "Interval seconds",
			I:        Interval(5 * Minute),
			Expected: "5m",
		},
		{
			Name:     "Interval seconds",
			I:        Interval(6 * Hour),
			Expected: "6h",
		},
		{
			Name:     "Interval days",
			I:        Interval(100 * Hour),
			Expected: "5d",
		},
		{
			Name:     "Due seconds",
			I:        Due(parseTime("2017-01-17T00:01:40Z")),
			Expected: "2017-01-17 00:01:40",
		},
		{
			Name:     "Due days",
			I:        Due(parseTime("1970-04-11T00:00:00Z")),
			Expected: "1970-04-11",
		},
	}
	for _, test := range tests {
		if result := test.I.String(); result != test.Expected {
			t.Errorf("%s:\n\tExpected '%s'\n\t  Actual: '%s'\n", test.Name, test.Expected, result)
		}
	}
}

func TestDueIn(t *testing.T) {
	result := DueIn(10 * Minute)
	expected := "2017-01-01 00:10:00"
	if result.String() != expected {
		t.Errorf("Due in 10 minutes:\n\tExpected: %s\n\t  Actual: %s\n", expected, result)
	}

	result = DueIn(15 * Day)
	expected = "2017-01-16"
	if result.String() != expected {
		t.Errorf("Due in 15 days:\n\tExpected: %s\n\t  Actual: %s\n", expected, result)
	}
}

func TestAdd(t *testing.T) {
	result := DueIn(3 * Hour).Add(Interval(9000 * Second))
	expected := "2017-01-01 05:30:00"
	if result.String() != expected {
		t.Errorf("Add 9000 seconds2017-01-01 05:30:00:\n\tExpected: %s\n\t  Actual: %s\n", expected, result)
	}

	result = DueIn(3 * Hour).Add(Interval(24 * Hour))
	expected = "2017-01-02"
	if result.String() != expected {
		t.Errorf("Add 24 hours:\n\tExpected: %s\n\t  Actual: %s\n", expected, result)
	}

	result = DueIn(0).Add(Interval(24 * Hour))
	expected = "2017-01-02"
	if result.String() != expected {
		t.Errorf("Add 24 hours:\n\tExpected: %s\n\t  Actual: %s\n", expected, result)
	}

	result = DueIn(3 * Hour).Add(Interval(9000 * Hour))
	expected = "2018-01-11"
	if result.String() != expected {
		t.Errorf("Add 9000 hours:\n\tExpected: %s\n\t  Actual: %s\n", expected, result)
	}
}

func TestOn(t *testing.T) {
	ts, e := time.Parse(time.RFC3339, "2016-01-01T01:01:01+00:00")
	if e != nil {
		t.Fatal(e)
	}
	d := On(ts)
	expected, e := time.Parse("2006-01-02", "2016-01-01")
	if e != nil {
		t.Fatal(e)
	}
	if !time.Time(d).Equal(expected) {
		t.Errorf("Unexpected result: %v", d)
	}
}

func TestNow(t *testing.T) {
	now := time.Now()
	n := Now()
	if s := time.Time(n).Sub(now).Seconds(); s > 0.000001 {
		t.Errorf("Result differs by %fs", s)
	}
}

func TestToday(t *testing.T) {
	today := Today()
	expected := now().Truncate(time.Duration(Day))
	if !expected.Equal(time.Time(today)) {
		t.Errorf("Unepxected result: %v", today)
	}
}

func TestDueSub(t *testing.T) {
	d := Due(parseTime("2017-01-02T00:00:00Z"))
	result := d.Sub(Due(parseTime("2017-01-01T00:00:00Z")))
	expected := Interval(24 * time.Hour)
	if result != expected {
		t.Errorf("Unexpected result: %v", result)
	}
}

func TestDueEqual(t *testing.T) {
	d := Due(parseTime("2017-01-02T00:00:00Z"))
	t.Run("equal", func(t *testing.T) {
		d2 := Due(parseTime("2017-01-02T00:00:00Z"))
		if !d.Equal(d2) {
			t.Errorf("Expected equality")
		}
	})
	t.Run("unequal", func(t *testing.T) {
		d2 := Due(parseTime("2017-01-01T00:00:00Z"))
		if d.Equal(d2) {
			t.Errorf("Expected inequality")
		}
	})
}

func TestDueAfter(t *testing.T) {
	d := Due(parseTime("2017-01-02T00:00:00Z"))
	t.Run("after", func(t *testing.T) {
		d2 := Due(parseTime("2017-01-01T00:00:00Z"))
		if !d.After(d2) {
			t.Error("Expected after")
		}
	})
	t.Run("before", func(t *testing.T) {
		d2 := Due(parseTime("2018-01-01T00:00:00Z"))
		if d.After(d2) {
			t.Error("Expected not after")
		}
	})
	t.Run("equal", func(t *testing.T) {
		d2 := Due(parseTime("2017-01-02T00:00:00Z"))
		if d.After(d2) {
			t.Error("Expected not after")
		}
	})
}

func TestMidnight(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "noon local tz",
			input:    parseTime("2017-06-06T12:00:00Z"),
			expected: parseTime("2017-06-06T00:00:00+00:00"),
		},
		{
			name:     "midnight utc",
			input:    parseTime("2017-06-06T00:00:00+00:00"),
			expected: parseTime("2017-06-06T00:00:00+00:00"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := midnight(test.input)
			if !result.Equal(test.expected) {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestParseInterval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Interval
		err      string
	}{
		{
			name:  "completely bogus",
			input: "completely bogus",
			err:   `strconv.ParseInt: parsing "completely bogu": invalid syntax`,
		},
		{
			name:  "invalid unit",
			input: "89q",
			err:   "Unknown unit in '89q'",
		},
		{
			name:     "seconds",
			input:    "15s",
			expected: Interval(15 * time.Second),
		},
		{
			name:     "minutes",
			input:    "15m",
			expected: Interval(15 * time.Minute),
		},
		{
			name:     "hours",
			input:    "15h",
			expected: Interval(15 * time.Hour),
		},
		{
			name:     "days",
			input:    "15d",
			expected: Interval(15 * 24 * time.Hour),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseInterval(test.input)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if !result.Equal(test.expected) {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestIntervalMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    Interval
		expected string
		err      string
	}{
		{
			name:     "seconds",
			input:    Interval(15 * time.Second),
			expected: `-15`,
		},
		{
			name:     "minutes",
			input:    Interval(15 * time.Minute),
			expected: "-900",
		},
		{
			name:     "hours",
			input:    Interval(15 * time.Hour),
			expected: "-54000",
		},
		{
			name:     "days",
			input:    Interval(15 * 24 * time.Hour),
			expected: "15",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.MarshalJSON()
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.JSON([]byte(test.expected), result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestIntervalUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Interval
		err      string
	}{
		{
			name:  "invalid json",
			input: "invalid json",
			err:   `strconv.Atoi: parsing "invalid json": invalid syntax`,
		},
		{
			name:     "seconds",
			input:    "-15",
			expected: Interval(15 * time.Second),
		},
		{
			name:     "minutes",
			input:    "-900",
			expected: Interval(15 * time.Minute),
		},
		{
			name:     "hours",
			input:    "-54000",
			expected: Interval(15 * time.Hour),
		},
		{
			name:     "days",
			input:    "15",
			expected: Interval(15 * 24 * time.Hour),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result Interval
			err := result.UnmarshalJSON([]byte(test.input))
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
