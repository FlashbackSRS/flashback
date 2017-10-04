package model

import (
	"context"
	"sort"
	"sync"
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
func (r *Repo) DeckList(ctx context.Context) ([]*Deck, error) {
	udb, err := r.userDB(ctx)
	if err != nil {
		return nil, err
	}
	decks, err := deckReducedStats(ctx, udb)
	if err != nil {
		return nil, err
	}

	ts := now()
	errs := make(chan error, 5)
	var wg sync.WaitGroup
	for _, deck := range decks {
		if deck.MatureCards+deck.LearningCards == 0 {
			continue
		}
		wg.Add(1)
		go func(deck *Deck) {
			dueCount, e := dueCount(ctx, udb, deck.ID, ts)
			deck.DueCards = dueCount
			errs <- e
			wg.Done()
		}(deck)
	}
	wg.Wait()
	close(errs)
	err = nil
	for e := range errs {
		if err == nil {
			err = e
		}
	}
	if err != nil {
		return nil, err
	}
	allDeck := &Deck{
		Name: "All",
	}
	for _, deck := range decks {
		deck.Name = deck.ID
		allDeck.TotalCards += deck.TotalCards
		allDeck.DueCards += deck.DueCards
		allDeck.LearningCards += deck.LearningCards
		allDeck.MatureCards += deck.MatureCards
		allDeck.SuspendedCards += deck.SuspendedCards
		allDeck.NewCards += deck.NewCards
	}
	sort.Slice(decks, func(i, j int) bool {
		return decks[i].Name < decks[j].Name
	})
	return append([]*Deck{allDeck}, decks...), nil
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

func deckReducedStats(ctx context.Context, db querier) ([]*Deck, error) {
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
			deck.SuspendedCards = values[0]
		case "new":
			deck.NewCards = values[0]
		case "old":
			deck.LearningCards = values[1]
			deck.MatureCards = values[0] - values[1]
		}
	}
	if e := rows.Err(); e != nil {
		return nil, e
	}
	decks := make([]*Deck, 0, len(deckMap))
	for _, deck := range deckMap {
		decks = append(decks, deck)
	}
	return decks, nil
}

func deckName(ctx context.Context, db getter, deckID string) (string, error) {
	row, err := db.Get(ctx, deckID)
	if err != nil {
		return "", err
	}
	var doc struct {
		Name string `json:"name"`
	}
	e := row.ScanDoc(&doc)
	return doc.Name, e
}
