package model

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/log"
	"github.com/pkg/errors"
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

	if err := r.syncBundles(ctx, &docsRead, &docsWritten); err != nil {
		return errors.Wrap(err, "bundle sync")
	}

	log.Debugf("Synced %d docs from server, %d to server\n", docsRead, docsWritten)
	return r.updateSyncTime(ctx)
}

const lastSyncTimestampDocID = "_local/lastSyncTimestamp"

type lastSyncTimestampDoc struct {
	ID       string    `json:"_id"`
	Rev      string    `json:"_rev"`
	LastSync time.Time `json:"lastSync"`
}

// updateSyncTime updates the local timestamp for the last sync.
func (r *Repo) updateSyncTime(ctx context.Context) error {
	rev, _, err := r.lastSyncTime(ctx)
	if err != nil && kivik.StatusCode(err) != kivik.StatusNotFound {
		return err
	}

	u, err := r.CurrentUser()
	if err != nil {
		return err
	}
	db, err := r.local.DB(ctx, "user-"+u)
	if err != nil {
		return err
	}
	doc := lastSyncTimestampDoc{
		ID:       lastSyncTimestampDocID,
		Rev:      rev,
		LastSync: now(),
	}
	_, err = db.Put(ctx, lastSyncTimestampDocID, doc)
	return err
}

// lastSyncTime returns the last time the database was synced.
func (r *Repo) lastSyncTime(ctx context.Context) (rev string, lastSync time.Time, err error) {
	u, err := r.CurrentUser()
	if err != nil {
		return "", time.Time{}, err
	}
	db, err := r.local.DB(ctx, "user-"+u)
	if err != nil {
		return "", time.Time{}, err
	}
	row, err := db.Get(ctx, lastSyncTimestampDocID)
	if err != nil {
		return "", time.Time{}, err
	}
	var doc lastSyncTimestampDoc
	if e := row.ScanDoc(&doc); e != nil {
		return "", time.Time{}, e
	}
	return doc.Rev, doc.LastSync, nil
}

type clientReplicator interface {
	Replicate(context.Context, string, string, ...kivik.Options) (*kivik.Replication, error)
}

func dbDSN(db clientNamer) string {
	dsn := db.Client().DSN()
	dbName := db.Name()
	if dsn != "" && !strings.HasSuffix(dsn, "/") {
		return dsn + "/" + dbName
	}
	return dsn + dbName
}

func replicate(ctx context.Context, client clientReplicator, target, source string, count *int32) error {
	defer profile(fmt.Sprintf("replicate %s -> %s", source, target))()
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
	defer profile("syncBundles")()
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

	var bundles []string
	for rows.Next() {
		var result struct {
			ID string `json:"_id"`
		}
		if err := rows.ScanDoc(&result); err != nil {
			return errors.Wrapf(err, "failed to scan bundle %s", rows.ID())
		}
		bundles = append(bundles, result.ID)
	}
	log.Debugf("bundles = %v\n", bundles)
	for _, bundle := range bundles {
		log.Debugf("Creating remote bundle: %s\n", bundle)
		rdb := r.remoteDSN(bundle)
		if err := r.remote.CreateDB(ctx, bundle); err != nil && kivik.StatusCode(err) != kivik.StatusPreconditionFailed {
			return errors.Wrap(err, "create remote bundle")
		}
		if err := replicate(ctx, r.local, rdb, bundle, writes); err != nil {
			return errors.Wrap(err, "bundle push")
		}
		if err := replicate(ctx, r.local, bundle, rdb, reads); err != nil {
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
