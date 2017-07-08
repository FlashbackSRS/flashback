package model

import (
	"context"
	"log"
	"net/http"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb" // CouchDB driver
	"github.com/flimzy/kivik/driver/couchdb/chttp"
	"github.com/pkg/errors"

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
		return errors.Wrap(err, "failed to validate session")
	}
	r.user = response.Ctx.Name
	return r.storeUser(ctx)
}

type user struct {
	ID       string `json:"_id"`
	Rev      string `json:"_rev,omitempty"`
	Username string `json:"username"`
}

func (r *Repo) fetchUser(ctx context.Context) (user, error) {
	var u user
	row, err := r.state.Get(ctx, currentUserDoc)
	if err != nil {
		return user{}, err
	}
	if e := row.ScanDoc(&u); e != nil {
		return user{}, e
	}
	return u, nil
}

func (r *Repo) storeUser(ctx context.Context) error {
	u, err := r.fetchUser(ctx)
	if err != nil && kivik.StatusCode(err) != kivik.StatusNotFound {
		return err
	}
	u.ID = currentUserDoc
	u.Username = r.user
	if _, e := r.state.Put(ctx, currentUserDoc, u); e != nil {
		return errors.Wrap(e, "failed to store local state")
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
	if r.user == "" {
		u, err := r.fetchUser(context.TODO())
		if err != nil && kivik.StatusCode(err) != kivik.StatusNotFound {
			log.Printf("Error fetching current user: %s", err)
		}
		r.user = u.Username
	}
	return r.user
}
