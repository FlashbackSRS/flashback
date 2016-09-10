package test

import (
	"os"
	"testing"

	"github.com/flimzy/testify/require"

	"github.com/flimzy/go-pouchdb"
	"github.com/gopherjs/gopherjs/js"
	"github.com/pborman/uuid"

	"github.com/FlashbackSRS/flashback-model"
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
	"bundle-2trorbeivl7faygmo6hbpx7azgyzu43g",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.AqMTKqXMHUFjJmB8eJW55Uf9Okk.0",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.AqMTKqXMHUFjJmB8eJW55Uf9Okk.1",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.BD97IR9wXQE0ILn_T2wkgOynng8.0",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.BD97IR9wXQE0ILn_T2wkgOynng8.1",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.PSVLhct0aM92bz4wJVDQf1zF7R0.0",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.PSVLhct0aM92bz4wJVDQf1zF7R0.1",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.YqTDhtD-MFNfFXrORP7R-VBF8LQ.0",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.YqTDhtD-MFNfFXrORP7R-VBF8LQ.1",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.d2y5jA1tO727U-6lFh8_Bh7IEjw.0",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.d2y5jA1tO727U-6lFh8_Bh7IEjw.1",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.yTSB3AjHJ4eYIruVlRYl-9Vawrw.0",
	"card-2trorbeivl7faygmo6hbpx7azgyzu43g.yTSB3AjHJ4eYIruVlRYl-9Vawrw.1",
}

var expectedBundleDBIDs = []string{
	"_design/idx-691fd0e525e654428e875bcb3aacb6ac",
	"bundle-2trorbeivl7faygmo6hbpx7azgyzu43g",
	"deck--4EyzQLG50dc0JcDKOjV1WS4ulk",
	"deck-m7om6hW6WIc_m6xrBDUJMiG9X2o",
	"note-AqMTKqXMHUFjJmB8eJW55Uf9Okk",
	"note-BD97IR9wXQE0ILn_T2wkgOynng8",
	"note-PSVLhct0aM92bz4wJVDQf1zF7R0",
	"note-YqTDhtD-MFNfFXrORP7R-VBF8LQ",
	"note-d2y5jA1tO727U-6lFh8_Bh7IEjw",
	"note-yTSB3AjHJ4eYIruVlRYl-9Vawrw",
	"theme-uQ3TFsQgm9Y29vlgC-lphauhK3M",
}

func TestImport(t *testing.T) {
	require := require.New(t)
	fbb, err := os.Open(fbbFile)
	if err != nil {
		t.Fatalf("Error reading %s: %s", fbbFile, err)
	}

	u, err := fb.NewUser(uuid.Parse("9d11d024-a100-4045-a5b7-9f1ccf96cc9f"), "mrsmith")
	if err != nil {
		t.Fatalf("Error creating user: %s\n", err)
	}

	user := &repo.User{u}
	udb, err := user.DB()
	if err != nil {
		t.Fatalf("Error connecting to User DB: %s", err)
	}

	if err := repo.Import(user, fbb); err != nil {
		t.Fatalf("Error importing file: %+v", err)
	}

	var allUserDocs, allBundleDocs map[string]interface{}
	if err := udb.AllDocs(&allUserDocs, pouchdb.Options{}); err != nil {
		t.Fatalf("Error fetching AllDocs(): %s", err)
	}
	// Remove the revs, because they're random
	urows := allUserDocs["rows"].([]interface{})
	uids := make([]string, len(urows))
	for i, row := range urows {
		uids[i] = row.(map[string]interface{})["id"].(string)
	}
	require.DeepEqual(expectedUserDBIDs, uids, "User DB IDs")

	bdb, err := repo.NewDB("bundle-2trorbeivl7faygmo6hbpx7azgyzu43g")
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
