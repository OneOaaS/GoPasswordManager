package main

import (
	"path"
	"path/filepath"

	"golang.org/x/net/context"
)

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

type PassDirent struct {
	File bool
	Name string
}

// PassTx represents a read-only transaction on a PassStore.
type PassTx interface {
	// Type determines the whether path exists and if it's a file or directory.
	Type(path string) (exists bool, file bool)

	// List lists files in a directory. The Name of each PassDirent is the
	// basename of each file, not its full path. List will not be affect by Put
	// and Delete.
	List(path string) ([]PassDirent, error)

	// Get gets a specific file. It is an error to call Get on a path that has
	// been previously written to in the same transaction.
	Get(path string) ([]byte, error)

	// Recipients gets the list of recipients (key IDs) path (and possibly
	// subdirectories) should be encrypted to.
	Recipients(path string) ([]string, error)

	// GetAffectedFiles get a list of files that will be affected by a change
	// of recipients at path. These are the files that need to be reencrpted
	// and passed to SetRecipients.
	GetAffectedFiles(path string) ([]string, error)
}

// PassTxW represents a write transaction on a PassStore. No explicit actions
// are required to roll back an uncommitted transaction.
type PassTxW interface {
	// A writeable transaction is not required to include writes in future
	// reads: writing or deleting a file and then reading it later will likely
	// produce unexpected results.
	PassTx

	// Put puts a specific file. path must end with .gpg, and all parent
	// directories must not end with .gpg.
	Put(path string, contents []byte)

	// Delete removes a specific file (or directory).
	Delete(path string)

	// SetRecipients sets the list of recipients (key IDs) on path, which must
	// be a directory. In order for the transaction to succeed, all affected
	// files must be re-saved using Put or deleted with Delete.
	SetRecipients(path string, recipients []string)

	// Commit writes the changes to the repository to disk.
	Commit(userName, message string) error
}

type PassStore interface {
	Begin() (PassTx, error)
	BeginW() (PassTxW, error)
}

type PassWalker interface {
	Walk(root string, fn PassWalkFn) error
}

// PassStoreWalkFn is the function called by PassStore; the Name field of p is
// the full path of the file.
// Returning filepath.SkipDir will skip a directory. Any other errors are
// immediately passed back to the caller of Walk.
type PassWalkFn func(p PassDirent) error

func PassWalk(ps PassTx, root string, fn PassWalkFn) error {
	if walker, ok := ps.(PassWalker); ok {
		return walker.Walk(root, fn)
	} else {
		q := []PassDirent{{
			Name: "/",
			File: false,
		}}
		var d PassDirent
		for len(q) > 0 {
			d, q = q[0], q[1:]
			if err := fn(d); err == filepath.SkipDir {
				continue
			} else if err != nil {
				return err
			}
			if !d.File {
				// this is a directory, recurse
				if l, err := ps.List(d.Name); err != nil {
					return err
				} else {
					for _, f := range l {
						f.Name = path.Join(d.Name, f.Name)
						q = append(q, f)
					}
				}
			}
		}
		return nil
	}
}
