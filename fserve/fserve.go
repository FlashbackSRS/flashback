package fserve

import (
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/iframes"
	"github.com/FlashbackSRS/flashback/repository"
)

func init() {
	iframes.RegisterListener("fserve", fserve)
}

func fserve(cardID string, payload *js.Object, respond iframes.Respond) error {
	path := payload.String()
	log.Debugf("fserve request: File '%s' for card '%s'\n", path, cardID)
	att, err := fetchAttachment(cardID, path)
	if err != nil {
		return errors.Wrap(err, "fetch file")
	}
	data := js.NewArrayBuffer(att.Content)
	return respond("fserve", map[string]interface{}{
		"path":         path,
		"content_type": att.ContentType,
		"data":         data,
	}, data)
}

func fetchAttachment(cardID, filename string) (*repo.Attachment, error) {
	u, err := repo.CurrentUser()
	if err != nil {
		return nil, errors.Wrap(err, "current user")
	}
	card, err := u.GetCard(cardID)
	if err != nil {
		return nil, errors.Wrap(err, "fetch card")
	}
	return card.GetAttachment(filename)
}
