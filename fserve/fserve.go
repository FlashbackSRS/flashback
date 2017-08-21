package fserve

import (
	"context"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/iframes"
	"github.com/FlashbackSRS/flashback/model"
)

// Register registers a new fserve attached to the specified repo.
func Register(repo *model.Repo) {
	log.Debug("Registering fserve listener\n")
	iframes.RegisterListener("fserve", fserve(repo))
	log.Debug("Done registering fserve listener\n")
}

func fserve(repo *model.Repo) iframes.ListenerFunc {
	return func(cardID string, payload *js.Object, respond iframes.Respond) error {
		path := payload.String()
		log.Debugf("fserve request: File '%s' for card '%s'\n", path, cardID)
		att, err := repo.FetchAttachment(context.TODO(), cardID, path)
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
}
