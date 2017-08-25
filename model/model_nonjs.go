// +build !js

package model

import (
	"context"

	"github.com/flimzy/kivik"
	_ "github.com/go-kivik/memorydb" // Kivik Memory driver
)

func localConnection() (kivikClient, error) {
	c, err := kivik.New(context.Background(), "memory", "")
	if err != nil {
		return nil, err
	}
	return wrapClient(c), nil
}

func remoteConnection(_ string) (kivikClient, error) {
	c, err := kivik.New(context.Background(), "memory", "remote")
	if err != nil {
		return nil, err
	}
	return wrapClient(c), nil
}
