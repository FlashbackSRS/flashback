package model

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/flimzy/diff"
)

func TestDeckList(t *testing.T) {
	tests := []struct {
		name     string
		repo     *Repo
		expected []Deck
		err      string
	}{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}

func TestDeckStats(t *testing.T) {
	tests := []struct {
		name     string
		db       querier
		expected []Deck
		err      string
	}{
		{
			name: "query error",
			db:   &mockQuerier{err: errors.New("some error")},
			err:  "some error",
		},
		{
			name:     "no cards",
			db:       &mockQuerier{rows: []*mockRows{{}}},
			expected: []Deck{},
		},
		{
			name: "one deck",
			db: &mockQuerier{
				rows: []*mockRows{{
					rows:   []string{"", "", ""},
					values: []string{"[234,6]", "[1811,56]", "[52,9]"},
					keys: []string{
						`["new","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["old","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["suspended","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
					},
				}},
			},
			expected: []Deck{
				{
					ID:             "deck-Brm5eFOpF0553VTksh7hlySt6M8",
					TotalCards:     2097,
					DueCards:       0,
					LearningCards:  56,
					MatureCards:    1755,
					NewCards:       234,
					SuspendedCards: 52,
				},
			},
		},
		{
			name: "multiple decks",
			db: &mockQuerier{
				rows: []*mockRows{{
					rows: []string{"", "", "", "", "", "", ""},
					values: []string{
						"[234,6]", "[1811,56]", "[52,9]",
						"[100,0]", "[100,20]", "[5,1]",
						"[50,0]",
					},
					keys: []string{
						`["new","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["old","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["suspended","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["new","deck-foo"]`,
						`["old","deck-foo"]`,
						`["suspended","deck-foo"]`,
						`["new","deck-bar"]`,
					},
				}},
			},
			expected: []Deck{
				{
					ID:             "deck-Brm5eFOpF0553VTksh7hlySt6M8",
					TotalCards:     2097,
					DueCards:       0,
					LearningCards:  56,
					MatureCards:    1755,
					NewCards:       234,
					SuspendedCards: 52,
				},
				{
					ID:         "deck-bar",
					TotalCards: 50,
					NewCards:   50,
				},
				{
					ID:             "deck-foo",
					TotalCards:     205,
					DueCards:       0,
					LearningCards:  20,
					MatureCards:    80,
					NewCards:       100,
					SuspendedCards: 5,
				},
			},
		}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := deckStats(context.Background(), test.db)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if test.err != errMsg {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			sort.Slice(result, func(i, j int) bool {
				return result[i].ID < result[j].ID
			})
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
