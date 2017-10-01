package fb

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/flimzy/kivik"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

var nilUser = uuid.UUID([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

// User repressents a user of Flashback
type User struct {
	Rev       string    `json:"_rev,omitempty"`
	Name      string    `json:"name"`
	Roles     []string  `json:"roles"`
	Password  string    `json:"password"`
	Salt      string    `json:"salt"`
	FullName  string    `json:"fullname,omitempty"`
	Email     string    `json:"email,omitempty"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
	LastLogin time.Time `json:"lastLogin,omitempty"`
}

// GenerateUser creates a new user account, with a random ID.
func GenerateUser() *User {
	user, _ := NewUser(B32enc(uuid.NewUUID()))
	return user
}

// Validate validates that all of the data in the user appears valid and self
// consistent. A nil return value means no errors were detected.
func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name required")
	}
	if u.Created.IsZero() {
		return errors.New("created time required")
	}
	if u.Modified.IsZero() {
		return errors.New("modified time required")
	}
	return nil
}

// ID returns the document ID for the user record.
func (u *User) ID() string {
	return kivik.UserPrefix + u.Name
}

// NewUser returns a new User object, based on the provided UUID and username.
func NewUser(name string) (*User, error) {
	u := &User{
		Name:     name,
		Created:  now().UTC(),
		Modified: now().UTC(),
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return u, nil
}

// NilUser returns a special user, whose UUID bits are all set to zero, to be
// used as a placeholder when the actual user isn't known.
func NilUser() *User {
	u, e := NewUser(B32enc(nilUser))
	if e != nil {
		panic(e)
	}
	return u
}

type userAlias User

// MarshalJSON implements the json.Marshaler interface for the User type.
func (u *User) MarshalJSON() ([]byte, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	doc := struct {
		userAlias
		ID        string     `json:"_id"`
		Type      string     `json:"type"`
		Roles     []string   `json:"roles"`
		LastLogin *time.Time `json:"lastLogin,omitempty"`
	}{
		ID:        u.ID(),
		Type:      "user",
		userAlias: userAlias(*u),
	}
	if !u.LastLogin.IsZero() {
		doc.LastLogin = &u.LastLogin
	}
	if len(u.Roles) == 0 {
		// To ensure non-`null` value
		doc.Roles = []string{}
	} else {
		doc.Roles = u.Roles
	}
	return json.Marshal(doc)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the User type.
func (u *User) UnmarshalJSON(data []byte) error {
	doc := &struct {
		userAlias
		ID string `json:"_id"`
	}{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return errors.Wrap(err, "failed to unmarshal user")
	}
	if !strings.HasPrefix(doc.ID, kivik.UserPrefix) {
		return errors.New("id must have '" + kivik.UserPrefix + "' prefix")
	}
	if doc.Name != strings.TrimPrefix(doc.ID, kivik.UserPrefix) {
		return errors.New("user name and id must match")
	}
	*u = User(doc.userAlias)
	return u.Validate()
}
