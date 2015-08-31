package clientstate

import (
	"github.com/flimzy/flashback/user"
	"golang.org/x/net/context"
)

type State struct {
	User  *user.User
	Stack []context.Context
}

func New() *State {
	return &State{
		nil,
		make([]context.Context, 0, 5),
	}
}
