package model

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
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

func TestFetchRelatedCards(t *testing.T) {
	tests := []struct {
		name     string
		db       allDocer
		cardID   string
		expected []*fb.Card
		err      string
	}{
		{
			name:   "db error",
			db:     &mockAllDocer{err: errors.New("db error")},
			cardID: "card-foo.bar.0",
			err:    "db error",
		},
		{
			name:   "iteration error",
			db:     &mockAllDocer{rows: &mockRows{err: errors.New("db error")}},
			cardID: "card-foo.bar.0",
			err:    "db error",
		},
		{
			name: "invalid json",
			db: &mockAllDocer{
				rows: &mockRows{
					rows: []string{
						`{"_id":"card-foo.bar.1", "created":"2017-01-01T01:01:01Z", "modified":12345, "model": "theme-Zm9v/0"}`,
					},
				},
			},
			cardID: "card-foo.bar.0",
			err:    `scan doc: parsing time "12345" as ""2006-01-02T15:04:05Z07:00"": cannot parse "12345" as """`,
		},
		{
			name: "success",
			db: &mockAllDocer{
				rows: &mockRows{
					rows: []string{
						`{"_id":"card-foo.bar.0", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "model": "theme-Zm9v/0"}`,
						`{"_id":"card-foo.bar.1", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "model": "theme-Zm9v/0"}`,
					},
				},
			},
			cardID: "card-foo.bar.0",
			expected: []*fb.Card{
				{
					ID:       "card-foo.bar.1",
					ModelID:  "theme-Zm9v/0",
					Created:  parseTime(t, "2017-01-01T01:01:01Z"),
					Modified: parseTime(t, "2017-01-01T01:01:01Z"),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := fetchRelatedCards(context.Background(), test.db, test.cardID)
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

type buryClient struct {
	kivikClient
	db kivikDB
}

var _ kivikClient = &buryClient{}

func (c *buryClient) DB(_ context.Context, _ string, _ ...kivik.Options) (kivikDB, error) {
	return c.db, nil
}

func TestBuryRelated(t *testing.T) {
	tests := []struct {
		name string
		repo *Repo
		card *fbCard
		err  string
	}{
		{
			name: "not logged in",
			repo: &Repo{},
			card: &fbCard{Card: &fb.Card{ID: "card-foo.bar.0"}},
			err:  "not logged in",
		},
		{
			name: "fetch error",
			repo: &Repo{user: "bob",
				local: &buryClient{
					db: &mockAllDocer{
						err: errors.New("db error"),
					},
				},
			},
			card: &fbCard{Card: &fb.Card{ID: "card-foo.bar.0"}},
			err:  "db error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.repo.BuryRelatedCards(context.Background(), test.card)
			checkErr(t, test.err, err)
		})
	}
}

func TestSetBurials(t *testing.T) {
	tests := []struct {
		name     string
		interval fb.Interval
		cards    []*fb.Card
		expected []*fb.Card
	}{
		{
			name:     "no cards",
			cards:    []*fb.Card{},
			expected: []*fb.Card{},
		},
		{
			name:     "two cards",
			interval: fb.Interval(24 * time.Hour),
			cards: []*fb.Card{
				{}, // new
				{
					ReviewCount: 1,
					Interval:    fb.Interval(24 * time.Hour),
				}, // Minimal burial
				{
					ReviewCount: 1,
					BuriedUntil: fb.Due(parseTime(t, "2018-01-01T00:00:00Z")),
				}, // Should not be re-buried
			},
			expected: []*fb.Card{
				{BuriedUntil: fb.Due(parseTime(t, "2017-01-08T00:00:00Z"))},
				{
					ReviewCount: 1,
					Interval:    fb.Interval(24 * time.Hour),
					BuriedUntil: fb.Due(parseTime(t, "2017-01-02T00:00:00Z")),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := setBurials(test.interval, test.cards)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
			for _, x := range result {
				fmt.Printf("%v\n", x.BuriedUntil)
			}
		})
	}
}
