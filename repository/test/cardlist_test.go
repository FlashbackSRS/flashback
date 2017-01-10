package test

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pborman/uuid"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/repository"
)

var listUser *fb.User
var listImportComplete bool
var listImportMu sync.Mutex

func init() {
	// pouchdb.Debug("pouchdb:find")
	u, err := fb.NewUser(uuid.Parse("9d11d024-a100-4045-a5b7-9f1ccf96cc9e"), "mr_jones")
	if err != nil {
		panic(fmt.Sprintf("Error creating user: %s\n", err))
	}
	listUser = u
}

func parseTime(ts string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", ts)
	return t
}

func listImport(t *testing.T) {
	listImportMu.Lock()
	defer listImportMu.Unlock()
	if listImportComplete {
		return
	}
	fbb, err := os.Open(fbbFile)
	if err != nil {
		t.Fatalf("Error reading %s: %s", fbbFile, err)
	}

	user := &repo.User{listUser}
	if err := repo.Import(user, fbb); err != nil {
		t.Fatalf("Error importing file: %s", err)
	}

	db, _ := user.DB()

	for _, card := range cards {
		c, err := fb.NewCard("theme-0000000000", 0, card.ID)
		if err != nil {
			t.Errorf("error creating card %s: %s\n", card.ID, err)
		}
		due := parseTime(card.Due)
		c.Due = &due
		ivl, err := time.ParseDuration(card.Interval)
		if err != nil {
			t.Errorf("Invalid interval '%s': %s\n", card.Interval, err)
		}
		c.Interval = &ivl
		if _, err := db.Put(c); err != nil {
			t.Fatalf("Failed to put additional card: %s", err)
		}
	}

	listImportComplete = true
}

type testCard struct {
	ID       string
	Due      string
	Interval string
}

var cards = []testCard{
	testCard{
		ID:       "card-0000.0000.0",
		Due:      "2017-01-01 00:00:00",
		Interval: "24h",
	},
	testCard{
		ID:       "card-0000.0001.0",
		Due:      "2017-01-01 00:00:00",
		Interval: "48h",
	},
	testCard{
		ID:       "card-0000.0002.0",
		Due:      "2016-12-31 00:00:00",
		Interval: "48h",
	},
}

var now = parseTime("2017-01-01 00:00:00")

func TestCardList(t *testing.T) {
	listImport(t)

	u := repo.User{listUser}
	db, _ := u.DB()

	cl, err := repo.GetCards(db, now, 100)
	if err != nil {
		t.Fatalf("GetCards() failed: %s", err)
	}

	expectedCount := 15
	if len(cl) != expectedCount {
		t.Errorf("Expected %d results, got %d\n", expectedCount, len(cl))
	}
	expectedOrder := []string{"card-0000.0002.0", "card-0000.0001.0", "card-0000.0000.0", "card-alnlcvykyjxsjtijzonc3456kd5u4757.udROb8T8RmRASG5zGHNKnKL25zI.0", "card-alnlcvykyjxsjtijzonc3456kd5u4757.rRm8q5nIKgIMC__jMxYmhXRF_2I.0"}
	for i, exp := range expectedOrder {
		if cl[i].DocID() != exp {
			t.Errorf("Position %d expected %s, got %s\n", i, exp, cl[i].DocID())
		}
	}
}
