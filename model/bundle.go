package model

import (
	"context"

	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
)

// SaveBundle saves the bundle.
func (r *Repo) SaveBundle(ctx context.Context, bundle *fb.Bundle) error {
	if bundle == nil || !bundle.ID.Valid() {
		return errors.New("invalid bundle")
	}
	if _, err := r.CurrentUser(); err != nil {
		return err
	}
	udb, err := r.userDB(ctx)
	if err != nil {
		return errors.Wrap(err, "userDB")
	}
	bdb, err := r.bundleDB(ctx, bundle)
	if err != nil {
		return errors.Wrap(err, "bundleDB")
	}
	if err := saveDoc(ctx, bdb, bundle); err != nil {
		return errors.Wrap(err, "bundle db write")
	}
	bundle.Rev = nil
	if err := saveDoc(ctx, udb, bundle); err != nil {
		return errors.Wrap(err, "user db write")
	}
	return nil
}
