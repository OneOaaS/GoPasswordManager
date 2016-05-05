package main

import (
	"errors"
	"io"
	"net/http"

	"golang.org/x/net/context"
)

var ErrNotImplemented = errors.New("not implemented")

func notImplemented(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	io.WriteString(rw, "not implemented")
}
