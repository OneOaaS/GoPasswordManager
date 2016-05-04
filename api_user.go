package main

import (
	"net/http"

	"golang.org/x/net/context"
)

/*
GET /api/user/:id - get a user or the list of users
*/
func handleGetUser(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	notImplemented(ctx, rw, r)
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
	notImplemented(ctx, rw, r)
}

/*
PATCH /api/user/:id - modify a current user
{
  "name": "Full Name",
  "password": "plaintext password"
}
*/
func handlePatchUser(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	notImplemented(ctx, rw, r)
}

/*
DELETE /api/user/:id - delete a user
(no body)
*/
func handleDeleteUser(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	notImplemented(ctx, rw, r)
}
