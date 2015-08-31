package user

import ()

type User struct {
	name       string
	attributes map[string]string
}

func New(name string) *User {
	return &User{
		name,
		make(map[string]string),
	}
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Set(key, value string) {
	u.attributes[key] = value
}

func (u *User) Get(key string) string {
	return u.attributes[key]
}
