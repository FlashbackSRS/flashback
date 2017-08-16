package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
)

func init() {
	now = func() time.Time {
		t, _ := time.Parse(time.RFC3339, "2017-01-01T12:00:00Z")
		return t
	}
}

type mockQuerier struct {
	rows map[string]*mockRows
	err  error
}

var _ querier = &mockQuerier{}

func (db *mockQuerier) Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error) {
	if db.err != nil {
		return nil, db.err
	}
	limit, _ := options[0]["limit"].(int)
	offset, _ := options[0]["offset"].(int)
	rows, ok := db.rows[view]
	if !ok {
		return &mockRows{}, nil
	}
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
func (r *mockRows) ID() string       { panic("ID() not implemented") }
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
			err:   "query failed: query failed",
		},
		{
			name: "invalid JSON",
			db: &mockQuerier{rows: map[string]*mockRows{
				"test": &mockRows{rows: []string{"foo"}},
			}},
			limit: 1,
			view:  "test",
			err:   "invalid character 'o' in literal false (expecting 'a')",
		},
		{
			name:     "no results",
			db:       &mockQuerier{rows: map[string]*mockRows{}},
			limit:    1,
			expected: []*fb.Card{},
		},
		{
			name: "successful fetch",
			db: &mockQuerier{rows: map[string]*mockRows{
				"test": &mockRows{rows: storedCards},
			}},
			limit:    10,
			view:     "test",
			expected: expectedCards,
		},
		{
			name:  "limit 0",
			db:    &mockQuerier{},
			limit: 0,
			err:   "invalid limit",
		},
		{
			name: "limit reached",
			db: &mockQuerier{rows: map[string]*mockRows{
				"test": &mockRows{rows: storedCards},
			}},
			limit:    1,
			view:     "test",
			expected: expectedCards[0:1],
		},
		{
			name: "aggregate necessary",
			db: &mockQuerier{rows: map[string]*mockRows{
				"test": &mockRows{rows: func() []string {
					rows := make([]string, 150)
					for i := 0; i < 150; i++ {
						rows[i] = storedCards[0]
					}
					return append(rows, storedCards...)
				}()},
			}},
			limit:    5,
			view:     "test",
			expected: expectedCards,
		},
		{
			name: "ignore card seen today",
			db: &mockQuerier{rows: map[string]*mockRows{
				"test": &mockRows{rows: []string{
					`{"type": "card", "_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.1", "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8", "created": "2016-07-31T15:08:24.730156517Z", "modified": "2016-07-31T15:08:24.730156517Z", "model": "theme-VGVzdCBUaGVtZQ/0", "interval": 5, "lastReview": "2017-01-01T11:50:00Z"}`,
				}},
			}},
			limit:    5,
			view:     "test",
			expected: []int{},
		},
		{
			name: "skip same-day forward fuzzing",
			db: &mockQuerier{rows: map[string]*mockRows{
				"test": &mockRows{rows: []string{
					`{"type": "card", "_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.1", "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8", "created": "2016-07-31T15:08:24.730156517Z", "modified": "2016-07-31T15:08:24.730156517Z", "model": "theme-VGVzdCBUaGVtZQ/0", "interval": -1800, "due": "2017-01-01 12:01:00"}`,
				}},
			}},
			limit:    5,
			view:     "test",
			expected: []int{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cards, err := getCardsFromView(context.Background(), test.db, test.view, test.limit, 0)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, cards); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCardPriority(t *testing.T) {
	type cpTest struct {
		due      fb.Due
		interval fb.Interval
		now      time.Time
		expected float64
	}
	tests := []cpTest{
		{
			due:      parseDue(t, "2017-01-01 00:00:00"),
			interval: fb.Day,
			expected: 1,
		},
		{
			due:      parseDue(t, "2017-01-01 12:00:00"),
			interval: fb.Day,
			expected: 0.125,
		},
		{
			due:      parseDue(t, "2016-12-31 12:00:00"),
			interval: fb.Day,
			expected: 3.375,
		},
		{
			due:      parseDue(t, "2017-02-01 00:00:00"),
			interval: 60 * fb.Day,
			expected: 0.112912,
		},
		{
			due:      parseDue(t, "2017-01-02 00:00:00"),
			interval: fb.Day,
			expected: 0,
		},
		{
			due:      parseDue(t, "2016-01-02 00:00:00"),
			interval: 7 * fb.Day,
			expected: 150084.109375,
		},
		{
			due:      parseDue(t, "2017-01-24 11:16:59"),
			interval: 10 * fb.Minute,
			expected: 132.520996,
			now:      parseTime(t, "2017-01-24T11:57:58+01:00"),
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v / %v", test.due, test.interval), func(t *testing.T) {
			nowTime := test.now
			if nowTime.IsZero() {
				nowTime = parseTime(t, "2017-01-01T00:00:00Z")
			}
			prio := cardPriority(test.due, test.interval, nowTime)
			if !floatCompare(float64(prio), test.expected) {
				t.Errorf("Unexpected result %f", prio)
			}
		})
	}
}

