package main

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrMissingID        = errors.New("missing id")
	ErrKeyAlreadyExists = errors.New("key already exists")
)

const initQuery = `
CREATE TABLE IF NOT EXISTS users (
	uid TEXT PRIMARY KEY NOT NULL,
	name TEXT,
	password BLOB,
	requiresPasswordReset BOOL NOT NULL
); -- potentially WITHOUT ROWID


CREATE TABLE IF NOT EXISTS public_keys (
	kid TEXT PRIMARY KEY NOT NULL,
	uid TEXT REFERENCES users(uid) ON DELETE SET DEFAULT, -- NULL if an external key
	armored BLOB NOT NULL
); -- potentially WITHOUT ROWID


CREATE TABLE IF NOT EXISTS private_keys (
	kid TEXT PRIMARY KEY NOT NULL,
	uid TEXT NOT NULL REFERENCES users(uid) ON DELETE CASCADE,
	armored BLOB NOT NULL
); -- potentially WITHOUT ROWID
`

func initDB(driver, dsn string) (DBStore, error) {
	if db, err := sqlx.Open(driver, dsn); err != nil {
		return DBStore{}, err
	} else if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
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

type dbKey struct {
	KeyID      string         `db:"kid"`
	UserID     sql.NullString `db:"uid"` // only can be NULL for public keys
	ArmoredKey []byte         `db:"armored"`
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
	if userID == "" {
		return nil, ErrMissingID
	}

	var keys []dbKey
	if err := s.DB.Select(&keys, `SELECT kid, armored FROM public_keys WHERE uid = ?;`, userID); err != nil {
		return nil, err
	}

	ret := make(map[string][]byte, len(keys))
	for _, key := range keys {
		ret[key.KeyID] = key.ArmoredKey
	}
	return ret, nil
}

func (s DBStore) GetPublicKeyIDs(userID string) ([]string, error) {
	if userID == "" {
		return nil, ErrMissingID
	}

	var keys []string
	if err := s.DB.Select(&keys, `SELECT kid FROM public_keys WHERE uid = ?;`, userID); err != nil {
		return nil, err
	}

	return keys, nil
}

func (s DBStore) AddPublicKey(userID, keyID string, armoredKey []byte) error {
	// use a transaction and check for existence so we can give better errors to the UI.
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	var key dbKey
	if err := tx.Get(&key, `SELECT kid, uid FROM public_keys WHERE kid = ?;`, keyID); err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	} else if err != sql.ErrNoRows && key.UserID.Valid {
		// we got a row, and it already has a user ID
		tx.Rollback()
		return ErrKeyAlreadyExists
	}

	if key.UserID.Valid {
		// allow adopting external keys
		if _, err := tx.Exec(`UPDATE public_keys SET uid = ?, armored = ? WHERE kid = ? AND uid == NULL;`, userID, armoredKey, keyID); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if _, err := tx.Exec(`INSERT INTO public_keys (kid, uid, armored) VALUES (?, ?, ?);`, keyID, userID, armoredKey); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s DBStore) RemovePublicKey(userID, keyID string) error {
	_, err := s.DB.Exec(`UPDATE public_keys
	                     SET uid = NULL
	                     WHERE kid = ? AND uid = ?;`,
		keyID, userID,
	)
	return err
}

func (s DBStore) AddExternalPublicKey(keyID string, armoredKey []byte) error {
	// use a transaction and check for existence so we can give better errors to the UI.
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	var existsID string
	if err := tx.Get(&existsID, `SELECT kid FROM public_keys WHERE kid = ?;`, keyID); err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	} else if err != sql.ErrNoRows {
		tx.Rollback()
		return ErrKeyAlreadyExists
	}

	if _, err := tx.Exec(`INSERT INTO public_keys (kid, armored) VALUES (?, ?);`, keyID, armoredKey); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s DBStore) GetUserForPublicKey(keyID string) (string, error) {
	var userID sql.NullString
	err := s.DB.Get(&userID, `SELECT uid FROM public_keys WHERE kid = ?;`, keyID)
	return userID.String, err
}

func (s DBStore) GetPrivateKeys(userID string) (map[string][]byte, error) {
	if userID == "" {
		return nil, ErrMissingID
	}

	var keys []dbKey
	if err := s.DB.Select(&keys, `SELECT kid, armored FROM private_keys WHERE uid = ?;`, userID); err != nil {
		return nil, err
	}

	ret := make(map[string][]byte, len(keys))
	for _, key := range keys {
		ret[key.KeyID] = key.ArmoredKey
	}
	return ret, nil
}

func (s DBStore) GetPrivateKeyIDs(userID string) ([]string, error) {
	if userID == "" {
		return nil, ErrMissingID
	}

	var keys []string
	if err := s.DB.Select(&keys, `SELECT kid FROM private_keys WHERE uid = ?;`, userID); err != nil {
		return nil, err
	}

	return keys, nil
}

func (s DBStore) AddPrivateKey(userID, keyID string, armoredKey []byte) error {
	// use a transaction and check for existence so we can give better errors to the UI.
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	var existsID string
	if err := tx.Get(&existsID, `SELECT kid FROM private_keys WHERE kid = ?;`, keyID); err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	} else if err != sql.ErrNoRows {
		tx.Rollback()
		return ErrKeyAlreadyExists
	}

	if _, err := tx.Exec(`INSERT INTO private_keys (kid, uid, armored) VALUES (?, ?, ?);`, keyID, userID, armoredKey); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s DBStore) RemovePrivateKey(userID, keyID string) error {
	_, err := s.DB.Exec(`DELETE FROM private_keys
	                     WHERE kid = ? AND uid = ?;`,
		keyID, userID,
	)
	return err
}
