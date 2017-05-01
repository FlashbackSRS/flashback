package test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/flimzy/kivik"
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
	client, _ := kivik.New(context.TODO(), "pouch", "")
	db, _ := client.DB(context.TODO(), "db", map[string]interface{}{"db": memdown})
	return &repo.DB{
		DB:     db,
		DBName: "db",
	}
}

func TestRepo(t *testing.T) {
	require := require.New(t)
	fbb, err := ioutil.ReadFile(fbbFile)
	require.Nil(err, "Error reading %s: %s", fbbFile, err)

	pkg := &fb.Package{}
	err = json.Unmarshal(fbb, pkg)
	require.Nil(err, "Error unmarshaling Art.fbb: %s (Did you forget to ungzip it?)", err)

	db := DB()
	th := pkg.Themes[0]
	err = db.Save(th)
	require.Nil(err, "Error saving theme: %s", err)

	var i interface{}
	row, err := db.Get(context.TODO(), th.DocID())
	if err != nil {
		t.Fatalf("Failed to fetch theme doc: %s", err)
	}
	if err := row.ScanDoc(&i); err != nil {
		t.Errorf("Failied to scan theme doc: %s", err)
	}

	e := Expected(th.DocID(), i.(map[string]interface{})["_rev"].(string))
	require.DeepEqualJSON(e, i, "Theme")
}
