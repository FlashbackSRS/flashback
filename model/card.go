package model

import (
	"context"
	"errors"
	"math"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/kivik"
)

// The priority for new cards.
const newPriority = 0.5

var now = time.Now

type querierWrapper struct {
	*kivik.DB
}

var _ querier = &querierWrapper{}

func (db *querierWrapper) Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error) {
	return db.DB.Query(ctx, ddoc, view, options...)
}

func newQuerier(db *kivik.DB) querier {
	return &querierWrapper{db}
}

type querier interface {
	Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error)
}

type kivikRows interface {
	Close() error
	Next() bool
	ScanDoc(dest interface{}) error
	TotalRows() int64
}

// limitPadding is a number added to the limit parameter passed to the
// getCardsFromView function. This is added, because there's no automated way
// to eliminate buried cards from the view, so they must be filtered in the
// client, but this could lead to queries with no results, so we pad the number
// of results to help reduce this chance.
const limitPadding = 100

func getCardsFromView(ctx context.Context, db querier, view string, limit, offset int) ([]*fb.Card, error) {
	if limit <= 0 {
		return nil, errors.New("invalid limit")
	}
	rows, err := db.Query(context.TODO(), "cards", view, map[string]interface{}{
		"limit":        limit + limitPadding,
		"offset":       offset,
		"include_docs": true,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	cards := make([]*fb.Card, 0, limit)
	var count int
	for rows.Next() {
		count++
		card := &fb.Card{}
		if err := rows.ScanDoc(card); err != nil {
			return nil, err
		}
		if card.BuriedUntil != nil && card.BuriedUntil.After(fb.Due(now())) {
			continue
		}
		if card.Interval != nil {
			// Skip cards we already saw today, with an interval >= 1d; they would make no progress.
			if card.LastReview != nil && card.Interval.Days() >= 1 && !time.Time(fb.On(now())).After(*card.LastReview) {
				continue
			}
			// Skip sub-day intervals that aren't due yet. We only allow forward-fuzzing for intervals > 1day
			if card.Due != nil && card.Interval.Days() == 0 && card.Due.After(fb.Due(now())) {
				continue
			}
		}
		cards = append(cards, card)
		if len(cards) == limit {
			return cards, nil
		}
	}
	if rows.TotalRows() > int64(limit+offset) {
		more, err := getCardsFromView(ctx, db, view, limit-len(cards), offset+count)
		return append(cards, more...), err
	}
	return cards, nil
}

// cardPriority returns a number 0 or greater, as a priority to be used in
// determining card study order.
func cardPriority(due fb.Due, interval fb.Interval, now time.Time) float64 {
	if due.IsZero() || interval == 0 {
		return newPriority
	}
	// Remove the timezone
	_, offset := now.Zone()
	utc := now.UTC().Add(time.Duration(offset) * time.Second)

	return float64(math.Pow(1+float64(utc.Sub(time.Time(due)))/float64(time.Duration(interval)), 3))
}
