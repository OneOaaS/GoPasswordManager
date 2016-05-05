package main

import (
	"encoding/json"
	"net/http"
	"path"

	"goji.io/pat"
	"goji.io/pattern"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

/*
GET /api/user/:id - get a user or the list of users
*/
func handleGetUser(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	idi := ctx.Value(pattern.Variable("id"))
	if id, _ := idi.(string); id != "" {
		if u, err := UserStoreFromContext(ctx).GetUser(id); err != nil {
			http.Error(rw, "not found", http.StatusNotFound)
			return
		} else {
			RenderFromContext(ctx).JSON(rw, http.StatusOK, u)
		}
	} else { // if id == ""
		if us, err := UserStoreFromContext(ctx).ListUsers(); err != nil {
			rlog(ctx, "Could not list users: ", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		} else {
			RenderFromContext(ctx).JSON(rw, http.StatusOK, us)
		}
	}
}

/*
POST /api/user - create a user
{
  "id": "user_name",
  "name": "Full Name",
  "password": "plaintext password"
}
*/
func handlePostUser(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var ui struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	var u User

	us := UserStoreFromContext(ctx)

	if err := json.NewDecoder(r.Body).Decode(&ui); err != nil {
		rlog(ctx, "Could not decode JSON: ", err)
		http.Error(rw, "invalid JSON", http.StatusBadRequest)
		return
	} else if ui.ID == "" {
		http.Error(rw, "invalid id", http.StatusBadRequest)
		return
	} else if ui.Password == "" {
		http.Error(rw, "invalid password", http.StatusBadRequest)
		return
	} else if pass, err := bcrypt.GenerateFromPassword([]byte(ui.Password), bcrypt.DefaultCost); err != nil {
		rlog(ctx, "Could not hash password: ", err)
		http.Error(rw, "could not hash password", http.StatusInternalServerError)
		return
	} else {
		u.ID = ui.ID
		u.Name = ui.Name
		u.Password = pass
	}

	if err := us.PostUser(u); err != nil {
		rlog(ctx, "Could not create user: ", err)
		http.Error(rw, "could not create user", http.StatusInternalServerError)
		return
	}

	http.Redirect(rw, r, path.Join("user", ui.ID), http.StatusCreated)
}

/*
PATCH /api/user/:id - modify a user (can only modify the current user)
{
  "name": "Full Name",
  "oldPassword": "old plaintext password",
  "password": "plaintext password"
}

oldPassword is only needed if password is being set.
*/
func handlePatchUser(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var ui struct {
		Name        *string `json:"name"`
		OldPassword *string `json:"oldPassword"`
		Password    *string `json:"password"`
	}

	us := UserStoreFromContext(ctx)
	u := UserFromContext(ctx)
	rid := pat.Param(ctx, "id")

	if rid != u.ID {
		http.Error(rw, "cannot modify user", http.StatusForbidden)
		return
	} else if err := json.NewDecoder(r.Body).Decode(&ui); err != nil {
		rlog(ctx, "Could not decode JSON: ", err)
		http.Error(rw, "invalid JSON", http.StatusBadRequest)
		return
	}

	if ui.Name != nil {
		u.Name = *ui.Name
	}
	if ui.Password != nil {
		if ui.OldPassword == nil {
			http.Error(rw, "missing oldPassword", http.StatusBadRequest)
			return
		} else if err := bcrypt.CompareHashAndPassword(u.Password, []byte(*ui.OldPassword)); err != nil {
			http.Error(rw, "invalid password", http.StatusForbidden)
			return
		} else if pass, err := bcrypt.GenerateFromPassword([]byte(*ui.Password), bcrypt.DefaultCost); err != nil {
			rlog(ctx, "Could not hash password: ", err)
			http.Error(rw, "could not hash password", http.StatusInternalServerError)
			return
		} else {
			u.Password = pass
			u.RequiresPasswordReset = false
		}
	}

	if err := us.PutUser(u); err != nil {
		rlog(ctx, "Could not update user: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
}

/*
DELETE /api/user/:id - delete a user
{
	"passsword": "user's password",
}
*/
func handleDeleteUser(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var ui struct {
		Password string `json:"password"`
	}

	us := UserStoreFromContext(ctx)
	u := UserFromContext(ctx)
	rid := pat.Param(ctx, "id")

	if rid != u.ID {
		http.Error(rw, "cannot delete user", http.StatusForbidden)
		return
	} else if err := json.NewDecoder(r.Body).Decode(&ui); err != nil {
		rlog(ctx, "Could not decode JSON: ", err)
		http.Error(rw, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(ui.Password)); err != nil {
		http.Error(rw, "invalid password", http.StatusForbidden)
		return
	}

	if err := us.DeleteUser(u.ID); err != nil {
		rlog(ctx, "Could not delete user: ", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
}
