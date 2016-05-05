package main

import "golang.org/x/net/context"

// UserMeta represents short metadata about a user.
type UserMeta struct {
	// ID is the user's ID (what users use to log in)
	ID string `json:"id" db:"id"`

	// Name is the user's full name.
	Name string `json:"name" db:"name"`
}

// UserFull contains all public information about a user.
type UserFull struct {
	UserMeta

	// PublicKeys are the public keys owned by the user.
	PublicKeys map[string][]byte `json:"publicKeys,omitempty"`
}

// User contains all public and private information about a user.
type User struct {
	UserFull

	// Password is the hash (bcrypt) of the user's authentication password.
	Password []byte `json:"-" db:"password"`

	// RequiresPasswordReset is true if the password needs to be reset by the
	// user.
	RequiresPasswordReset bool `json:"requiresPasswordReset,omitempty" db:"requiresPasswordReset"`

	// PrivateKeys are the private keys owned by the user.
	PrivateKeys map[string][]byte `json:"privateKeys,omitempty"`
}

// Store represents a place to store users.
type Store interface {
	// GetUser retrieves the user from the store with the given id.
	GetUser(userID string) (User, error)
	// ListUsers retrieves metadata about all users in the store.
	ListUsers() ([]UserMeta, error)
	// PostUser adds a user to the store. The ID and Password fields of the
	// user must not be empty.
	PostUser(User) error
	// PutUser updates a user's metadata in the store. The ID and password
	// fields must not be blank.
	PutUser(User) error
	// DeleteUser removes the user from the store with the given id.
	DeleteUser(userID string) error

	// GetPublicKeys gets the public keys that belong to a user.
	GetPublicKeys(userID string) (map[string][]byte, error)
	// GetPublicKeyIDs gets a list of public key IDs that belong to the user.
	GetPublicKeyIDs(userID string) ([]string, error)
	// AddPublicKey associates a key with a user.
	AddPublicKey(userID, keyID string, armoredKey []byte) error
	// RemovePublicKey removes a key from a user. The key itself is not removed
	// from the store, however.
	RemovePublicKey(userID, keyID string) error

	// AddExternalPublicKey adds the key to the store, but doesn't associate it
	// with any user.
	AddExternalPublicKey(keyID string, armoredKey []byte) error

	// GetUserForPublicKey finds the user id owning the given key. If the key
	// does not belong to any users, GetUserForPublicKey returns the empty
	// string.
	GetUserForPublicKey(keyID string) (user string, err error)

	// GetPublicKey gets information about a public key. If the key is an
	// external public key (i.e., it doesn't belong to any user), the first
	// return value will be the empty string.
	GetPublicKey(keyID string) (user string, armoredKey []byte, err error)

	// GetPrivateKeys gets the private keys that belong to the user.
	GetPrivateKeys(userID string) (map[string][]byte, error)
	// GetPrivateKeyIDs gets a list of private key IDs that belong to the user.
	GetPrivateKeyIDs(userID string) ([]string, error)
	// AddPrivateKey associates a key with a user.
	AddPrivateKey(userID, keyID string, armoredKey []byte) error
	// PutPrivateKey updates a user's private key.
	PutPrivateKey(userID, keyID string, armoredKey []byte) error
	// RemovePrivateKey removes a key from a user. The key itself should also
	// be removed from the store.
	RemovePrivateKey(userID, keyID string) error

	// GetPrivateKey gets information about a private key.
	GetPrivateKey(keyID string) (user string, armoredKey []byte, err error)
}

func GetUser(ctx context.Context, userID string) (User, error) {
	return StoreFromContext(ctx).GetUser(userID)
}
