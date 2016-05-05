package main

import "github.com/jmoiron/sqlx"

const initQuery = `
CREATE TABLE IF NOT EXISTS users (
	uid TEXT PRIMARY KEY NOT NULL,
	name TEXT,
	password BLOB,
	requiresPasswordReset BOOL NOT NULL
); -- potentially WITHOUT ROWID
INSERT OR IGNORE INTO users (uid, name, password, requiresPasswordReset) 
VALUES ("tolar2", "Jeffrey Tolar", "$2a$08$NrDJh5azlzGCvCaXYDI.O.0KLhKci7gmRC2D0yeBFi5q3xKU7ZTIq", 0); -- password = "tolar2"
`

func initDB(driver, dsn string) (DBStore, error) {
	if db, err := sqlx.Open(driver, dsn); err != nil {
		return DBStore{}, err
	} else if _, err := db.Exec(initQuery); err != nil {
		return DBStore{}, err
	} else {
		return DBStore{
			DB: db,
		}, nil
	}
}

type DBStore struct {
	DB *sqlx.DB
}

// GetUser retrieves the user from the store with the given id.
func (s DBStore) GetUser(userID string) (User, error) {
	var u User
	err := s.DB.Get(&u, `SELECT uid AS id, name, password, requiresPasswordReset FROM users WHERE uid = ?;`, userID)
	return u, err
}

// ListUsers retrieves metadata about all users in the store.
func (s DBStore) ListUsers() ([]UserMeta, error) {
	return nil, ErrNotImplemented
}

// PostUser adds a user to the store. The ID and Password fields of the
// user must not be empty.
func (s DBStore) PostUser(User) error {
	return ErrNotImplemented
}

// PutUser updates a user in the store. The ID field must not be blank.
func (s DBStore) PutUser(User) error {
	return ErrNotImplemented
}

// DeleteUser removes the user from the store with the given id.
func (s DBStore) DeleteUser(userID string) error {
	return ErrNotImplemented
}

// GetPublicKeys gets a list of public keys that belong to the user.
func (s DBStore) GetPublicKeys(userID string) ([]string, error) {
	return nil, ErrNotImplemented
}

// AddPublicKey associates a key with a user.
func (s DBStore) AddPublicKey(userID, keyID string) error {
	return ErrNotImplemented
}

// GetPrivateKeys gets a list of private key that belong to the user.
func (s DBStore) GetPrivateKeys(userID string) ([]string, error) {
	return nil, ErrNotImplemented
}

// AddPrivateKey associates a key with a user.
func (s DBStore) AddPrivateKey(userID, keyID string) error {
	return ErrNotImplemented
}
