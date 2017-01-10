package test

import (
	"os"
	"sync"
	"testing"

	"github.com/flimzy/testify/require"

	"github.com/flimzy/go-pouchdb"
	"github.com/gopherjs/gopherjs/js"

	"github.com/FlashbackSRS/flashback/repository"
)

func init() {
	memdown := js.Global.Call("require", "memdown")
	repo.PouchDBOptions = pouchdb.Options{
		DB: memdown,
	}
}

var bundleID = "bundle-alnlcvykyjxsjtijzonc3456kd5u4757"

var expectedUserDBIDs = []string{
	"_design/idx-740f58f89a91c3283d5ef9b896e9ac9f",
	bundleID,
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.1",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.efn_5zJV184Q7hZzE8zmlclqllY.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.efn_5zJV184Q7hZzE8zmlclqllY.1",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.l2ZvABKQnCBZbuaYOlSIrZWqRQI.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.l2ZvABKQnCBZbuaYOlSIrZWqRQI.1",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.qT9Gr_a9D_jkaapw1xy7KYfvTOs.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.qT9Gr_a9D_jkaapw1xy7KYfvTOs.1",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.rRm8q5nIKgIMC__jMxYmhXRF_2I.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.rRm8q5nIKgIMC__jMxYmhXRF_2I.1",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.udROb8T8RmRASG5zGHNKnKL25zI.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.udROb8T8RmRASG5zGHNKnKL25zI.1",
}

var expectedBundleDBIDs = []string{
	"_design/idx-740f58f89a91c3283d5ef9b896e9ac9f",
	bundleID,
	"deck-MoAm80CALRtrMk7Y4eOmGtCzzjY",
	"deck-sXxr9js6DFSFotJ6yuISxOQCuKU",
	"note-ZR4TpeX38xRzRvXprlgJpP4Ribo",
	"note-efn_5zJV184Q7hZzE8zmlclqllY",
	"note-l2ZvABKQnCBZbuaYOlSIrZWqRQI",
	"note-qT9Gr_a9D_jkaapw1xy7KYfvTOs",
	"note-rRm8q5nIKgIMC__jMxYmhXRF_2I",
	"note-udROb8T8RmRASG5zGHNKnKL25zI",
	"theme-ELr8cEJJOvJU4lYz-VTXhH8wLTo",
}

var importComplete bool
var importMu sync.Mutex

func testImport(t *testing.T) {
	importMu.Lock()
	defer importMu.Unlock()
	if importComplete {
		return
	}
	require := require.New(t)
	fbb, err := os.Open(fbbFile)
	require.Nil(err, "Error reading %s: %s", fbbFile, err)

	user := &repo.User{testUser}
	err = repo.Import(user, fbb)
	require.Nil(err, "Error importing file: %+v", err)

	importComplete = true
}

func TestImport(t *testing.T) {
	require := require.New(t)

	user := &repo.User{testUser}
	udb, err := user.DB()
	require.Nil(err, "Error connecting to User DB: %s", err)

	var allUserDocs, allBundleDocs map[string]interface{}
	err = udb.AllDocs(&allUserDocs, pouchdb.Options{})
	require.Nil(err, "Error fetching AllDocs(): %s", err)

	// Remove the revs, because they're random
	urows := allUserDocs["rows"].([]interface{})
	uids := make([]string, len(urows))
	for i, row := range urows {
		uids[i] = row.(map[string]interface{})["id"].(string)
	}
	require.DeepEqual(expectedUserDBIDs, uids, "User DB IDs")

	bdb, err := user.NewDB(bundleID)
	require.Nil(err, "Error connecting to Bundle DB: %s", err)

	if err := bdb.AllDocs(&allBundleDocs, pouchdb.Options{}); err != nil {
		t.Fatalf("Error fetching AllDocs(): %s", err)
	}
	// Remove the revs, because they're random
	brows := allBundleDocs["rows"].([]interface{})
	bids := make([]string, len(brows))
	for i, row := range brows {
		bids[i] = row.(map[string]interface{})["id"].(string)
	}
	require.DeepEqual(expectedBundleDBIDs, bids, "Bundle DB IDs")
}
