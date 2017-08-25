package model

import (
	"context"
	"io"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/pkg/errors"
)

// BuryRelatedCards buries related cards.
//
// The burial strategy involves the following:
//
// 1. New cards (reviewCount = 0) are buried for NewBuryTime. This should help
//    start of new card review to get off on the right foot.
// 2. All related cards are buried for a minimum of 1 day (this is all Anki does)
// 3. The target burial time is the current card's interval, divided by the number
//    of related cards. NewBuryTime is used as the minimum target burial time.
// 4. The maximum burial is MaxBuryRatio of the card's interval.
func (r *Repo) BuryRelatedCards(ctx context.Context, card *fb.Card) error {
	defer profile("BuryRelatedCards")()
	db, err := r.userDB(ctx)
	if err != nil {
		return err
	}
	cards, err := fetchRelatedCards(ctx, db, card.ID)
	if err != nil {
		return err
	}
	toBury := setBurials(card.Interval, cards)
	if len(toBury) == 0 {
		return nil
	}
	return updateDocs(ctx, db, toBury)
}

func setBurials(interval fb.Interval, cards []*fb.Card) []*fb.Card {
	if len(cards) == 0 {
		return cards
	}
	buryTarget := interval / fb.Interval(len(cards))
	burials := make([]*fb.Card, 0, len(cards))
	for _, card := range cards {
		newInterval := buryInterval(buryTarget, card.Interval, card.ReviewCount == 0)
		buryUntil := fb.Due(now().UTC()).Add(newInterval)
		// buryUntil := fb.DueIn(newInterval)
		// Now update the card, but only if we're trying to bury it longer
		// than it already is, to avoid unnecessary updates.
		if buryUntil.After(card.BuriedUntil) {
			card.BuriedUntil = buryUntil
			burials = append(burials, card)
		}

	}
	return burials
}

// fetchRelatedCards fetches cards related to the provided card ID.
func fetchRelatedCards(ctx context.Context, db allDocer, cardID string) ([]*fb.Card, error) {
	startKey, endKey := relatedKeyRange(cardID)
	rows, err := db.AllDocs(context.TODO(), map[string]interface{}{
		"include_docs": true,
		"start_key":    startKey,
		"end_key":      endKey,
	})
	if err != nil {
		return nil, err
	}
	cards := make([]*fb.Card, 0)
	for rows.Next() {
		if cardID == rows.ID() {
			// Skip the reference card
			continue
		}
		var card fb.Card
		if err := rows.ScanDoc(&card); err != nil {
			return nil, errors.Wrap(err, "scan doc")
		}
		cards = append(cards, &card)
	}
	if err := rows.Err(); err != nil && err != io.EOF {
		return nil, err
	}
	return cards, nil
}

// NewBuryTime sets the time to bury related new cards.
const NewBuryTime = 7 * fb.Day

// MinBuryTime is the minimal burial time.
const MinBuryTime = 1 * fb.Day

// MaxBuryRatio is the maximum burial-time / interval ratio.
const MaxBuryRatio = 0.20

func buryInterval(bury fb.Interval, ivl fb.Interval, new bool) fb.Interval {
	if new {
		return NewBuryTime
	}
	var maxBury fb.Interval
	if ivl > 0 {
		maxBury = fb.Interval(float64(ivl) * MaxBuryRatio)
	}
	if bury > maxBury {
		bury = maxBury
	}
	if bury < MinBuryTime {
		bury = MinBuryTime
	}
	return bury
}
