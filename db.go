package main

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrMissingID = errors.New("missing id")

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

func (s DBStore) GetUser(userID string) (User, error) {
	var u User
	err := s.DB.Get(&u, `SELECT uid AS id, name, password, requiresPasswordReset FROM users WHERE uid = ?;`, userID)
	return u, err
}

func (s DBStore) ListUsers() ([]UserMeta, error) {
	var us []UserMeta
	err := s.DB.Select(&us, `SELECT uid AS id, name FROM users;`)
	return us, err
}

func (s DBStore) PostUser(u User) error {
	if u.ID == "" {
		return ErrMissingID
	} else if len(u.Password) == 0 {
		return errors.New("missing password")
	}
	_, err := s.DB.Exec(`INSERT INTO users (uid, name, password, requiresPasswordReset)
	                     VALUES (?, ?, ?, ?);`,
		u.ID, u.Name, u.Password, u.RequiresPasswordReset,
	)
	return err
}

func (s DBStore) PutUser(u User) error {
	if u.ID == "" {
		return ErrMissingID
	} else if len(u.Password) == 0 {
		return errors.New("missing password")
	}
	_, err := s.DB.Exec(`UPDATE users SET
	                       name = ?,
	                       password = ?,
	                       requiresPasswordReset = ?
	                     WHERE uid = ?;`,
		u.Name, u.Password, u.RequiresPasswordReset, u.ID,
	)
	return err
}

func (s DBStore) DeleteUser(userID string) error {
	if userID == "" {
		return ErrMissingID
	}
	_, err := s.DB.Exec(`DELETE FROM users
	                     WHERE uid = ?;`,
		userID,
	)
	return err
}

func (s DBStore) GetPublicKeys(userID string) (map[string][]byte, error) {
	return nil, ErrNotImplemented
}

func (s DBStore) GetPublicKeyIDs(userID string) ([]string, error) {
	return nil, ErrNotImplemented
}

func (s DBStore) AddPublicKey(userID, keyID string, armoredKey []byte) error {
	return ErrNotImplemented
}

func (s DBStore) RemovePublicKey(userID, keyID string) error {
	return ErrNotImplemented
}

func (s DBStore) AddExternalPublicKey(keyID string, armoredKey []byte) error {
	return ErrNotImplemented
}

func (s DBStore) GetUserForPublicKey(keyID string) (string, error) {
	return "", ErrNotImplemented
}

func (s DBStore) GetPrivateKeys(userID string) (map[string][]byte, error) {
	return nil, ErrNotImplemented
}

func (s DBStore) GetPrivateKeyIDs(userID string) ([]string, error) {
	return nil, ErrNotImplemented
}

func (s DBStore) AddPrivateKey(userID, keyID string, armoredKey []byte) error {
	return ErrNotImplemented
}

func (s DBStore) RemovePrivateKey(userID, keyID string) error {
	return ErrNotImplemented
}
