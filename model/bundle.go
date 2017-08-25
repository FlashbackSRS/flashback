package model

import (
	"context"

	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
)

// SaveBundle saves the bundle.
func (r *Repo) SaveBundle(ctx context.Context, bundle *fb.Bundle) error {
	if _, err := r.CurrentUser(); err != nil {
		return err
	}
	if bundle == nil {
		return errors.New("nil bundle")
	}
	if err := bundle.Validate(); err != nil {
		return errors.Wrap(err, "invalid bundle")
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
	bundle.Rev = ""
	if err := saveDoc(ctx, udb, bundle); err != nil {
		return errors.Wrap(err, "user db write")
	}
	return nil
}
