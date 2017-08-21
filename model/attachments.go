package model

import (
	"context"

	fb "github.com/FlashbackSRS/flashback-model"
)

// FetchAttachment fetches the requested attachment associated with the specified
// card.
func (r *Repo) FetchAttachment(ctx context.Context, cardID, filename string) (*fb.Attachment, error) {
	udb, err := r.userDB(ctx)
	if err != nil {
		return nil, err
	}
	card := &fb.Card{}
	if err := getDoc(ctx, udb, cardID, &card); err != nil {
		return nil, err
	}
	fbCard := &fbCard{Card: card}
	if err := fbCard.fetch(ctx, r.local); err != nil {
		return nil, err
	}
	return nil, nil
	// u, err := repo.CurrentUser()
	// if err != nil {
	// 	return nil, errors.Wrap(err, "current user")
	// }
	// card, err := u.GetCard(cardID)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "fetch card")
	// }
	// return card.GetAttachment(filename)
}
