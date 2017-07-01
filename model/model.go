package model

import (
	"context"
	"net/http"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb" // CouchDB driver
	"github.com/flimzy/kivik/driver/couchdb/chttp"

	"github.com/flimzy/flashback-server2/auth"
)

// stateDB is a local database, which is never synced, for storing of persistent
// state.
const stateDB = "flashback"

// currentUserDoc is the doc ID for storing the current user state.
const currentUserDoc = "_local/currentUser"

// Repo represents an instance of the Couch/Pouch model.
type Repo struct {
	chttp  *chttp.Client
	remote *kivik.Client
	local  *kivik.Client
	state  *kivik.DB
	user   string
}

// New returns a new Repo instance, pointing to the specified remote server.
func New(ctx context.Context, remoteURL string) (*Repo, error) {
	remoteClient, err := kivik.New(ctx, "couch", remoteURL)
	if err != nil {
		return nil, err
	}
	httpClient, err := chttp.New(ctx, remoteURL)
	if err != nil {
		return nil, err
	}
	localClient, err := localConnection()
	if err != nil {
		return nil, err
	}
	if e := localClient.CreateDB(ctx, stateDB); e != nil && kivik.StatusCode(e) != kivik.StatusConflict {
		return nil, e
	}
	stateDB, err := localClient.DB(ctx, stateDB)
	if err != nil {
		return nil, err
	}
	return &Repo{
		chttp:  httpClient,
		remote: remoteClient,
		local:  localClient,
		state:  stateDB,
	}, nil
}

// Auth attempts to authenticate with the provided OAuth2 provider/token pair.
func (r *Repo) Auth(ctx context.Context, provider, token string) error {
	auth := auth.NewOAuth2(provider, token)
	if err := r.chttp.Auth(ctx, auth); err != nil {
		return err
	}
	var response struct {
		Ctx struct {
			Name string `json:"name"`
		} `json:"userCtx"`
	}
	if _, err := r.chttp.DoJSON(ctx, http.MethodGet, "/_session", nil, &response); err != nil {
		return err
	}
	r.user = response.Ctx.Name
	if _, e := r.state.Put(ctx, currentUserDoc, map[string]string{"username": r.user}); e != nil {
		return e
	}
	return nil
}

// Logout clears the auth session.
func (r *Repo) Logout(ctx context.Context) error {
	if err := r.chttp.Logout(ctx); err != nil {
		return err
	}
	r.user = ""
	if _, e := r.state.Delete(ctx, currentUserDoc, ""); e != nil {
		return e
	}
	return nil
}

// CurrentUser returns the currently registered user.
func (r *Repo) CurrentUser() string {
	return r.user
}
