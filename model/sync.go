package model

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/flimzy/kivik"
	"github.com/flimzy/log"
	multierror "github.com/hashicorp/go-multierror"
)

// Sync performs a bi-directional sync.
func (r *Repo) Sync(ctx context.Context) error {
	ldb, err := r.userDB(ctx)
	if err != nil {
		return err
	}
	rdb, err := r.remoteUserDB(ctx)
	if err != nil {
		return err
	}

	var docsWritten, docsRead int32
	errCh := make(chan error, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		// local to remote
		if err := replicate(ctx, r.local, rdb, ldb, &docsWritten); err != nil {
			errCh <- err
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// remote to local
		if err := replicate(ctx, r.local, ldb, rdb, &docsRead); err != nil {
			errCh <- err
		}
		wg.Done()
	}()

	wg.Wait()
	close(errCh)
	log.Debugf("Synced %d docs from server, %d to server\n", docsRead, docsWritten)
	var errs error
	for err := range errCh {
		errs = multierror.Append(errs, err)
	}
	return errs
}

func dbDSN(db *kivik.DB) string {
	return fmt.Sprintf("%s/%s", db.Client().DSN(), db.Name())
}

type clientReplicator interface {
	Replicate(context.Context, string, string, ...kivik.Options) (*kivik.Replication, error)
}

func replicate(ctx context.Context, client clientReplicator, target, source *kivik.DB, count *int32) error {
	replication, err := client.Replicate(ctx, "", "", map[string]interface{}{
		"target": target,
		"source": source,
	})
	if err != nil {
		return err
	}
	c, err := processReplication(ctx, replication)
	atomic.AddInt32(count, c)
	return err
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
