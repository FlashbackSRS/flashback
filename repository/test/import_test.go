package test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/flimzy/testify/require"
	"github.com/gopherjs/gopherjs/js"

	"github.com/FlashbackSRS/flashback/repository"
)

func init() {
	memdown := js.Global.Call("require", "memdown")
	repo.PouchDBOptions = map[string]interface{}{
		"db": memdown,
	}
}

var bundleID = "bundle-alnlcvykyjxsjtijzonc3456kd5u4757"

var expectedUserDBIDs = []string{
	"_design/cards",
	bundleID,
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.71ARDtSu7S-pF3Lsys21n8I8g2Y.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.71ARDtSu7S-pF3Lsys21n8I8g2Y.1",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.71ARDtSu7S-pF3Lsys21n8I8g2Y.3",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.71ARDtSu7S-pF3Lsys21n8I8g2Y.4",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.1",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.aucxuxHEw1A-0ziIaL02Qzh70nY.0",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.aucxuxHEw1A-0ziIaL02Qzh70nY.1",
	"card-alnlcvykyjxsjtijzonc3456kd5u4757.aucxuxHEw1A-0ziIaL02Qzh70nY.2",
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
	bundleID,
	"deck-MoAm80CALRtrMk7Y4eOmGtCzzjY",
	"deck-sXxr9js6DFSFotJ6yuISxOQCuKU",
	"note-71ARDtSu7S-pF3Lsys21n8I8g2Y",
	"note-ZR4TpeX38xRzRvXprlgJpP4Ribo",
	"note-aucxuxHEw1A-0ziIaL02Qzh70nY",
	"note-efn_5zJV184Q7hZzE8zmlclqllY",
	"note-l2ZvABKQnCBZbuaYOlSIrZWqRQI",
	"note-qT9Gr_a9D_jkaapw1xy7KYfvTOs",
	"note-rRm8q5nIKgIMC__jMxYmhXRF_2I",
	"note-udROb8T8RmRASG5zGHNKnKL25zI",
	"theme-0eLFOKi_y8y7JTlmxlVa2Fv7smg",
	"theme-ELr8cEJJOvJU4lYz-VTXhH8wLTo",
	"theme-jZ8Wj_XJNJksDh9aGzMmbLi-6UE",
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

	user := &repo.User{User: testUser}
	err = repo.Import(user, fbb)
	require.Nil(err, "Error importing file: %+v", err)

	importComplete = true
}

func TestImport(t *testing.T) {
	require := require.New(t)

	testImport(t)

	user := &repo.User{User: testUser}
	udb, err := user.DB()
	require.Nil(err, "Error connecting to User DB: %s", err)

	rows, err := udb.AllDocs(context.TODO())
	if err != nil {
		t.Fatalf("AllDocs failed for userdb: %s", err)
	}
	var uids []string
	for rows.Next() {
		uids = append(uids, rows.ID())
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("Userdb AllDocs iteration failed: %s", err)
	}

	require.DeepEqual(expectedUserDBIDs, uids, "User DB IDs")

	bdb, err := user.NewDB(bundleID)
	require.Nil(err, "Error connecting to Bundle DB: %s", err)

	rows, err = bdb.AllDocs(context.TODO())
	if err != nil {
		t.Fatalf("AllDocs failed for bundle: %s", err)
	}
	var bids []string
	for rows.Next() {
		bids = append(bids, rows.ID())
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("Bundle AllDocs iteration failed: %s", err)
	}
	require.DeepEqual(expectedBundleDBIDs, bids, "Bundle DB IDs")
}
