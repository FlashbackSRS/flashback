package model

import (
	"context"
	"time"

	"github.com/flimzy/kivik"
)

// Deck represents a single deck.
type Deck struct {
	Name           string
	ID             string
	TotalCards     int
	DueCards       int
	LearningCards  int
	MatureCards    int
	NewCards       int
	SuspendedCards int
}

// DeckList returns a complete list of decks available for study.
func (r *Repo) DeckList(ctx context.Context) ([]Deck, error) {
	udb, err := r.userDB(ctx)
	if err != nil {
		return nil, err
	}
	decks, err := deckReducedStats(ctx, udb)
	if err != nil {
		return nil, err
	}
	return decks, nil
}

func dueCount(ctx context.Context, db querier, deckID string, ts time.Time) (int, error) {
	rows, err := db.Query(ctx, mainDDoc, mainView, kivik.Options{
		"startkey":     []interface{}{"old", deckID},
		"endkey":       []interface{}{"old", deckID, ts.Format(time.RFC3339)},
		"reduce":       false,
		"include_docs": false,
	})
	if err != nil {
		return 0, err
	}
	var count int
	for rows.Next() {
		count++
	}
	return count, rows.Err()
}

func deckReducedStats(ctx context.Context, db querier) ([]Deck, error) {
	rows, err := db.Query(ctx, mainDDoc, mainView, kivik.Options{
		"group_level": 2,
	})
	if err != nil {
		return nil, err
	}
	var key []string
	var values []int
	deckMap := make(map[string]*Deck)
	for rows.Next() {
		if e := rows.ScanValue(&values); e != nil {
			return nil, e
		}
		if e := rows.ScanKey(&key); e != nil {
			return nil, e
		}
		deck, ok := deckMap[key[1]]
		if !ok {
			deck = &Deck{ID: key[1]}
			deckMap[key[1]] = deck
		}

		deck.TotalCards += values[0]
		switch key[0] {
		case "suspended":
			deck.SuspendedCards += values[0]
		case "new":
			deck.NewCards += values[0]
		case "old":
			deck.LearningCards += values[1]
			deck.MatureCards = values[0] - values[1]
		}
	}
	if e := rows.Err(); e != nil {
		return nil, e
	}
	decks := make([]Deck, 0, len(deckMap))
	for _, deck := range deckMap {
		decks = append(decks, *deck)
	}
	return decks, nil
}
