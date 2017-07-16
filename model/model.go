package model

import (
	"context"
	"net/http"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb" // CouchDB driver
	"github.com/flimzy/kivik/driver/couchdb/chttp"
	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
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
		return errors.Wrap(err, "OAuth2 auth failed")
	}
	var response struct {
		Ctx struct {
			Name string `json:"name"`
		} `json:"userCtx"`
	}
	if _, err := r.chttp.DoJSON(ctx, http.MethodGet, "/_session", nil, &response); err != nil {
		return errors.Wrap(err, "failed to validate session")
	}
	if response.Ctx.Name == "" {
		return errors.New("no user set in session")
	}
	return r.setUser(ctx, response.Ctx.Name)
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

func (r *Repo) setUser(ctx context.Context, username string) error {
	u, err := r.fetchUser(ctx)
	if err != nil && kivik.StatusCode(err) != kivik.StatusNotFound {
		return err
	}
	u.ID = currentUserDoc
	u.Username = username
	if _, e := r.state.Put(ctx, currentUserDoc, u); e != nil {
		return errors.Wrap(e, "failed to store local state")
	}
	r.user = username
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

func (r *Repo) userDB(ctx context.Context) (*kivik.DB, error) {
	user := r.CurrentUser()
	if user == "" {
		return nil, errors.New("not logged in")
	}
	return r.local.DB(ctx, user)
}

func (r *Repo) bundleDB(ctx context.Context, bundle *fb.Bundle) (*kivik.DB, error) {
	if bundle == nil || !bundle.ID.Valid() {
		return nil, errors.New("invalid bundle")
	}
	if r.CurrentUser() == "" {
		return nil, errors.New("not logged in")
	}
	return r.local.DB(ctx, bundle.ID.String())
}
