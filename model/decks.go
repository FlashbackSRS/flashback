package model

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"

	fb "github.com/FlashbackSRS/flashback-model"
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
	defer profile("deck list")()
	udb, err := r.userDB(ctx)
	if err != nil {
		return nil, err
	}
	if e := checkDDocVersion(ctx, udb); e != nil {
		return nil, e
	}

	decks, err := deckReducedStats(ctx, udb)
	if err != nil {
		return nil, err
	}

	if err := fleshenDecks(ctx, udb, decks); err != nil {
		return nil, err
	}
	sort.Slice(decks, func(i, j int) bool {
		return decks[i].Name < decks[j].Name
	})
	allDeck := &Deck{
		ID:   allDeckID,
		Name: allDeckName,
	}
	for _, deck := range decks {
		allDeck.TotalCards += deck.TotalCards
		allDeck.DueCards += deck.DueCards
		allDeck.LearningCards += deck.LearningCards
		allDeck.MatureCards += deck.MatureCards
		allDeck.NewCards += deck.NewCards
		allDeck.SuspendedCards += deck.SuspendedCards
	}
	return append([]*Deck{allDeck}, decks...), nil
}

func fleshenDecks(ctx context.Context, db kivikDB, decks []*Deck) error {
	sem := make(chan struct{}, 3) // Run at most 3 simultaneous fetches
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error)
	var err error
	go func() {
		for e := range errCh {
			if err == nil {
				cancel()
				err = e
			}
		}
	}()
	ts := now()
	for _, deck := range decks {
		sem <- struct{}{}
		go func(deck *Deck) {
			var e error
			deck.Name, e = deckName(ctx, db, deck.ID)
			if e != nil {
				errCh <- e
			}
			<-sem
		}(deck)
		if deck.MatureCards+deck.LearningCards == 0 {
			continue
		}
		sem <- struct{}{}
		go func(deck *Deck) {
			var e error
			deck.DueCards, e = dueCount(ctx, db, deck.ID, ts)
			if e != nil {
				errCh <- e
			}
			<-sem
		}(deck)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	return err
}

func dueCount(ctx context.Context, db querier, deckID string, ts time.Time) (int, error) {
	defer profile(fmt.Sprintf("due count for %s", deckID))()
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
		var doc struct {
			BuriedUntil fb.Due `json:"buriedUntil"`
		}
		if e := rows.ScanValue(&doc); e != nil {
			return count, e
		}
		if time.Time(doc.BuriedUntil).After(ts) {
			continue
		}
		count++
	}
	return count, rows.Err()
}

func deckReducedStats(ctx context.Context, db querier) ([]*Deck, error) {
	defer profile("deck reduced stats")()
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
		if e := rows.ScanKey(&key); e != nil {
			return nil, e
		}
		if key[1] == allDeckID {
			// Skip the aggregate 'all' deck; it's calculated later
			continue
		}
		if e := rows.ScanValue(&values); e != nil {
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

const (
	orphanedCardDeckID   = "x"
	orphanedCardDeckName = "[No Deck]"
	allDeckID            = ""
	allDeckName          = "All"
)

func deckName(ctx context.Context, db getter, deckID string) (string, error) {
	if deckID == orphanedCardDeckID {
		return orphanedCardDeckName, nil
	}
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

// copied from github.com/flimzy/flashback-server2
const (
	UserDDocID      = "_design/index"
	UserDDocVersion = 1
)

func checkDDocVersion(ctx context.Context, db kivikDB) error {
	row, err := db.Get(ctx, UserDDocID)
	if err != nil {
		return err
	}
	var ddoc struct {
		Version int `json:"version"`
	}
	if e := row.ScanDoc(&ddoc); e != nil {
		return e
	}
	if ddoc.Version != UserDDocVersion {
		return errors.Status(kivik.StatusNotFound, "current ddoc not found")
	}
	return nil
}
