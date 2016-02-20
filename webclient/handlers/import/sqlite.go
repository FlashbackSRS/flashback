package import_handler

import (
	"errors"
	"fmt"
	"github.com/flimzy/flashback/anki"
	"github.com/flimzy/web/worker"
	"github.com/gopherjs/gopherjs/js"
	// 	"honnef.co/go/js/console"
)

func readSQLite(dbbuf []byte) (*anki.Collection, error) {
	fmt.Printf("Gonna pretend to read the SQLite data now...\n")

	w := worker.New("js/worker.sql.js")
	defer w.Terminate()
	w.Send(map[string]interface{}{
		"action": "open",
		"buffer": js.NewArrayBuffer(dbbuf),
	})
	response, err := w.Receive()
	if err != nil {
		return nil, err
	}
	if msg, ok := response.(map[string]interface{}); !ok {
		return nil, errors.New("Received unexpected response from sqlite")
	} else if ready, ok := msg["ready"].(bool); !ok || !ready {
		return nil, errors.New("Ready status not true")
	}

	collection, err := readCollections(w)
	if err != nil {
		return nil, err
	}
	if err := readCards(w, collection); err != nil {
		return nil, err
	}
	if err := readNotes(w, collection); err != nil {
		return nil, err
	}
	if err := readRevlog(w, collection); err != nil {
		return nil, err
	}
	if err := readGraves(w, collection); err != nil {
		return nil, err
	}
	return collection, err
}

type rowFunc func(map[string]interface{})

func readCollections(w *worker.Worker) (c *anki.Collection, err error) {
	var read bool
	rowFn := func(row map[string]interface{}) {
		if read {
			err = errors.New("Read more than one collection. This shouldn't happen")
		} else {
			c, err = anki.SqliteToCollection(row)
			read = true
		}
	}
	err2 := readX(w, "SELECT * FROM col", rowFn)
	if err == nil {
		err = err2
	}
	return
}

func readCards(w *worker.Worker, c *anki.Collection) error {
	rowFn := func(row map[string]interface{}) {
		c.AddCardFromSqlite(row)
	}
	return readX(w, `
		SELECT c.*
		FROM cards c
		LEFT JOIN graves g ON g.type=0 AND g.oid=c.id
		WHERE g.oid IS NULL
	`, rowFn)
}

func readNotes(w *worker.Worker, c *anki.Collection) error {
	rowFn := func(row map[string]interface{}) {
		c.AddNoteFromSqlite(row)
	}
	return readX(w, `
		SELECT n.*
		FROM notes n
		LEFT JOIN graves g ON g.type=1 AND g.oid=n.id
		WHERE g.oid IS NULL
	`, rowFn)
}

func readRevlog(w *worker.Worker, c *anki.Collection) error {
	rowFn := func(row map[string]interface{}) {
		c.AddReviewFromSqlite(row)
	}
	return readX(w, `
		SELECT r.*
		FROM revlog r
		LEFT JOIN graves g ON g.type=1 AND g.oid=r.cid
		WHERE g.oid IS NULL
	`, rowFn)
}

func readGraves(w *worker.Worker, c *anki.Collection) error {
	rowFn := func(row map[string]interface{}) {
		c.DeleteDeck(int64(row["oid"].(float64)))
	}
	return readX(w, `
		SELECT oid
		FROM graves
		WHERE type=2
	`, rowFn)
}

func readX(w *worker.Worker, query string, fn rowFunc) error {
	w.Send(map[string]string{
		"action": "each",
		"sql":    query,
	})
	for {
		response, err := w.Receive()
		if err != nil {
			fmt.Printf("Error reading row: %s\n", err)
			return err
		}
		msg, ok := response.(map[string]interface{})
		if !ok {
			return errors.New("Unable to convert response to map\n")
		}
		if val, ok := msg["finished"].(bool); val && ok {
			return nil
		}
		if row, ok := msg["row"].(map[string]interface{}); ok {
			fn(row)
		}
	}
}
