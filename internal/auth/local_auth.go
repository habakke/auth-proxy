package auth

import (
	"github.com/habakke/auth-proxy/internal/auth/providers"
)

type LocalUser struct {
	Username string
	Password string
}

func (u LocalUser) GetID() string {
	return ""
}

func (u LocalUser) GetUsername() string {
	return u.Username
}

func (u LocalUser) GetName() string {
	return ""
}

func (u LocalUser) GetEmail() string {
	return ""
}

type LocalAuth struct {
	users map[string]*LocalUser
}

func NewAuthLocal() *LocalAuth {
	return &LocalAuth{
		users: make(map[string]*LocalUser),
	}
}

func (a *LocalAuth) AddUser(user *LocalUser) {
	a.users[user.Username] = user
}

func (a *LocalAuth) RemoveUser(username string) {
	delete(a.users, username)
}

func (a *LocalAuth) Authenticate(username string, password string) (providers.User, bool) {
	if u, ok := a.users[username]; ok {
		if u.Password == password {
			return u, true
		}
	}
	return nil, false
}
