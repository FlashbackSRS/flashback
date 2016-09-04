package test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	. "github.com/FlashbackSRS/flashback-model/test/util"
	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
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
	fbb, err := ioutil.ReadFile("Art.fbb")
	if err != nil {
		t.Fatalf("Error reading Art.fbb: %s", err)
	}

	pkg := &fb.Package{}
	err = json.Unmarshal(fbb, pkg)
	if err != nil {
		t.Fatalf("Error unmarshaling Art.fbb: %s", err)
	}

	db := DB()
	th := pkg.Themes[0]
	err = db.Save(th)
	if err != nil {
		t.Fatalf("Error saving theme: %s", err)
	}

	var i interface{}
	if err := db.Get(th.DocID(), &i, pouchdb.Options{}); err != nil {
		t.Fatalf("Error re-fetching Theme: %s", err)
	}
	e := Expected(th.DocID(), i.(map[string]interface{})["_rev"].(string))

	DeepEqualJSON(t, "Theme", i, e)

	// if !reflect.DeepEqual(th, fth) {
	// 	fmt.Printf(" th: %+v\n", th)
	// 	fmt.Printf("fth: %+v\n", fth)
	// 	t.Fatalf("Retrieved theme does not match")
	// }
	// fmt.Printf("fth: %+v\n", fth)
}
