package repo

import (
	"context"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb"

	"github.com/flimzy/flashback-server2/auth"
)

func Auth(provider, token, url string) error {
	client, err := kivik.New(context.TODO(), "couch", url)
	if err != nil {
		return err
	}
	auth := auth.NewOAuth2(provider, token)
	return client.Authenticate(context.TODO(), auth)
}
