package main

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/elithrar/goji-logger"
	"github.com/goji/ctx-csrf"
	"github.com/gorilla/securecookie"
	"github.com/unrolled/render"

	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

func PanicHandler(next goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				debug.PrintStack()
				log.Print("Recovering from panic: ", r)
			}
		}()
		next.ServeHTTPC(ctx, rw, r)
	})
}

func main() {
	config := Config{
		CookieSecret: []byte("alskjdlkfaj zxcxvnafsflkasj rewoiiw"),
		CookieName:   "pass",
		Dev:          true,
	}

	us := StaticUserStore{}
	us.AddUser("tolar2", "Jeffrey Tolar", "tolar2")

	sc := securecookie.New(config.CookieSecret, nil)
	sc.SetSerializer(securecookie.JSONEncoder{})

	rootCtx := context.Background()
	rootCtx = ContextWithConfig(rootCtx, config)
	rootCtx = ContextWithUserStore(rootCtx, us)
	rootCtx = ContextWithSecureCookie(rootCtx, sc)
	rootCtx = ContextWithRender(rootCtx, render.New(render.Options{
		IsDevelopment: config.Dev,
		IndentJSON:    config.Dev,
	}))

	mux := goji.NewMux()
	apiMux := goji.SubMux()

	mux.UseC(PanicHandler)
	mux.UseC(logger.RequestID)
	mux.UseC(logger.Logger)

	if config.Dev {
		log.Print("[warning] Dev mode enabled: disabling CSRF protection")
	} else {
		mux.UseC(csrf.Protect(
			config.CookieSecret,
			csrf.RequestHeader("X-XSRF-TOKEN"),
			csrf.CookieName("XSRF-TOKEN"),
		))
	}

	apiMux.UseC(Auth)

	apiMux.HandleFuncC(pat.Get("/user"), handleGetUser)
	apiMux.HandleFuncC(pat.Get("/user/:id"), handleGetUser)
	apiMux.HandleFuncC(pat.Post("/user"), handlePostUser)
	apiMux.HandleFuncC(pat.Patch("/user/:id"), handlePatchUser)
	apiMux.HandleFuncC(pat.Delete("/user/:id"), handleDeleteUser)

	mux.HandleFuncC(pat.Post("/login"), PostLogin)
	mux.HandleC(pat.New("/api/*"), apiMux)

	mux.Handle(pat.New("/*"), http.FileServer(http.Dir("app/")))

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Print("Listening on ", addr)
	panic(http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTPC(rootCtx, w, r)
	})))
}
