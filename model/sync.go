package model

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	if e := r.doSync(ctx, rdb, udbName, &docsWritten, &docsRead); e != nil {
		return errors.Wrap(e, "sync failed")
	}

	updated, err := r.upgradeSchema(ctx)
	if err != nil {
		return errors.Wrap(err, "schema upgrade failed")
	}
	if updated {
		fmt.Printf("Documents were updated\n")
		if e := r.doSync(ctx, rdb, udbName, &docsWritten, &docsRead); e != nil {
			return errors.Wrap(e, "resync failed")
		}
	}

	log.Debugf("Synced %d docs from server, %d to server\n", docsRead, docsWritten)

	return nil
}

func (r *Repo) doSync(ctx context.Context, remoteUserDBName, localUserDBName string, docsWritten, docsRead *int32) error {
	// local to remote
	if e := replicate(ctx, r.local, remoteUserDBName, localUserDBName, docsWritten); e != nil {
		return errors.Wrap(e, "sync local to remote")
	}

	// remote to local
	if e := replicate(ctx, r.local, localUserDBName, remoteUserDBName, docsRead); e != nil {
		return errors.Wrap(e, "sync remote to local")
	}

	if e := r.syncBundles(ctx, docsRead, docsWritten); e != nil {
		return errors.Wrap(e, "bundle sync")
	}

	return errors.Wrap(r.updateSyncTime(ctx), "fialed to store sync timestamp")
}

// upgradeSchema updates the local schema, if necessary, and returns true if
// any updates were made.
//
// This should be run after a sync, and in case of updates, a sync should be
// re-run. This is to reduce the chance of a race condition with multiple
// clients doing a simultaneous update.
func (r *Repo) upgradeSchema(ctx context.Context) (bool, error) {
	defer profile("upgrade")()
	db, err := r.userDB(ctx)
	if err != nil {
		return false, err
	}

	var newUpdated, oldUpdated bool
	var newErr, oldErr error
	cache := newCardDeckCache(r.local)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		newUpdated, newErr = upgradeSchemaFromView(ctx, db, cache, "newCards")
		wg.Done()
	}()
	go func() {
		oldUpdated, oldErr = upgradeSchemaFromView(ctx, db, cache, "oldCards")
		wg.Done()
	}()
	wg.Wait()
	if newErr != nil {
		return newUpdated || oldUpdated, newErr
	}
	return newUpdated || oldUpdated, oldErr
}

func upgradeSchemaFromView(ctx context.Context, db kivikDB, cache *cardDeckCache, view string) (bool, error) {
	defer profile(fmt.Sprintf("upgrade from view %s", view))()
	rows, err := db.Query(ctx, "index", view, map[string]interface{}{
		"startkey":     []interface{}{nil},
		"endkey":       []interface{}{nil, map[string]interface{}{}},
		"include_docs": true,
	})
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	var count int
	for rows.Next() {
		var card *fb.Card
		if e := rows.ScanDoc(&card); e != nil {
			return count != 0, e
		}
		deckID, err := cache.cardDeck(ctx, card)
		if err != nil {
			return count != 0, err
		}
		card.Deck = deckID
		count++
		if _, err := db.Put(ctx, card.ID, card); err != nil {
			return count != 0, err
		}
	}
	log.Debugf("%d of %d %s cards upgraded\n", count, rows.TotalRows(), view)
	return count != 0, rows.Err()
}

type cardDeckCache struct {
	client      kivikClient
	cache       map[string]string
	readBundles map[string]struct{}
}

func newCardDeckCache(client kivikClient) *cardDeckCache {
	return &cardDeckCache{
		client:      client,
		cache:       make(map[string]string),
		readBundles: make(map[string]struct{}),
	}
}

func (c *cardDeckCache) cardDeck(ctx context.Context, card *fb.Card) (string, error) {
	bundleID := card.BundleID()
	if _, ok := c.readBundles[bundleID]; !ok {
		if err := c.readBundle(ctx, bundleID); err != nil {
			return "", err
		}
	}
	if deckID, ok := c.cache[card.ID]; ok {
		return deckID, nil
	}
	return "", nil
}

func (c *cardDeckCache) readBundle(ctx context.Context, bundleID string) error {
	bdb, err := c.client.DB(ctx, bundleID)
	if err != nil {
		return err
	}
	c.readBundles[bundleID] = struct{}{}
	rows, err := bdb.AllDocs(ctx, map[string]interface{}{
		"startkey":     "deck-",
		"endkey":       "deck-" + kivik.EndKeySuffix,
		"include_docs": true,
	})
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var deck fb.Deck
		if err := rows.ScanDoc(&deck); err != nil {
			return err
		}
		for _, cardID := range deck.Cards.All() {
			c.cache[cardID] = deck.ID
		}
	}
	return nil
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
