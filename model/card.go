package model

import (
	"context"
	"errors"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/kivik"
)

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
		if card.BuriedUntil != nil && card.BuriedUntil.After(fb.Now()) {
			continue
		}
		// Skip cards we already saw today, with an interval >= 1d; they would make no progress.
		if card.Interval != nil && card.LastReview != nil && card.Interval.Days() >= 1 && !time.Time(fb.Today()).After(*card.LastReview) {
			continue
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
