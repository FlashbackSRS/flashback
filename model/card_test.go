package model

import (
	"context"
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
			repo: &Repo{user: "mjxwe", local: func() kivikClient {
				c := testClient(t)
				if e := c.CreateDB(context.Background(), "user-mjxwe"); e != nil {
					t.Fatal(e)
				}
				return c
			}()},
			err: func() string {
				if env == "js" {
					return "newCards: query failed: not_found: missing"
				}
				return "newCards: query failed: kivik: not yet implemented in memory driver"
			}(),
		},
		{
			name: "no cards",
			repo: &Repo{user: "mjxwe", local: &gctsClient{
				db: &gctsDB{
					q: &mockQuerier{rows: map[string]*mockRows{
						"newCards": &mockRows{},
						"oldCards": &mockRows{},
					}},
				}},
			},
		},
		{
			name: "success",
			repo: &Repo{
				user: "bob",
				local: &gctsClient{
					db: &gctsDB{
						q:     &mockQuerier{rows: map[string]*mockRows{"newCards": &mockRows{rows: storedCards[1:2]}}},
						note:  `{"_id":"note-Zm9v", "theme":"theme-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z"}`,
						theme: `{"_id":"theme-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "_attachments":{}, "files":[], "modelSequence":1, "models": [{"id":0, "files":[], "modelType":"foo"}]}`,
					},
				},
			},
			expected: map[string]interface{}{"id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.1", "model": 0},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.repo.getCardToStudy(context.Background())
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
			err: `newCards: invalid character 'i' looking for beginning of value`,
		},
		{
			name: "old query failure",
			db: &mockQuerier{rows: map[string]*mockRows{
				"oldCards": &mockRows{rows: []string{"invalid json"}},
			}},
			err: `oldCards: invalid character 'i' looking for beginning of value`,
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
		card     *Card
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
		card     *Card
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
// 		card     *Card
// 		expected interface{}
// 		err      string
// 	}
// 	tests := []gnTest{
// 		{
// 			name: "not logged in",
// 			card: &Card{repo: &Repo{
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
		card     *Card
		client   kivikClient
		expected *Card
		err      string
	}
	tests := []cfTest{
		{
			name:     "already loaded",
			card:     &Card{Card: &fb.Card{ID: "card-foo.bar.0"}, note: &fbNote{}},
			expected: &Card{Card: &fb.Card{ID: "card-foo.bar.0"}, note: &fbNote{}},
		},
		{
			name:   "db error",
			card:   &Card{Card: &fb.Card{ID: "card-foo.bar.0"}},
			client: &cfClient{dbErr: errors.New("db error")},
			err:    "db error",
		},
		{
			name:   "note err",
			card:   &Card{Card: &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-foo/0"}},
			client: &cfClient{db: &gctsDB{note: "invalid json"}},
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name:   "theme err",
			card:   &Card{Card: &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-foo/0"}},
			client: &cfClient{db: &gctsDB{note: `{}`, theme: "bad json"}},
			err:    "id required",
		},
		{
			name:   "corrupt theme",
			card:   &Card{Card: &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-Zm9v/0"}},
			client: &cfClient{db: &gctsDB{note: `{"_id":"note-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z"}`, theme: `{"_id":"theme-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "_attachments":{}, "files":[], "modelSequence":1}`}},
			err:    "card's theme has no model",
		},
		{
			name:   "valid",
			card:   &Card{Card: &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-Zm9v/0"}},
			client: &cfClient{db: &gctsDB{note: `{"_id":"note-Zm9v", "theme":"theme-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z"}`, theme: `{"_id":"theme-Zm9v", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "_attachments":{}, "files":[], "modelSequence":1, "models":[{"id":0,"files":[], "modelType":"foo"}]}`}},
			expected: func() *Card {
				themeAtt := fb.NewFileCollection()
				theme := &fb.Theme{
					ID:            "theme-Zm9v",
					Created:       parseTime(t, "2017-01-01T01:01:01Z"),
					Modified:      parseTime(t, "2017-01-01T01:01:01Z"),
					Attachments:   themeAtt,
					Files:         themeAtt.NewView(),
					ModelSequence: 1,
				}
				model := &fb.Model{
					ID:    0,
					Type:  "foo",
					Theme: theme,
					Files: themeAtt.NewView(),
				}
				theme.Models = []*fb.Model{model}
				return &Card{
					Card: &fb.Card{ID: "card-foo.bar.0", ModelID: "theme-Zm9v/0"},
					note: &fbNote{Note: &fb.Note{
						ID:          "note-Zm9v",
						ThemeID:     "theme-Zm9v",
						Created:     parseTime(t, "2017-01-01T01:01:01Z"),
						Modified:    parseTime(t, "2017-01-01T01:01:01Z"),
						Attachments: fb.NewFileCollection(),
						Model:       model,
					}},
					model: &fbModel{
						Model: model,
					},
				}
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.card.fetch(context.Background(), test.client)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if test.card.model != nil {
				if test.card.model.db == nil {
					t.Fatalf("db not set")
				}
				test.card.model.db = nil
			}
			if d := diff.Interface(test.expected, test.card); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCardBody(t *testing.T) {
	tests := []struct {
		name     string
		card     *Card
		face     int
		expected string
		err      string
	}{
		{
			name: "unknown face",
			face: 3,
			err:  "unrecognized card face 3",
		},
		{
			name: "unfetched card",
			card: &Card{},
			err:  "card hasn't been fetched",
		},
		{
			name: "controller error",
			card: &Card{
				note:  &fbNote{},
				model: &fbModel{Model: &fb.Model{Type: "foo"}},
			},
			err: "failed to get FuncMap: ModelController for 'foo' not found",
		},
		{
			name: "template error",
			card: func() *Card {
				theme, _ := fb.NewTheme("theme-Zm9v")
				return &Card{
					note:  &fbNote{},
					model: &fbModel{Model: &fb.Model{Type: "basic", Files: theme.Attachments.NewView()}},
				}
			}(),
			err: "failed to generate template: main template '$template.0.html' not found in model",
		},
		{
			name: "success",
			card: func() *Card {
				theme, _ := fb.NewTheme("theme-Zm9v")
				modelFiles := theme.Attachments.NewView()
				_ = modelFiles.AddFile("$template.0.html", "text/html", []byte(`<body><div class="question" data-id="0">foo</div></body>`))
				model := &fb.Model{
					Theme: theme,
					Type:  "basic",
					Files: modelFiles,
				}
				return &Card{
					Card: &fb.Card{
						ID:       "card-foo.bar.0",
						Created:  time.Now(),
						Modified: time.Now(),
					},
					note:   &fbNote{Note: &fb.Note{ID: "note-Zm9v"}},
					model:  &fbModel{Model: model},
					appURL: "http://foo.com/",
				}
			}(),
			expected: `<!DOCTYPE html><html><head>
	<title>FB Card</title>
	<base href="http://foo.com/"/>
	<meta charset="UTF-8"/>
	<link rel="stylesheet" type="text/css" href="css/cardframe.css"/>
<script type="text/javascript">
'use strict';
var FB = {
	face:  0 ,
	card: {"id":"card-foo.bar.0","model":0},
	note: {"id":"note-Zm9v"}
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
<script type="text/javascript">alert('Hi!');</script></head>
<body class="card card1"><form id="mainform">foo</form></body></html>
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.card.Body(context.Background(), test.face)
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

type errReader struct{ err error }

var _ io.Reader = &errReader{}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

func TestCardPrepareBody(t *testing.T) {
	tests := []struct {
		name         string
		cardFace     string
		templateID   uint32
		iframeScript string
		r            io.Reader
		expected     string
		err          string
	}{
		{
			name: "goquery failure",
			r:    &errReader{err: errors.New("read fail")},
			err:  "goquery parse: read fail",
		},
		{
			name: "no div",
			r:    strings.NewReader(""),
			err:  "No div matching 'div.[data-id='0']' found in template output",
		},
		{
			name:     "success",
			cardFace: "question",
			r:        strings.NewReader(`<body><div class="question" data-id="0">foo</div></body>`),
			expected: `<html><head><script type="text/javascript"></script></head><body class="card card1"><form id="mainform">foo</form></body></html>`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := prepareBody(test.cardFace, test.templateID, test.iframeScript, test.r)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Text(test.expected, string(result)); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestRelatedKeyRange(t *testing.T) {
	startExp := "card-foo.bar."
	endExp := "card-foo.bar." + string(rune(0x10FFFF))
	start, end := relatedKeyRange("card-foo.bar.0")
	if start != startExp {
		t.Errorf("Unexpected start key: %s", start)
	}
	if end != endExp {
		t.Errorf("Unexpected end key: %s", end)
	}
}
