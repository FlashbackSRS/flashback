package state

import "time"

type State struct {
	currentUser string
	lastRead    time.Time
	lastError   error
}

var state *State

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
