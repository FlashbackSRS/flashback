package import_handler

import (
	"fmt"
	"errors"
	"github.com/gopherjs/gopherjs/js"
	"github.com/flimzy/web/worker"
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
	
	w.Send(map[string]interface{}{
		"action": "each",
		"sql": "SELECT * FROM col",
	})
	for {
		response, err := w.Receive()
		if err != nil {
			fmt.Printf("Error reading row: %s\n", err)
			return err
		}
		msg, ok := response.(map[string]interface{})
		if ! ok {
			return errors.New("Unable to convert response to map\n")
		}
		console.Log(msg)
	}
	return nil
}
