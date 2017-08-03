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
	rows *mockRows
	err  error
}

var _ querier = &mockQuerier{}

func (db *mockQuerier) Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error) {
	if db.err != nil {
		return nil, db.err
	}
	limit, _ := options[0]["limit"].(int)
	offset, _ := options[0]["offset"].(int)
	rows := db.rows
	rows.limit = limit + offset
	rows.i = offset
	return rows, nil
}

type mockRows struct {
	rows     []string
	i, limit int
}

var _ kivikRows = &mockRows{}

func (r *mockRows) Close() error     { return nil }
func (r *mockRows) Next() bool       { return r.i < r.limit && r.i < len(r.rows) }
func (r *mockRows) TotalRows() int64 { return int64(len(r.rows)) }
func (r *mockRows) ScanDoc(d interface{}) error {
	if r.i > r.limit || r.i >= len(r.rows) {
		return io.EOF
	}
	if err := json.Unmarshal([]byte(r.rows[r.i]), &d); err != nil {
		return err
	}
	r.i++
	return nil
}

var storedCards = []string{
	`{"type": "card", "_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.0", "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8", "created": "2016-07-31T15:08:24.730156517Z", "modified": "2016-07-15T15:07:24.730156517Z", "model": "theme-VGVzdCBUaGVtZQ/0", "buriedUntil": "2099-01-01"}`,
	`{"type": "card", "_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.1", "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8", "created": "2016-07-31T15:08:24.730156517Z", "modified": "2016-07-31T15:08:24.730156517Z", "model": "theme-VGVzdCBUaGVtZQ/0"}`,
	`{"type": "card", "_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.2", "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8", "created": "2016-07-31T15:08:24.730156517Z", "modified": "2016-07-31T15:08:24.730156517Z", "model": "theme-VGVzdCBUaGVtZQ/0"}`,
}

var expectedCards = []map[string]interface{}{
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
}

func TestGetCardsFromView(t *testing.T) {
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
			name:  "query error",
			db:    &mockQuerier{err: errors.New("query failed")},
			limit: 1,
			err:   "query failed",
		},
		{
			name: "invalid JSON",
			db: &mockQuerier{rows: &mockRows{
				rows: []string{"foo"},
			}},
			limit: 1,
			err:   "invalid character 'o' in literal false (expecting 'a')",
		},
		{
			name:     "no results",
			db:       &mockQuerier{rows: &mockRows{}},
			limit:    1,
			expected: []*fb.Card{},
		},
		{
			name:     "successful fetch",
			db:       &mockQuerier{rows: &mockRows{rows: storedCards}},
			limit:    10,
			expected: expectedCards,
		},
		{
			name:  "limit 0",
			db:    &mockQuerier{},
			limit: 0,
			err:   "invalid limit",
		},
		{
			name:     "limit reached",
			db:       &mockQuerier{rows: &mockRows{rows: storedCards}},
			limit:    1,
			expected: expectedCards[0:1],
		},
		{
			name: "aggregate necessary",
			db: &mockQuerier{rows: &mockRows{rows: func() []string {
				rows := make([]string, 150)
				for i := 0; i < 150; i++ {
					rows[i] = storedCards[0]
				}
				return append(rows, storedCards...)
			}()}},
			limit:    5,
			expected: expectedCards,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cards, err := getCardsFromView(context.Background(), test.db, test.view, test.limit, 0)
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
