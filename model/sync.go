package model

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/flimzy/kivik"
	"github.com/flimzy/log"
	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
)

func (r *Repo) remoteDSN(name string) string {
	dsn := r.remote.DSN()
	if strings.HasSuffix(dsn, "/") {
		return dsn + name
	}
	return dsn + "/" + name
}

// Sync performs a bi-directional sync.
func (r *Repo) Sync(ctx context.Context) error {
	u, err := r.CurrentUser()
	if err != nil {
		return err
	}
	udbName := "user-" + u
	rdb := r.remoteDSN(udbName)

	var docsWritten, docsRead int32

	// local to remote
	if err := replicate(ctx, r.local, rdb, udbName, &docsWritten); err != nil {
		return errors.Wrap(err, "sync local to remote")
	}

	// remote to local
	if err := replicate(ctx, r.local, udbName, rdb, &docsRead); err != nil {
		return errors.Wrap(err, "sync remote to local")
	}

	log.Debugf("Synced %d docs from server, %d to server\n", docsRead, docsWritten)
	return nil
}

type clientReplicator interface {
	Replicate(context.Context, string, string, ...kivik.Options) (*kivik.Replication, error)
}

func dbDSN(db *kivik.DB) string {
	dsn := db.Client().DSN()
	dbName := db.Name()
	if dsn != "" && !strings.HasSuffix(dsn, "/") {
		return dsn + "/" + dbName
	}
	return dsn + dbName
}

func replicate(ctx context.Context, client clientReplicator, target, source string, count *int32) error {
	replication, err := client.Replicate(ctx, target, source)
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

func (r *Repo) syncBundles(ctx context.Context, reads, writes *int32) error {
	udb, err := r.userDB(ctx)
	if err != nil {
		return err
	}
	log.Debugf("Reading bundles from user database...\n")
	rows, err := udb.Find(context.TODO(), map[string]interface{}{
		"selector": map[string]string{"type": "bundle"},
		"fields":   []string{"_id"},
	})
	if err != nil {
		return errors.Wrap(err, "failed to sync bundles")
	}

	var bundles []*fb.Bundle
	for rows.Next() {
		var bundle *fb.Bundle
		if err := rows.ScanDoc(&bundle); err != nil {
			return errors.Wrapf(err, "failed to scan bundle %s", rows.ID())
		}
		bundles = append(bundles, bundle)
	}
	log.Debugf("bundles = %v\n", bundles)
	for _, bundle := range bundles {
		log.Debugf("Bundle %s", bundle)
		ldb := bundle.ID.String()
		rdb := r.remoteDSN(ldb)
		if err := r.remote.CreateDB(ctx, rdb); err != nil {
			return err
		}
		if err := replicate(ctx, r.local, rdb, ldb, writes); err != nil {
			return errors.Wrap(err, "bundle push")
		}
		if err := replicate(ctx, r.local, ldb, rdb, reads); err != nil {
			return errors.Wrap(err, "bundle pull")
		}
	}
	return nil
}

/*
func SyncReviews(local, remote *repo.DB) (int32, error) {
	u, err := repo.CurrentUser()
	if err != nil {
		return 0, err
	}
	host := util.CouchHost()
	ldb, err := util.ReviewsSyncDbs()
	if err != nil {
		return 0, err
	}
	if ldb == nil {
		return 0, nil
	}
	before, err := ldb.Info()
	if err != nil {
		return 0, err
	}
	if before.DocCount == 0 {
		// Nothing at all to sync
		return 0, nil
	}
	rdb, err := repo.NewDB(host + "/" + u.MasterReviewsDBName())
	if err != nil {
		return 0, err
	}
	revsSynced, err := Sync(ldb, rdb)
	if err != nil {
		return 0, err
	}
	after, err := ldb.Info()
	if err != nil {
		return revsSynced, err
	}
	if before.DocCount != after.DocCount || before.UpdateSeq != after.UpdateSeq {
		log.Debugf("ReviewsDb content changed during sync. Refusing to delete.\n")
		return revsSynced, nil
	}
	log.Debugf("Ready to zap %s\n", after.DBName)
	err = util.ZapReviewsDb(ldb)
	return revsSynced, err
}
*/
