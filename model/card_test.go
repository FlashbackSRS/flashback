package model

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
)

func TestNewQuerier(t *testing.T) {
	db := &kivik.DB{}
	q := newQuerier(db)
	if qw, ok := q.(*querierWrapper); ok {
		if qw.DB != db {
			t.Errorf("Unexpected result")
		}
	} else {
		t.Errorf("Unexpected type")
	}
}

func TestQuerierQuery(t *testing.T) {
	db := testDB(t)
	q := newQuerier(db)
	_, err := q.Query(context.Background(), "", "")
	checkErr(t, "kivik: not yet implemented in memory driver", err)
}

type mockQuerier struct {
	rows kivikRows
	err  error
}

var _ querier = &mockQuerier{}

func (db *mockQuerier) Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error) {
	return db.rows, db.err
}

type mockRows struct {
	rows  []string
	total int64
}

var _ kivikRows = &mockRows{}

func (r *mockRows) Close() error     { return nil }
func (r *mockRows) Next() bool       { return len(r.rows) > 0 }
func (r *mockRows) TotalRows() int64 { return r.total }
func (r *mockRows) ScanDoc(d interface{}) error {
	if len(r.rows) == 0 {
		return io.EOF
	}
	if err := json.Unmarshal([]byte(r.rows[0]), &d); err != nil {
		return err
	}
	r.rows = r.rows[1:]
	return nil
}

func TestGetCards(t *testing.T) {
	type gcfvTest struct {
		name     string
		db       querier
		view     string
		limit    int
		expected interface{}
		err      string
	}
	tests := []gcfvTest{
		{
			name: "query error",
			db:   &mockQuerier{err: errors.New("query failed")},
			err:  "query failed",
		},
		{
			name: "invalid JSON",
			db: &mockQuerier{rows: &mockRows{
				total: 1,
				rows:  []string{"foo"},
			}},
			err: "invalid character 'o' in literal false (expecting 'a')",
		},
		{
			name:     "no results",
			db:       &mockQuerier{rows: &mockRows{total: 0}},
			expected: []*fb.Card{},
		},
		{
			name: "successful fetch",
			db: &mockQuerier{rows: &mockRows{
				total: 1,
				rows: []string{`{"type": "card", "_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.0", "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8", "created": "2016-07-31T15:08:24.730156517Z", "modified": "2016-07-15T15:07:24.730156517Z", "model": "theme-VGVzdCBUaGVtZQ/0", "buriedUntil": "2099-01-01"}`,
					`{"type": "card", "_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.1", "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8", "created": "2016-07-31T15:08:24.730156517Z", "modified": "2016-07-31T15:08:24.730156517Z", "model": "theme-VGVzdCBUaGVtZQ/0"}`,
					`{"type": "card", "_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.2", "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8", "created": "2016-07-31T15:08:24.730156517Z", "modified": "2016-07-31T15:08:24.730156517Z", "model": "theme-VGVzdCBUaGVtZQ/0"}`,
				},
			}},
			expected: []map[string]interface{}{
				{
					"type":     "card",
					"_id":      "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.1",
					"_rev":     "1-6e1b6fb5352429cf3013eab5d692aac8",
					"created":  "2016-07-31T15:08:24.730156517Z",
					"modified": "2016-07-31T15:08:24.730156517Z",
					"model":    "theme-VGVzdCBUaGVtZQ/0",
				},
				{
					"type":     "card",
					"_id":      "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.2",
					"_rev":     "1-6e1b6fb5352429cf3013eab5d692aac8",
					"created":  "2016-07-31T15:08:24.730156517Z",
					"modified": "2016-07-31T15:08:24.730156517Z",
					"model":    "theme-VGVzdCBUaGVtZQ/0",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cards, err := getCardsFromView(context.Background(), test.db, test.view, test.limit)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, cards); d != "" {
				t.Error(d)
			}
		})
	}
}
