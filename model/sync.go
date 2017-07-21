package model

import (
	"context"
	"fmt"

	"github.com/flimzy/kivik"
)

// Sync performs a bi-directional sync.
func (r *Repo) Sync(ctx context.Context) error {
	_, err := r.userDB(ctx)
	if err != nil {
		return err
	}
	return nil
}

func dbDSN(db *kivik.DB) string {
	return fmt.Sprintf("%s/%s", db.Client().DSN(), db.Name())
}

type clientReplicator interface {
	Replicate(context.Context, string, string, map[string]interface{}) (*kivik.Replication, error)
}

func replicate(ctx context.Context, client clientReplicator, target, source *kivik.DB) (int32, error) {
	replication, err := client.Replicate(ctx, "", "", map[string]interface{}{
		"target": target,
		"source": source,
	})
	if err != nil {
		return 0, err
	}
	return processReplication(ctx, replication)
}

type replication interface {
	IsActive() bool
	Update(context.Context) error
	Delete(context.Context) error
	Err() error
	DocsWritten() int64
}

func processReplication(ctx context.Context, rep replication) (int32, error) {
	// Just wait until the replication is complete
	// TODO: Visual updates
	for rep.IsActive() {
		if err := rep.Update(ctx); err != nil {
			_ = rep.Delete(ctx)
			return int32(rep.DocsWritten()), err
		}
	}
	return int32(rep.DocsWritten()), rep.Err()
}
