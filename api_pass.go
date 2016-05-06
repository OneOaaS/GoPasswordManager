package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path"
	"strings"

	"goji.io/pattern"

	"golang.org/x/net/context"
)

func apiPassName(s string) string {
	return strings.TrimSuffix(path.Base(s), ".gpg")
}

/*
GET /api/pass/* - get a password or a list of passwords
Reponse for files:
{
	"name": "base name of the file, minus the .gpg",
	"path": "full/path/to/file",
	"contents": "full file contents, base64 encoded",
	"recipients": ["key","ids","that","can","access"]
}

Reponse for directories:
{
	"children": [
		{
			"name": "name of the child",
			"path": "full path of the child",
			"type": "'dir' or 'file'"
		}
	],
	"recipients": ["key","ids","that","can","access","directory"]
}
*/
func handleGetPass(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	type responseFile struct {
		Name       string   `json:"name"`
		Path       string   `json:"path"`
		Contents   []byte   `json:"contents"`
		Recipients []string `json:"recipients"`
	}
	type responseDirEnt struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Type string `json:"type"`
	}
	type responseDir struct {
		Children   []responseDirEnt `json:"children"`
		Recipients []string         `json:"recipients"`
	}

	p := pattern.Path(ctx)
	ps := PassFromContext(ctx)
	var response interface{}
	if tx, err := ps.Begin(); err != nil {
		rlog(ctx, "Could not start transaction: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else if exists, isFile := tx.Type(p); !exists {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	} else if isFile {
		if contents, err := tx.Get(p); err != nil {
			rlog(ctx, "Could not get file contents: ", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		} else if recipients, err := getRecipients(bytes.NewReader(contents)); err != nil {
			rlog(ctx, "Could not get recipients: ", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		} else {
			response = responseFile{
				Name:       apiPassName(p),
				Path:       path.Clean(p),
				Contents:   contents,
				Recipients: recipients,
			}
		}
	} else {
		if recipients, err := tx.Recipients(p); err != nil {
			rlog(ctx, "Could not get recipients: ", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		} else if children, err := tx.List(p); err != nil {
			rlog(ctx, "Could not get directory listing: ", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		} else {
			rChildren := make([]responseDirEnt, 0, len(children))
			for _, c := range children {
				if c.File && !strings.HasSuffix(c.Name, ".gpg") {
					continue
				}
				var ch responseDirEnt
				ch.Name = apiPassName(c.Name)
				ch.Path = path.Join(p, c.Name)
				ch.Type = "file"
				if !c.File {
					ch.Type = "dir"
				}
				rChildren = append(rChildren, ch)
			}

			response = responseDir{
				Children:   rChildren,
				Recipients: recipients,
			}
		}
	}

	if err := RenderFromContext(ctx).JSON(rw, http.StatusOK, response); err != nil {
		rlog(ctx, "Could not render JSON: ", err)
	}
}

/*
POST /api/pass/* - save or create a password
{
	"contents": "full file contents, base64 encoded",
	"message": "commit message"
}
*/
func handlePostPass(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var req struct {
		Contents []byte `json:"contents"`
		Message  string `json:"message"`
	}
	p := pattern.Path(ctx)
	ps := PassFromContext(ctx)
	u := UserFromContext(ctx)
	tx, err := ps.BeginW()
	if err != nil {
		rlog(ctx, "Could not start transaction: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else if exists, isFile := tx.Type(p); exists && !isFile {
		http.Error(rw, "can't overwrite a directory", http.StatusBadRequest)
		return
	} else if uPubKeyIDs, err := StoreFromContext(ctx).GetPublicKeyIDs(u.ID); err != nil {
		rlog(ctx, "Could not get public keys: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else if recipients, err := tx.Recipients(p); err != nil {
		rlog(ctx, "Could not get recipients: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else if !containsAny(recipients, uPubKeyIDs) {
		http.Error(rw, "forbidden", http.StatusForbidden)
		return
	} else if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "invalid JSON", http.StatusBadRequest)
		return
	} else {
		tx.Put(p, req.Contents)
		if err := tx.Commit(u.Name, req.Message); err != nil {
			rlog(ctx, "Could not commit transaction: ", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

/*
DELETE /api/pass/* - delete a password
*/
func handleDeletePass(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	p := pattern.Path(ctx)
	ps := PassFromContext(ctx)
	u := UserFromContext(ctx)
	tx, err := ps.BeginW()
	if err != nil {
		rlog(ctx, "Could not start transaction: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else if exists, isFile := tx.Type(p); !exists {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	} else if !isFile {
		http.Error(rw, "can't delete a directory", http.StatusBadRequest)
		return
	} else if uPubKeyIDs, err := StoreFromContext(ctx).GetPublicKeyIDs(u.ID); err != nil {
		rlog(ctx, "Could not get public keys: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else if recipients, err := tx.Recipients(p); err != nil {
		rlog(ctx, "Could not get recipients: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else if !containsAny(recipients, uPubKeyIDs) {
		http.Error(rw, "forbidden", http.StatusForbidden)
		return
	} else {
		tx.Delete(p)
		if err := tx.Commit(u.Name, "Removed "+p+" from store."); err != nil {
			rlog(ctx, "Could not commit transaction: ", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func handleGetPerm(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	// p := pattern.Path(ctx)
	// ps := PassFromContext(ctx)
	// u := UserFromContext(ctx)
	// tx := ps.Begin()
}

func handlePostPerm(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	// p := pattern.Path(ctx)
	// ps := PassFromContext(ctx)
	// u := UserFromContext(ctx)
	// tx := ps.Begin()
}
