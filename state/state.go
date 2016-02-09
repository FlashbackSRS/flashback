package state

import (
	"time"
	"github.com/flimzy/go-pouchdb"
	"honnef.co/go/js/console"
)

type State struct {
	currentUser string
	lastRead    time.Time
	lastError	error
}

var state *State

func Read() {
	db := pouchdb.New("flashback")
	var newState = State{}
	if err := db.Get("_local/state", &newState, pouchdb.Options{}); err != nil {
		if pouchdb.IsNotExist(err) {
			// File not found, no problem
		} else {
			state.lastError = err
			console.Log(err)
			return
		}
	}
	state = &newState
	state.lastRead = time.Now()
}

func GetCurrentUser() string {
	return state.currentUser
}

func SetCurrentUser(user string) {
	state.currentUser = user
}

func GetError() error {
	err := state.lastError
	state.lastError = nil
	return err
}
