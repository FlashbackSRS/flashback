package model

import (
	"context"

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

func getCardsFromView(ctx context.Context, db querier, view string, limit int) ([]*fb.Card, error) {
	rows, err := db.Query(context.TODO(), "cards", view, map[string]interface{}{
		"limit":        int(float64(limit) * 1.5),
		"include_docs": true,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	cards := make([]*fb.Card, 0, limit)
	for rows.Next() {
		card := &fb.Card{}
		if err := rows.ScanDoc(card); err != nil {
			return nil, err
		}
		if card.BuriedUntil != nil && card.BuriedUntil.After(fb.Now()) {
			continue
		}
		cards = append(cards, card)
	}
	return cards, nil
}
