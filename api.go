package main

import (
	"io"
	"net/http"

	"golang.org/x/net/context"
)

func notImplemented(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	io.WriteString(rw, "not implemented")
}
