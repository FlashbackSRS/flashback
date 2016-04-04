package user

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pborman/uuid"
)

const USERPREFIX = "org.couchdb.user:"

// An identity/auth provider account, such as Facebook or Google
type AuthAccount struct {
	Provider    string `json:"$Provider"`
	FullName    string `json:"$FullName,omitempty"`
	Email       string `json:"$Email,omitempty"`
	UserID      string `json:"$UserID"`
	AccessToken string `json:"$AccessToken"`
}

type userDoc struct {
	ID       string   `json:"_id"`
	Rev      string   `json:"_rev,omitempty"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Roles    []string `json:"roles"`
	Password string   `json:"password"`
	Salt     string   `json:"salt"`
	// UserType serves as a convenient way to distinguish between Flashback and CouchDB (aka admin) users
	UserType     string         `json:"$UserType"`
	FullName     *string        `json:"$FullName,omitempty"`
	Email        *string        `json:"$Email,omitempty"`
	AuthAccounts []*AuthAccount `json:"$AuthAccounts"`
}

type User struct {
	uuid       uuid.UUID
	doc        userDoc
	FullName   string
	Email      string
	dbInitDone <-chan struct{}
}

func (u *User) newUserDoc() {
	u.doc = userDoc{
		Type:     "user",
		Roles:    []string{},
		UserType: "flashback",
		FullName: &u.FullName,
		Email:    &u.Email,
	}
	if u.uuid != nil {
		u.doc.Name = u.UUID().String()
		u.doc.ID = USERPREFIX + u.doc.Name
	}
}

// Create creates a new user with a random UUID
func Create() *User {
	u := &User{
		uuid: uuid.NewRandom(),
	}
	u.newUserDoc()
	return u
}

func (u *User) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &u.doc); err != nil {
		return err
	}
	if id, err := base64.URLEncoding.DecodeString(u.doc.Name); err != nil {
		return err
	} else {
		u.uuid = id
	}
	return nil
}

func (u *User) MarshalJSON() ([]byte, error) {
	x, err := json.Marshal(u.doc)
	fmt.Printf("%s\n", x)
	return x, err
}

func (u *User) AddAuthAccount(a *AuthAccount) error {
	for _, acct := range u.doc.AuthAccounts {
		if acct.Provider == a.Provider && acct.UserID == a.UserID {
			return fmt.Errorf("Provider %s account %s is already associated with this user", a.Provider, a.UserID)
		}
	}
	u.doc.AuthAccounts = append(u.doc.AuthAccounts, a)
	return nil
}

func (u *User) ID() string {
	return u.doc.Name
}

func (u *User) UUID() uuid.UUID {
	return u.uuid
}

func (u *User) DocID() string {
	return u.doc.ID
}

func (u *User) DBName() string {
	fmt.Printf("%v\n", u)
	return "user-" + u.ID()
}

func (u *User) MasterReviewsDBName() string {
	return "reviews-" + u.ID()
}

func (u *User) Salt() string {
	return u.doc.Salt
}
