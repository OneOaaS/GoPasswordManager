package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"goji.io/pat"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/net/context"
)

type keyResponse struct {
	KeyID   string `json:"key"`
	UserID  string `json:"user"`
	Armored []byte `json:"armored"`
}

/*
GET /api/publicKey/:id - get information about a public key.
*/
func handleGetPublicKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	keyID := pat.Param(ctx, "id")
	if userID, armored, err := StoreFromContext(ctx).GetPublicKey(keyID); err == sql.ErrNoRows {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	} else if err != nil {
		rlog(ctx, "Could not query public keys: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else {
		res := keyResponse{
			KeyID:   keyID,
			UserID:  userID,
			Armored: armored,
		}
		if err := RenderFromContext(ctx).JSON(rw, http.StatusOK, res); err != nil {
			rlog(ctx, "Could not render JSON: ", err)
		}
	}
}

/*
GET /api/publicKey?ids=comma,separated,list,of,key,IDs - get information about multiple keys; unknown keys will be silently ignored.
*/
func handleGetPublicKeys(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	ids := strings.FieldsFunc(r.URL.Query().Get("ids"), func(r rune) bool {
		return r == ',' || r == ' '
	})
	ret := make(map[string]keyResponse)
	store := StoreFromContext(ctx)
	for _, keyID := range ids {
		if userID, armored, err := store.GetPublicKey(keyID); err == sql.ErrNoRows {
			continue
		} else if err != nil {
			rlog(ctx, "Could not query private keys: ", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		} else {
			ret[keyID] = keyResponse{
				KeyID:   keyID,
				UserID:  userID,
				Armored: armored,
			}
		}
	}
	if err := RenderFromContext(ctx).JSON(rw, http.StatusOK, ret); err != nil {
		rlog(ctx, "Could not render JSON: ", err)
	}
}

/*
GET /api/user/:userID/publicKey - get list of user keys
*/
func handleListUserPublicKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	userID := pat.Param(ctx, "userID")
	if m, err := StoreFromContext(ctx).GetPublicKeys(userID); err != nil {
		rlog(ctx, "Could not query public keys: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else {
		ret := make(map[string]keyResponse, len(m))
		for keyID, armored := range m {
			ret[keyID] = keyResponse{
				KeyID:   keyID,
				UserID:  userID,
				Armored: armored,
			}
		}
		if err := RenderFromContext(ctx).JSON(rw, http.StatusOK, ret); err != nil {
			rlog(ctx, "Could not render JSON: ", err)
		}
	}
}

/*
GET /api/user/:userID/publicKey/:keyID - get information about a public key
*/
func handleGetUserPublicKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	rUserID, keyID := pat.Param(ctx, "userID"), pat.Param(ctx, "keyID")
	if userID, armored, err := StoreFromContext(ctx).GetPublicKey(keyID); err == sql.ErrNoRows || rUserID != userID {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	} else if err != nil {
		rlog(ctx, "Could not query public keys: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else {
		res := keyResponse{
			KeyID:   keyID,
			UserID:  userID,
			Armored: armored,
		}
		if err := RenderFromContext(ctx).JSON(rw, http.StatusOK, res); err != nil {
			rlog(ctx, "Could not render JSON: ", err)
		}
	}
}

/*
POST /api/user/:userID/publicKey - add a public key
<body should be an armored GPG key>
*/
func handlePostUserPublicKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	rUserID := pat.Param(ctx, "userID")
	if rUserID != UserFromContext(ctx).ID {
		http.Error(rw, "cannot add key to other user", http.StatusForbidden)
		return
	} else if b, err := ioutil.ReadAll(r.Body); err != nil {
		rlog(ctx, "Could not read entire request body: ", err)
		http.Error(rw, "bad request", http.StatusBadRequest)
		return
	} else if el, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(b)); err != nil {
		http.Error(rw, fmt.Sprintf("malformed key: %v", err), http.StatusBadRequest)
		return
	} else if len(el) == 0 {
		http.Error(rw, "no keys found", http.StatusBadRequest)
		return
	} else if len(el) != 1 {
		http.Error(rw, "multiple keys found", http.StatusBadRequest)
		return
	} else if el[0].PrimaryKey == nil {
		http.Error(rw, "missing public (signing) key", http.StatusBadRequest)
		return
	} else if keyID := fmt.Sprintf("%X", el[0].PrimaryKey.Fingerprint); false {
	} else if userID := UserFromContext(ctx).ID; false {
	} else if err := StoreFromContext(ctx).AddPublicKey(userID, keyID, b); err == ErrKeyAlreadyExists {
		http.Error(rw, "duplicate key", http.StatusConflict)
		return
	} else if err != nil {
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else {
		http.Redirect(rw, r, path.Join("/api/user", userID, "publicKey", keyID), http.StatusCreated)
		return
	}
}

/*
DELETE /api/user/:userID/publicKey/:keyID - delete a public key
*/
func handleDeleteUserPublicKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	userID, keyID := pat.Param(ctx, "userID"), pat.Param(ctx, "keyID")
	if userID != UserFromContext(ctx).ID {
		http.Error(rw, "cannot delete other user's key", http.StatusForbidden)
		return
	} else if err := StoreFromContext(ctx).RemovePublicKey(userID, keyID); err != nil {
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
}
