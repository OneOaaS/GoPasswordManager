package main

import (
	"net/http"

	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

func main() {
	config := Config{
		CookieSecret: []byte("alskjdlkfaj zxcxvnafsflkasj rewoiiw"),
		CookieName:   "pass",
	}

	us := StaticUserStore{}
	us.AddUser("tolar2", "Jeffrey Tolar", "tolar2")

	rootCtx := context.Background()
	rootCtx = ContextWithConfig(rootCtx, config)
	rootCtx = ContextWithUserStore(rootCtx, us)

	mux := goji.NewMux()
	apiMux := goji.SubMux()

	apiMux.UseC(Auth)

	mux.HandleFuncC(pat.Post("/login"), PostLogin)
	mux.HandleC(pat.New("/api"), apiMux)

	mux.Handle(pat.New("/*"), http.FileServer(http.Dir("app/")))

	// TODO: make this configurable
	panic(http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTPC(rootCtx, w, r)
	})))
}
