package fb

import (
	"testing"

	"github.com/flimzy/diff"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected *User
		err      string
	}{
		{
			name: "no name",
			err:  "name required",
		},
		{
			name: "valid",
			id:   "mjxwe",
			expected: &User{
				Name: "mjxwe",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewUser(test.id)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); err != nil {
				t.Error(d)
			}
		})
	}
}

func TestNilUser(t *testing.T) {
	u := NilUser()
	expected := &User{
		Name:     "aaaaaaaaabaabaaaaaaaaaaaaa",
		Created:  now(),
		Modified: now(),
	}
	u.Salt = "" // Non-deterministic
	if d := diff.Interface(expected, u); d != nil {
		t.Error(d)
	}
}

func TestUserMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected string
		err      string
	}{
		{
			name: "no name",
			user: &User{},
			err:  "name required",
		},
		{
			name: "null fields",
			user: &User{
				Name:     "mjxwe",
				Salt:     "salty",
				Password: "abc123",
				Created:  now(),
				Modified: now(),
			},
			expected: `{
			    "type":     "user",
				"_id":      "org.couchdb.user:mjxwe",
				"name":     "mjxwe",
				"roles":    [],
				"salt":     "salty",
				"password": "abc123",
				"created":  "2017-01-01T00:00:00Z",
				"modified": "2017-01-01T00:00:00Z"
            }`,
		},
		{
			name: "all fields",
			user: &User{
				Name:      "mjxwe",
				Roles:     []string{"foo", "bar"},
				Salt:      "salty",
				Password:  "abc123",
				FullName:  "Bob",
				Email:     "bob@bob.com",
				Created:   now(),
				Modified:  now(),
				LastLogin: now(),
			},
			expected: `{
				"type":     "user",
				"_id":       "org.couchdb.user:mjxwe",
				"name":      "mjxwe",
				"roles":     ["foo","bar"],
				"salt":      "salty",
				"password":  "abc123",
				"email":     "bob@bob.com",
				"fullname":  "Bob",
				"created":   "2017-01-01T00:00:00Z",
				"modified":  "2017-01-01T00:00:00Z",
				"lastLogin": "2017-01-01T00:00:00Z"
            }`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.user.MarshalJSON()
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.JSON([]byte(test.expected), result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestUserUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *User
		err      string
	}{
		{
			name:  "invalid json",
			input: "invalid json",
			err:   "failed to unmarshal user: invalid character 'i' looking for beginning of value",
		},
		{
			name:  "wrong doc id format",
			input: `{"_id":"foo"}`,
			err:   "id must have 'org.couchdb.user:' prefix",
		},
		{
			name:  "name-id mismatch",
			input: `{"_id":"org.couchdb.user:foo", "name":"bar"}`,
			err:   "user name and id must match",
		},
		{
			name:  "fails validation",
			input: `{"_id":"org.couchdb.user:foo", "name":"foo"}`,
			err:   "created time required",
		},
		{
			name: "null fields",
			input: `{
				"_id":      "org.couchdb.user:mjxwe",
				"name":     "mjxwe",
				"salt":     "salty",
				"password": "abc123",
				"created":  "2017-01-01T00:00:00Z",
				"modified": "2017-01-01T00:00:00Z"
            }`,
			expected: &User{
				Name:     "mjxwe",
				Salt:     "salty",
				Password: "abc123",
				Created:  now(),
				Modified: now(),
			},
		},
		{
			name: "all fields",
			input: `{
				"_id":       "org.couchdb.user:mjxwe",
				"name":      "mjxwe",
				"salt":      "salty",
				"password":  "abc123",
				"email":     "bob@bob.com",
				"fullname":  "Bob",
				"created":   "2017-01-01T00:00:00Z",
				"modified":  "2017-01-01T00:00:00Z",
				"lastLogin": "2017-01-01T00:00:00Z"
            }`,
			expected: &User{
				Name:      "mjxwe",
				Salt:      "salty",
				Password:  "abc123",
				Email:     "bob@bob.com",
				FullName:  "Bob",
				Created:   now(),
				Modified:  now(),
				LastLogin: now(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &User{}
			err := result.UnmarshalJSON([]byte(test.input))
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestUserValidate(t *testing.T) {
	tests := []validationTest{
		{
			name: "no name",
			v:    &User{},
			err:  "name required",
		},
		{
			name: "no created time",
			v:    &User{Name: "mzxw6"},
			err:  "created time required",
		},
		{
			name: "no modified time",
			v:    &User{Name: "mzxw6", Created: now()},
			err:  "modified time required",
		},
		{
			name: "no salt",
			v:    &User{Name: "mzxw6", Created: now(), Modified: now()},
			err:  "salt required",
		},
		{
			name: "valid",
			v:    &User{Name: "mjxwe", Created: now(), Modified: now(), Salt: "sea salt is the best"},
		},
	}
	testValidation(t, tests)
}

func TestGenerateSalt(t *testing.T) {
	salt := generateSalt()
	if len(salt) != 32 {
		t.Errorf("Generated salt is %d chars long, expected 32", len(salt))
	}
}

func TestGenerateUser(t *testing.T) {
	u := GenerateUser()
	if err := u.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestUserID(t *testing.T) {
	u, _ := NewUser("foo")
	expected := "org.couchdb.user:foo"
	if id := u.ID(); id != expected {
		t.Errorf("Unexpected id: %s", id)
	}
}
