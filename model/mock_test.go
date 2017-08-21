package model

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

type mockAttachmentGetter struct {
	attachments map[string]*kivik.Attachment
	err         error
}

var _ attachmentGetter = &mockAttachmentGetter{}

func (db *mockAttachmentGetter) GetAttachment(_ context.Context, _, _, filename string) (*kivik.Attachment, error) {
	if db.err != nil {
		return nil, db.err
	}
	if att, ok := db.attachments[filename]; ok {
		return att, nil
	}
	return nil, errors.Status(kivik.StatusNotFound, "not found")
}