func init() {
	rnd = rand.New(rand.NewSource(1))
}

func TestSelectWeightedCard(t *testing.T) {
	type swcTest struct {
		name     string
		cards    []*fb.Card
		expected int
	}
	tests := []swcTest{
		{
			name:     "no cards",
			expected: -1,
		},
		{
			name: "one card",
			cards: []*fb.Card{
				&fb.Card{Rev: "a"},
			},
			expected: 0,
		},
		{
			name: "two equal card",
			cards: []*fb.Card{
				&fb.Card{Rev: "a"},
				&fb.Card{Rev: "b"},
			},
			expected: 1,
		},
		{
			name: "three cards, different prios",
			cards: []*fb.Card{
				&fb.Card{Rev: "a"},
				&fb.Card{Rev: "b"},
				&fb.Card{
					Due:      parseDue(t, "2015-01-01"),
					Interval: fb.Interval(20 * fb.Day),
					Rev:      "c"},
			},
			expected: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := selectWeightedCard(test.cards)
			if test.expected == -1 {
				if result != nil {
					t.Errorf("Expected no result.")
				}
				return
			}
			if test.cards[test.expected] != result {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

type gctsClient struct {
	kivikClient
	db kivikDB
}

func (c *gctsClient) DB(_ context.Context, _ string, _ ...kivik.Options) (kivikDB, error) {
	return c.db, nil
}

type gctsDB struct {
	kivikDB
	q     *mockQuerier
	note  string
	theme string
}

func (db *gctsDB) Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error) {
	return db.q.Query(ctx, ddoc, view, options...)
}

func (db *gctsDB) Get(_ context.Context, id string, _ ...kivik.Options) (kivikRow, error) {
	if strings.HasPrefix(id, "note-") {
		return mockRow(db.note), nil
	}
	return mockRow(db.theme), nil
}

func TestRepoGetCardToStudy(t *testing.T) {
	type rgctsTest struct {
		name     string
		repo     *Repo
		expected interface{}
		err      string
	}
	tests := []rgctsTest{
		{
			name: "not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		{
			name: "logged in",
			repo: &Repo{user: "bob", local: func() kivikClient {
				c := testClient(t)
				if e := c.CreateDB(context.Background(), "user-bob"); e != nil {
					t.Fatal(e)
				}
				return c
			}()},
			err: "query failed: kivik: not yet implemented in memory driver",
		},
		{
			name: "success",
			repo: &Repo{
				user: "bob",
				local: &gctsClient{
					db: &gctsDB{
						q:     &mockQuerier{rows: map[string]*mockRows{"newCards": &mockRows{rows: storedCards[1:2]}}},
						note:  `{"_id":"note-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z"}`,
						theme: `{"_id":"theme-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "_attachments":{}, "files":[]}`,
					},
				},
			},
			expected: expectedCards[0],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.repo.GetCardToStudy(context.Background())
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestGetCardToStudy(t *testing.T) {
	type gctsTest struct {
		name     string
		db       querier
		expected interface{}
		err      string
	}
	tests := []gctsTest{
		{
			name: "no cards",
			db:   &mockQuerier{rows: map[string]*mockRows{}},
		},
		{
			name: "new query failure",
			db: &mockQuerier{rows: map[string]*mockRows{
				"newCards": &mockRows{rows: []string{"invalid json"}},
			}},
			err: `invalid character 'i' looking for beginning of value`,
		},
		{
			name: "old query failure",
			db: &mockQuerier{rows: map[string]*mockRows{
				"oldCards": &mockRows{rows: []string{"invalid json"}},
			}},
			err: `invalid character 'i' looking for beginning of value`,
		},
		{
			name: "one new card",
			db: &mockQuerier{rows: map[string]*mockRows{
				"newCards": &mockRows{rows: storedCards[1:2]},
			}},
			expected: expectedCards[0],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := getCardToStudy(context.Background(), test.db)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

/*
func TestCardButtons(t *testing.T) {
	type cbTest struct {
		name     string
		card     *fbCard
		face     int
		expected studyview.ButtonMap
		err      string
	}
	tests := []cbTest{

	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.card.Buttons(test.face)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
*/

/*
func TestGetCardModel(t *testing.T) {
	type cgmTest struct {
		name     string
		card     *fbCard
		expected interface{}
		err      string
	}
	tests := []cgmTest{
		{}
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.card.model(context.Background())
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
*/

// func TestGetNote(t *testing.T) {
// 	type gnTest struct {
// 		name     string
// 		card     *fbCard
// 		expected interface{}
// 		err      string
// 	}
// 	tests := []gnTest{
// 		{
// 			name: "not logged in",
// 			card: &fbCard{repo: &Repo{
// 				local: testClient(t),
// 			}},
// 			err: "not logged in",
// 		},
// 	}
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			result, err := test.card.getNote(context.Background())
// 			checkErr(t, test.err, err)
// 			if err != nil {
// 				return
// 			}
// 			if d := diff.AsJSON(test.expected, result); d != nil {
// 				t.Error(d)
// 			}
// 		})
// 	}
// }

type cfClient struct {
	kivikClient
	db    kivikDB
	dbErr error
}

func (c *cfClient) DB(_ context.Context, _ string, _ ...kivik.Options) (kivikDB, error) {
	return c.db, c.dbErr
}

func TestCardFetch(t *testing.T) {
	type cfTest struct {
		name     string
		card     *fb.Card
		client   kivikClient
		expected *fbCard
		err      string
	}
	tests := []cfTest{
		{
			name:   "db error",
			card:   &fb.Card{ID: "card-foo.bar.0"},
			client: &cfClient{dbErr: errors.New("db error")},
			err:    "db error",
		},
		{
			name:   "note err",
			card:   &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-foo/0"},
			client: &cfClient{db: &gctsDB{note: "invalid json"}},
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name:   "theme err",
			card:   &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-foo/0"},
			client: &cfClient{db: &gctsDB{note: `{}`, theme: "bad json"}},
			err:    "id required",
		},
		{
			name:     "valid",
			card:     &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-Zm9v/0"},
			client:   &cfClient{db: &gctsDB{note: `{"_id":"note-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z"}`, theme: `{"_id":"theme-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "_attachments":{}, "files":[]}`}},
			expected: &fbCard{Card: &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-Zm9v/0"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &fbCard{Card: test.card}
			err := card.fetch(context.Background(), test.client)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, card); d != nil {
				t.Error(d)
			}
		})
	}
}

/*
func TestCardBody(t *testing.T) {
	type cbTest struct {
		name     string
		card     *fb.Card
		face     int
		expected string
		err      string
	}
	tests := []cbTest{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &fbCard{
				Card: test.card,
			}
			result, err := card.Body(test.face)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Text(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
*/
