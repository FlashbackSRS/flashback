package pouchdb

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
	"github.com/go-kivik/pouchdb/bindings"
)

var _ driver.Finder = &db{}

// buildIndex merges the ddoc and name into the index structure, as reqiured
// by the PouchDB-find plugin.
func buildIndex(ddoc, name string, index interface{}) (*js.Object, error) {
	i, err := bindings.Objectify(index)
	if err != nil {
		return nil, err
	}
	o := js.Global.Get("Object").New(i)
	if ddoc != "" {
		o.Set("ddoc", ddoc)
	}
	if name != "" {
		o.Set("name", name)
	}
	return o, nil
}

func (d *db) CreateIndex(ctx context.Context, ddoc, name string, index interface{}) error {
	indexObj, err := buildIndex(ddoc, name, index)
	if err != nil {
		return err
	}
	_, err = d.db.CreateIndex(ctx, indexObj)
	return err
}

func (d *db) GetIndexes(ctx context.Context) (indexes []driver.Index, err error) {
	defer bindings.RecoverError(&err)
	result, err := d.db.GetIndexes(ctx)
	if err != nil {
		return nil, err
	}
	// This might not be the most efficient, but it's easy
	var final struct {
		Indexes []driver.Index `json:"indexes"`
	}
	err = json.Unmarshal([]byte(js.Global.Get("JSON").Call("stringify", result).String()), &final)
	return final.Indexes, err
}

// findIndex attempts to find the requested index definition
func (d *db) findIndex(ctx context.Context, ddoc, name string) (interface{}, error) {
	ddoc = "_design/" + strings.TrimPrefix(ddoc, "_design/")
	indexes, err := d.GetIndexes(ctx)
	if err != nil {
		return nil, err
	}
	for _, idx := range indexes {
		if idx.Type == "special" {
			continue
		}
		if idx.DesignDoc == ddoc && idx.Name == name {
			return map[string]interface{}{
				"ddoc": idx.DesignDoc,
				"name": idx.Name,
				"type": idx.Type,
				"def":  idx.Definition,
			}, nil
		}
	}
	return nil, errors.Status(kivik.StatusNotFound, "index does not exist")
}

func (d *db) DeleteIndex(ctx context.Context, ddoc, name string) error {
	index, err := d.findIndex(ctx, ddoc, name)
	if err != nil {
		return err
	}
	_, err = d.db.DeleteIndex(ctx, index)
	return err
}

func (d *db) Find(ctx context.Context, query interface{}) (driver.Rows, error) {
	result, err := d.db.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	return &findRows{
		Object: result,
	}, nil
}

type findRows struct {
	*js.Object
}

var _ driver.Rows = &findRows{}

func (r *findRows) Offset() int64     { return 0 }
func (r *findRows) TotalRows() int64  { return 0 }
func (r *findRows) UpdateSeq() string { return "" }
func (r *findRows) Warning() string {
	if w := r.Get("warning"); w != js.Undefined {
		return w.String()
	}
	return ""
}

func (r *findRows) Close() error {
	r.Delete("docs") // Free up memory used by any remaining rows
	return nil
}

func (r *findRows) Next(row *driver.Row) (err error) {
	defer bindings.RecoverError(&err)
	if r.Get("docs") == js.Undefined || r.Get("docs").Length() == 0 {
		return io.EOF
	}
	next := r.Get("docs").Call("shift")
	row.Doc = json.RawMessage(jsJSON.Call("stringify", next).String())
	return nil
}

type queryPlan struct {
	DBName   string                 `json:"dbname"`
	Index    map[string]interface{} `json:"index"`
	Selector map[string]interface{} `json:"selector"`
	Options  map[string]interface{} `json:"opts"`
	Limit    int64                  `json:"limit"`
	Skip     int64                  `json:"skip"`
	Fields   []interface{}          `json:"fields"`
	Range    map[string]interface{} `json:"range"`
}

func (d *db) Explain(ctx context.Context, query interface{}) (*driver.QueryPlan, error) {
	result, err := d.db.Explain(ctx, query)
	if err != nil {
		return nil, err
	}
	planJSON := js.Global.Get("JSON").Call("stringify", result).String()
	var plan queryPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return nil, err
	}
	return &driver.QueryPlan{
		DBName:   plan.DBName,
		Index:    plan.Index,
		Selector: plan.Selector,
		Options:  plan.Options,
		Limit:    plan.Limit,
		Skip:     plan.Skip,
		Fields:   plan.Fields,
		Range:    plan.Range,
	}, nil

}
