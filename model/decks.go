package model

import (
	"context"

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
func (r *Repo) DeckList() ([]Deck, error) {
	return []Deck{
		{
			Name:           "All",
			ID:             "",
			TotalCards:     1142,
			DueCards:       120,
			LearningCards:  150,
			MatureCards:    400,
			NewCards:       577,
			SuspendedCards: 15,
		},
		{
			Name:           "Foo",
			ID:             "deck-asdf",
			TotalCards:     600,
			DueCards:       20,
			LearningCards:  50,
			MatureCards:    50,
			NewCards:       497,
			SuspendedCards: 3,
		},
		{
			Name:           "Bar",
			ID:             "deck-qwerty",
			TotalCards:     542,
			DueCards:       100,
			LearningCards:  100,
			MatureCards:    350,
			NewCards:       80,
			SuspendedCards: 12,
		},
	}, nil
}

func deckStats(ctx context.Context, db querier) ([]Deck, error) {
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
