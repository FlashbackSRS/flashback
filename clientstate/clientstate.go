package clientstate

import (
	"golang.org/x/net/context"
)

type State struct {
	CurrentUser		string
	Stack			[]context.Context
}

func New() *State {
	return &State{
		"",
		make([]context.Context, 0, 5),
	}
}
