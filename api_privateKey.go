package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"goji.io/pat"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/net/context"
)

/*
GET /api/user/:userID/privateKey - get list of user private keys
*/
func handleListUserPrivateKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	userID := pat.Param(ctx, "userID")
	if userID != UserFromContext(ctx).ID {
		http.Error(rw, "cannot list other user's private keys", http.StatusForbidden)
		return
	} else if m, err := StoreFromContext(ctx).GetPrivateKeys(userID); err != nil {
		rlog(ctx, "Could not query private keys: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else {
		ret := make([]keyResponse, 0, len(m))
		for keyID, armored := range m {
			ret = append(ret, keyResponse{
				KeyID:   keyID,
				UserID:  userID,
				Armored: armored,
			})
		}
		if err := RenderFromContext(ctx).JSON(rw, http.StatusOK, ret); err != nil {
			rlog(ctx, "Could not render JSON: ", err)
		}
	}
}

/*
GET /api/user/:userID/privateKey/:keyID - get information about a user's private key
*/
func handleGetUserPrivateKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	rUserID, keyID := pat.Param(ctx, "userID"), pat.Param(ctx, "keyID")
	if rUserID != UserFromContext(ctx).ID {
		http.Error(rw, "cannot get other user's private keys", http.StatusForbidden)
		return
	} else if userID, armored, err := StoreFromContext(ctx).GetPrivateKey(keyID); err == sql.ErrNoRows || rUserID != userID {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	} else if err != nil {
		rlog(ctx, "Could not query private keys: ", err)
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
POST /api/user/:userID/privateKey - add a private key
<body should be an armored GPG key>
*/
func handlePostUserPrivateKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	userID := pat.Param(ctx, "userID")
	if userID != UserFromContext(ctx).ID {
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
	} else if el[0].PrivateKey == nil {
		http.Error(rw, "missing private key", http.StatusBadRequest)
		return
	} else if keyID := el[0].PrivateKey.KeyIdString(); false {
	} else if err := StoreFromContext(ctx).AddPrivateKey(userID, keyID, b); err == ErrKeyAlreadyExists {
		http.Error(rw, "duplicate key", http.StatusConflict)
		return
	} else if err != nil {
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else {
		http.Redirect(rw, r, path.Join("/api/user", userID, "privateKey", keyID), http.StatusCreated)
		return
	}
}

/*
PUT /api/user/:userID/privateKey/:keyID - update a private key
<body should be an armored GPG key>
*/
func handlePutUserPrivateKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	userID, rKeyID := pat.Param(ctx, "userID"), pat.Param(ctx, "keyID")
	if userID != UserFromContext(ctx).ID {
		http.Error(rw, "cannot update other user's key", http.StatusForbidden)
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
	} else if el[0].PrivateKey == nil {
		http.Error(rw, "missing private key", http.StatusBadRequest)
		return
	} else if keyID := el[0].PrivateKey.KeyIdString(); keyID != rKeyID {
		http.Error(rw, "mismatching keys", http.StatusBadRequest)
		return
	} else if err := StoreFromContext(ctx).PutPrivateKey(userID, keyID, b); err == ErrUnknownKey {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	} else if err != nil {
		rlog(ctx, "Could not update private key: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
}

/*
DELETE /api/user/:userID/privateKey/:keyID - delete a private key
*/
func handleDeleteUserPrivateKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	userID, keyID := pat.Param(ctx, "userID"), pat.Param(ctx, "keyID")
	if userID != UserFromContext(ctx).ID {
		http.Error(rw, "cannot delete other user's key", http.StatusForbidden)
		return
	} else if err := StoreFromContext(ctx).RemovePrivateKey(userID, keyID); err == ErrUnknownKey {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	} else if err != nil {
		rlog(ctx, "Could not remove private key: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
}
