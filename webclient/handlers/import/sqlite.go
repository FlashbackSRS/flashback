package import_handler

import (
	"fmt"
	"errors"
	"github.com/gopherjs/gopherjs/js"
	"github.com/flimzy/web/worker"
	"github.com/flimzy/flashback/anki"
	"honnef.co/go/js/console"
)

func readSQLite(dbbuf []byte) error {
	fmt.Printf("Gonna pretend to read the SQLite data now...\n")

	w := worker.New("js/worker.sql.js")
	defer w.Terminate()
	w.Send(map[string]interface{}{
		"action": "open",
		"buffer": js.NewArrayBuffer(dbbuf),
	})
	response, err := w.Receive()
	if err != nil {
		return err
	}
	if msg, ok := response.(map[string]interface{}); ! ok {
		return errors.New("Received unexpected response from sqlite")
	} else if ready,ok := msg["ready"].(bool); !ok || !ready {
		return errors.New("Ready status not true")
	}
	
	collections, err := readCollections(w)
	if err != nil {
		return err
	}
	cards, err := readCards(w)
	if err != nil {
		return err
	}
	graves, err := readGraves(w)
	if err != nil {
		return err
	}
	notes, err := readNotes(w)
	if err != nil {
		return err
	}
	revlog, err := readRevlog(w)
	if err != nil {
		return err
	}
	console.Log(collections)
	console.Log(cards)
	console.Log(graves)
	console.Log(notes)
	console.Log(revlog)
	return nil
}

type rowFunc func(map[string]interface{})

func readCollections(w *worker.Worker) (*[]*anki.Collection, error) {
	var collections []*anki.Collection
	rowFn := func(row map[string]interface{}) {
		collections = append( collections, anki.SqliteToCollection( row ) )
	}
	err := readX(w, "SELECT * FROM col", rowFn)
	return &collections, err
}

func readCards(w *worker.Worker) (*[]*anki.Card, error) {
	var cards []*anki.Card
	rowFn := func(row map[string]interface{}) {
		cards = append( cards, anki.SqliteToCard( row ) )
	}
	err := readX(w, "SELECT * FROM cards", rowFn)
	return &cards, err
}

func readGraves(w *worker.Worker) (*[]*anki.Grave, error) {
	var graves []*anki.Grave
	rowFn := func(row map[string]interface{}) {
		graves = append( graves, anki.SqliteToGrave( row ) )
	}
	err := readX(w, "SELECT * FROM graves", rowFn)
	return &graves, err
}

func readNotes(w *worker.Worker) (*[]*anki.Note, error) {
	var notes []*anki.Note
	rowFn := func(row map[string]interface{}) {
		notes = append( notes, anki.SqliteToNote( row ) )
	}
	err := readX(w, "SELECT * FROM notes", rowFn)
	return &notes, err
}

func readRevlog(w *worker.Worker) (*[]*anki.Revlog, error) {
	var revlog []*anki.Revlog
	rowFn := func(row map[string]interface{}) {
		revlog = append( revlog, anki.SqliteToRevlog( row ) )
	}
	err := readX(w, "SELECT * FROM revlog", rowFn)
	return &revlog, err
}

func readX(w *worker.Worker, query string, fn rowFunc) error {
	w.Send(map[string]string{
		"action": "each",
		"sql": query,
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
			console.Log("Finished\n");
			return nil
		}
		if row, ok := msg["row"].(map[string]interface{}); ok {
			fn(row)
		}
	}
}
