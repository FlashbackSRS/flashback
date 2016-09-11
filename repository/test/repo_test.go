package test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/flimzy/testify/require"
	"github.com/gopherjs/gopherjs/js"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/repository"
)

var memdown *js.Object

func init() {
	memdown = js.Global.Call("require", "memdown")
	js.Global.Call("require", "pouchdb")
	js.Global.Call("require", "pouchdb-find")
}

func DB() *repo.DB {
	db := pouchdb.NewWithOpts("db", pouchdb.Options{DB: memdown})
	return &repo.DB{
		PouchDB:         db,
		PouchPluginFind: find.New(db),
		DBName:          "db",
	}
}

func TestRepo(t *testing.T) {
	require := require.New(t)
	fbb, err := ioutil.ReadFile(fbbFile)
	require.Nil(err, "Error reading %s: %s", fbbFile, err)

	pkg := &fb.Package{}
	err = json.Unmarshal(fbb, pkg)
	require.Nil(err, "Error unmarshaling Art.fbb: %s", err)

	db := DB()
	th := pkg.Themes[0]
	err = db.Save(th)
	require.Nil(err, "Error saving theme: %s", err)

	var i interface{}
	err = db.Get(th.DocID(), &i, pouchdb.Options{})
	require.Nil(err, "Error re-fetching Theme: %s", err)

	e := Expected(th.DocID(), i.(map[string]interface{})["_rev"].(string))
	require.DeepEqualJSON(e, i, "Theme")
}
