package test

import (
	"context"
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

	user := &repo.User{User: listUser}
	if err := repo.Import(user, fbb); err != nil {
		t.Fatalf("Error importing file: %s", err)
	}

	db, _ := user.DB()

	for _, card := range cards {
		c, err := fb.NewCard("theme-0000000000", 0, card.ID)
		if err != nil {
			t.Errorf("error creating card %s: %s\n", card.ID, err)
		}
		if card.Due != "" {
			due, _ := fb.ParseDue(card.Due)
			c.Due = &due
		}
		if card.Interval != "" {
			ivl, err := fb.ParseInterval(card.Interval)
			if err != nil {
				t.Errorf("Invalid interval '%s': %s\n", card.Interval, err)
			}
			c.Interval = &ivl
		}
		c.Created = time.Now()
		if _, err := db.Put(context.TODO(), c.DocID(), c); err != nil {
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
		ID:       "card-0000.0002.0",
		Due:      "2016-12-31 00:00:00",
		Interval: "48h",
	},
}

var now = parseTime("2017-01-01 00:00:00")

func TestCardList(t *testing.T) {
	listImport(t)

	u := repo.User{User: listUser}
	db, _ := u.DB()

	cl, err := repo.GetCardList(db, 50)
	if err != nil {
		t.Fatalf("GetCards() failed: %s", err)
	}

	expectedCount := 14
	if len(cl) != expectedCount {
		t.Errorf("Expected %d results, got %d\n", expectedCount, len(cl))
	}
	expectedOrder := []string{
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.udROb8T8RmRASG5zGHNKnKL25zI.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.rRm8q5nIKgIMC__jMxYmhXRF_2I.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.efn_5zJV184Q7hZzE8zmlclqllY.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.qT9Gr_a9D_jkaapw1xy7KYfvTOs.0",
		"card-0000.0002.0",
		"card-0000.0000.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.71ARDtSu7S-pF3Lsys21n8I8g2Y.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.71ARDtSu7S-pF3Lsys21n8I8g2Y.3",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.aucxuxHEw1A-0ziIaL02Qzh70nY.1",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.71ARDtSu7S-pF3Lsys21n8I8g2Y.4",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.aucxuxHEw1A-0ziIaL02Qzh70nY.2",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.aucxuxHEw1A-0ziIaL02Qzh70nY.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.71ARDtSu7S-pF3Lsys21n8I8g2Y.1",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.udROb8T8RmRASG5zGHNKnKL25zI.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.rRm8q5nIKgIMC__jMxYmhXRF_2I.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.efn_5zJV184Q7hZzE8zmlclqllY.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.qT9Gr_a9D_jkaapw1xy7KYfvTOs.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.udROb8T8RmRASG5zGHNKnKL25zI.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.rRm8q5nIKgIMC__jMxYmhXRF_2I.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.efn_5zJV184Q7hZzE8zmlclqllY.0",
		"card-alnlcvykyjxsjtijzonc3456kd5u4757.qT9Gr_a9D_jkaapw1xy7KYfvTOs.0",
	}
	for i := range cl {
		if cl[i].ID != expectedOrder[i] {
			t.Errorf("Position %d expected %s, got %s\n", i, expectedOrder[i], cl[i].ID)
		}
	}
}
