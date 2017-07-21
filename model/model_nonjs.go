// +build !js

package model

import (
	"context"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/memory" // Memory driver
)

func localConnection() (*kivik.Client, error) {
	return kivik.New(context.Background(), "memory", "local")
}

func remoteConnection(_ string) (*kivik.Client, error) {
	return kivik.New(context.Background(), "memory", "remote")
}
