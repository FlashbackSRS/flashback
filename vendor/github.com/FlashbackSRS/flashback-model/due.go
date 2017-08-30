package fb

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// Duration unit available for scheduling
const (
	Second = Interval(time.Second)
	Minute = Interval(time.Minute)
	Hour   = Interval(time.Hour)
	Day    = Interval(time.Hour * 24)
)

// This allows overriding time.Now() for tests
var now = time.Now

// Due formatting constants
const (
	DueDays    = "2006-01-02"
	DueSeconds = "2006-01-02 15:04:05"
)

// Due represents the time/date a card is due.
type Due time.Time

// Today returns today's date as a Due value
func Today() Due {
	return On(now())
}

// On returns the passed day's date as a Due value
func On(t time.Time) Due {
	return Due(t.Truncate(time.Duration(Day)))
}

// Now returns the current time as a Due value. Intended for use in comparisons.
func Now() Due {
	return Due(now())
}

// ParseDue attempts to parse the provided string as a due time.
func ParseDue(src string) (Due, error) {
	if t, err := time.Parse(DueDays, src); err == nil {
		return Due(t), nil
	}
	if t, err := time.Parse(DueSeconds, src); err == nil {
		return Due(t), nil
	}
	return Due{}, fmt.Errorf("Unrecognized input: %s", src)
}

// DueIn returns a new Due time i interval into the future. Durations greater
// than 24 hours into the future are rounded to the day.
func DueIn(i Interval) Due {
	return Due(now().UTC()).Add(i)
}

// IsZero returns true if the value is zero
func (d Due) IsZero() bool {
	return time.Time(d).IsZero()
}

// Add returns a new Due time with the duration added to it.
func (d Due) Add(ivl Interval) Due {
	dur := time.Duration(ivl)
	if ivl < Day {
		return Due(time.Time(d).Add(dur))
	}
	// Round up to the next whole day
	return Due(time.Time(d).Truncate(time.Duration(Day)).Add(time.Duration(ivl + Day - 1)).Truncate(time.Duration(Day)))
}

// Sub returns the interval between d and s
func (d Due) Sub(s Due) Interval {
	return Interval(time.Time(d).Sub(time.Time(s)))
}

// Equal returns true if the two due dates are equal.
func (d Due) Equal(d2 Due) bool {
	t1 := time.Time(d)
	t2 := time.Time(d2)
	return t1.Equal(t2)
}

// After returns true if d2 is after d.
func (d Due) After(d2 Due) bool {
	t1 := time.Time(d)
	t2 := time.Time(d2)
	return t1.After(t2)
}

func (d Due) String() string {
	t := time.Time(d)
	if t.Truncate(time.Duration(Day)).Equal(t) {
		return t.Format(DueDays)
	}
	return t.Format(DueSeconds)
}

// Time converts the due date to a standard time.Time
func (d Due) Time() time.Time {
	return time.Time(d)
}

func midnight(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

// MarshalJSON implements the json.Marshaler interface
func (d Due) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", d)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *Due) UnmarshalJSON(src []byte) error {
	due, err := ParseDue(string(bytes.Trim(src, "\"")))
	*d = due
	return err
}

// Interval represents the number of days or seconds between reviews
type Interval time.Duration

var unitMap = map[string]Interval{
	"s": Second,
	"m": Minute,
	"h": Hour,
	"d": Day,
}

// ParseInterval parses an interval string. An interval string is string
// containing a positive integer, follwed by a unit suffix. e.g. "300s" or "15d".
// No spaces or other characters are allowed. Valid units are "s", "m", "h", and
// "d".
func ParseInterval(s string) (Interval, error) {
	q, err := strconv.ParseInt(s[:len(s)-1], 10, 64)
	if err != nil {
		return 0, err
	}
	unit, ok := unitMap[s[len(s)-1:]]
	if !ok {
		return 0, fmt.Errorf("Unknown unit in '%s'", s)
	}
	return Interval(q) * unit, nil
}

func (i Interval) String() string {
	if days := i.Days(); days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	s := int(time.Duration(i).Seconds())
	if s%3600 == 0 {
		return fmt.Sprintf("%dh", s/3600)
	}
	if s%60 == 0 {
		return fmt.Sprintf("%dm", s/60)
	}
	return fmt.Sprintf("%ds", s)
}

// Equal returns true if the two intervals are effectively equal.
func (i Interval) Equal(i2 Interval) bool {
	return i.String() == i2.String()
}

// MarshalJSON implements the json.Marshaler interface
//
// Values > 1 day are stored as a positive integer. Sub-day values are stored
// as negative seconds. This is to make sorting easy in PouchDB.
func (i Interval) MarshalJSON() ([]byte, error) {
	var str string
	if days := i.Days(); days > 0 {
		str = strconv.Itoa(days)
	} else {
		str = strconv.Itoa(-int(time.Duration(i).Seconds()))
	}
	return []byte(str), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (i *Interval) UnmarshalJSON(src []byte) error {
	num, err := strconv.Atoi(string(src))
	if err != nil {
		return err
	}
	var ivl Interval
	if num < 0 {
		ivl = -Interval(num) * Second
	} else {
		ivl = Interval(num) * Day
	}
	*i = ivl
	return nil
}

// Days returns the number of whole days in the interval. Intervals
// less than one day return 0. Intervals greater than one day always
// round up to the next whole day.
func (i Interval) Days() int {
	if i >= Day {
		return int(time.Duration(i+24*Hour-1).Hours() / 24)
	}
	return 0
}
