package main

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

// A User represents a user.
type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Password   []byte `json:"-"`
	PublicKey  []byte `json:"publicKey"`
	PrivateKey []byte `json:"privateKey"`
}

// UserStore represents a place to store users.
type UserStore interface {
	GetUser(id string) (User, error)

	// TODO(tolar2): add methods for adding, modifying users
}

func GetUser(ctx context.Context, id string) (User, error) {
	return UserStoreFromContext(ctx).GetUser(id)
}

// A StaticUserStore stores a list of predefined (static) users.
type StaticUserStore map[string]User

func (s StaticUserStore) GetUser(id string) (User, error) {
	if u, ok := s[id]; ok {
		return u, nil
	}
	return User{}, errors.New("invalid user")
}

func (s StaticUserStore) AddUser(id, name, password string) {
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	s[id] = User{
		ID:       id,
		Name:     name,
		Password: pass,
	}
}
