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

var expectedUserDBIDs = []string{
	"_design/idx-691fd0e525e654428e875bcb3aacb6ac",
	"bundle-iw5x7ie66fsepm67hey2fqjms6fywi6v",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.4WpHslICjKMtkmw-KKpSJECrnuc.0",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.4WpHslICjKMtkmw-KKpSJECrnuc.1",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.Kzup-0SvVwg3DqbkLk7-bqBFBBo.0",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.Kzup-0SvVwg3DqbkLk7-bqBFBBo.1",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.O6ZKmF5cdjqlCabeBQScdMpMYas.0",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.O6ZKmF5cdjqlCabeBQScdMpMYas.1",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.aGSlwlH5Qqi6wFUSPvohUjYRH5M.0",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.aGSlwlH5Qqi6wFUSPvohUjYRH5M.1",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.aXWDFzKQPNJ-oA3vVJF16PvsE5E.0",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.aXWDFzKQPNJ-oA3vVJF16PvsE5E.1",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.mrHVpnN28q1ekKVF-qYfwolFNzM.0",
	"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.mrHVpnN28q1ekKVF-qYfwolFNzM.1",
}

var expectedBundleDBIDs = []string{
	"_design/idx-691fd0e525e654428e875bcb3aacb6ac",
	"bundle-iw5x7ie66fsepm67hey2fqjms6fywi6v",
	"deck-HXlOHW5PP55PRvZPPZ76YyhOI0w",
	"deck-y83bDbYxPI0tF7vUsqbvArJR-KY",
	"note-4WpHslICjKMtkmw-KKpSJECrnuc",
	"note-Kzup-0SvVwg3DqbkLk7-bqBFBBo",
	"note-O6ZKmF5cdjqlCabeBQScdMpMYas",
	"note-aGSlwlH5Qqi6wFUSPvohUjYRH5M",
	"note-aXWDFzKQPNJ-oA3vVJF16PvsE5E",
	"note-mrHVpnN28q1ekKVF-qYfwolFNzM",
	"theme-94hk99pCpQ5DAMGZvpb5_HR5oqs",
}

var importComplete bool
var importMu sync.Mutex

func testImport(t *testing.T) {
	importMu.Lock()
	if importComplete {
		return
	}
	defer importMu.Unlock()
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

	bdb, err := repo.NewDB("bundle-iw5x7ie66fsepm67hey2fqjms6fywi6v")
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
