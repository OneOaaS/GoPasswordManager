package main

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

// UserMeta represents short metadata about a user.
type UserMeta struct {
	// ID is the user's ID (what users use to log in)
	ID string `json:"id"`

	// Name is the user's full name.
	Name string `json:"name"`
}

// UserFull contains all public information about a user.
type UserFull struct {
	UserMeta

	// PublicKeys is the list of public keys owned by the user
	PublicKeys []string
}

// User contains all public and private information about a user.
type User struct {
	UserFull

	// Password is the hash (bcrypt) of the user's authentication password.
	Password []byte `json:"-"`

	// RequiresPasswordReset is true if the password needs to be reset by the
	// user.
	RequiresPasswordReset bool `json:"requiresPasswordReset"`

	// PrivateKeys is the list of private keys owned by the user
	PrivateKeys []string
}

// UserStore represents a place to store users.
type UserStore interface {
	// GetUser retrieves the user from the store with the given id.
	GetUser(userID string) (User, error)
	// ListUsers retrieves metadata about all users in the store.
	ListUsers() ([]UserMeta, error)
	// PostUser adds a user to the store. The ID and Password fields of the
	// user must not be empty.
	PostUser(User) error
	// PutUser updates a user in the store. The ID field must not be blank.
	PutUser(User) error
	// DeleteUser removes the user from the store with the given id.
	DeleteUser(userID string) error

	// GetPublicKeys gets a list of public keys that belong to the user.
	GetPublicKeys(userID string) ([]string, error)
	// AddPublicKey associates a key with a user.
	AddPublicKey(userID, keyID string) error

	// GetPrivateKeys gets a list of private key that belong to the user.
	GetPrivateKeys(userID string) ([]string, error)
	// AddPrivateKey associates a key with a user.
	AddPrivateKey(userID, keyID string) error
}

func GetUser(ctx context.Context, userID string) (User, error) {
	return UserStoreFromContext(ctx).GetUser(userID)
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
	var u User
	u.ID = id
	u.Name = name
	u.Password = pass
	s[id] = u
}

func (s StaticUserStore) ListUsers() ([]UserMeta, error) {
	l := make([]UserMeta, 0, len(s))
	for _, u := range s {
		l = append(l, u.UserMeta)
	}
	return l, nil
}
func (s StaticUserStore) PostUser(User) error            { return errors.New("not implemented") }
func (s StaticUserStore) PutUser(User) error             { return errors.New("not implemented") }
func (s StaticUserStore) DeleteUser(userID string) error { return errors.New("not implemented") }

func (s StaticUserStore) GetPublicKeys(userID string) ([]string, error) { return nil, nil }
func (s StaticUserStore) AddPublicKey(userID, keyID string) error {
	return errors.New("not implemented")
}
func (s StaticUserStore) GetPrivateKeys(userID string) ([]string, error) { return nil, nil }
func (s StaticUserStore) AddPrivateKey(userID, keyID string) error {
	return errors.New("not implemented")
}
