package main

import (
	"net/http"

	"github.com/goji/ctx-csrf"
	"github.com/gorilla/securecookie"

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

	sc := securecookie.New(config.CookieSecret, nil)
	sc.SetSerializer(securecookie.JSONEncoder{})

	rootCtx := context.Background()
	rootCtx = ContextWithConfig(rootCtx, config)
	rootCtx = ContextWithUserStore(rootCtx, us)
	rootCtx = ContextWithSecureCookie(rootCtx, sc)

	mux := goji.NewMux()
	apiMux := goji.SubMux()

	mux.UseC(csrf.Protect(
		config.CookieSecret,
		csrf.RequestHeader("X-XSRF-TOKEN"),
		csrf.CookieName("XSRF-TOKEN"),
	))

	apiMux.UseC(Auth)

	apiMux.HandleFuncC(pat.Get("/user"), handleGetUser)
	apiMux.HandleFuncC(pat.Get("/user/:id"), handleGetUser)
	apiMux.HandleFuncC(pat.Post("/user"), handlePostUser)
	apiMux.HandleFuncC(pat.Patch("/user/:id"), handlePatchUser)
	apiMux.HandleFuncC(pat.Delete("/user/:id"), handleDeleteUser)

	mux.HandleFuncC(pat.Post("/login"), PostLogin)
	mux.HandleC(pat.New("/api/*"), apiMux)

	mux.Handle(pat.New("/*"), http.FileServer(http.Dir("app/")))

	// TODO: make this configurable
	panic(http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTPC(rootCtx, w, r)
	})))
}
