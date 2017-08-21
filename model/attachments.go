package model

import (
	"context"

	fb "github.com/FlashbackSRS/flashback-model"
)

// FetchAttachment fetches the requested attachment associated with the specified
// card.
func (r *Repo) FetchAttachment(ctx context.Context, cardID, filename string) (*fb.Attachment, error) {
	defer profile("FetchAttachment %s %s", cardID, filename)()
	udb, err := r.userDB(ctx)
	if err != nil {
		return nil, err
	}
	card := &fb.Card{}
	if e := getDoc(ctx, udb, cardID, &card); e != nil {
		return nil, e
	}
	bdb, err := r.newDB(ctx, card.BundleID())
	if err != nil {
		return nil, err
	}
	fbCard := &fbCard{Card: card}
	if err := fbCard.fetch(ctx, r.local); err != nil {
		return nil, err
	}
	for _, attName := range fbCard.note.Note.Attachments.FileList() {
		if filename == attName {
			return getAttachment(ctx, bdb, fbCard.note.Note.ID, filename)
		}
	}
	for _, attName := range fbCard.model.Theme.Attachments.FileList() {
		if filename == attName {
			return getAttachment(ctx, bdb, fbCard.model.Theme.ID, filename)
		}
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
