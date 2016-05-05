package main

import (
	"database/sql"
	"net/http"
	"strings"

	"goji.io/pat"

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
func handleGetKey(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	keyID := pat.Param(ctx, "id")
	if userID, armored, err := StoreFromContext(ctx).GetPublicKey(keyID); err == sql.ErrNoRows {
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
GET /api/publicKey?ids=comma,separated,list,of,key,IDs - get information about multiple keys; unknown keys will be silently ignored.
*/
func handleGetKeys(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
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
