package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/xyedo/snippetbox/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})
	fileServer := http.FileServer(http.FS(ui.Files))

	router.Handler(http.MethodGet, "/static/*filepath", fileServer)
	dynamicmiddleware := func(fun http.Handler) http.Handler {
		return app.sessionManager.LoadAndSave(NoSurf(app.authenticate(fun)))
	}
	router.Handler(http.MethodGet, "/", dynamicmiddleware(http.HandlerFunc(app.home)))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamicmiddleware(http.HandlerFunc(app.snippetView)))
	router.Handler(http.MethodGet, "/user/signup", dynamicmiddleware(http.HandlerFunc(app.userSignupView)))
	router.Handler(http.MethodPost, "/user/signup", dynamicmiddleware(http.HandlerFunc(app.userSignupPost)))
	router.Handler(http.MethodGet, "/user/login", dynamicmiddleware(http.HandlerFunc(app.userLoginView)))
	router.Handler(http.MethodPost, "/user/login", dynamicmiddleware(http.HandlerFunc(app.userLoginPost)))
	router.Handler(http.MethodGet, "/about", dynamicmiddleware(http.HandlerFunc(app.aboutView)))
	protected := func(fun http.Handler) http.Handler {
		return dynamicmiddleware(app.requireAuth(fun))
	}
	router.Handler(http.MethodGet, "/snippet/create", protected(http.HandlerFunc(app.snippetCreateView)))
	router.Handler(http.MethodPost, "/snippet/create", protected(http.HandlerFunc(app.createSnippetPost)))
	router.Handler(http.MethodGet, "/account/view", protected(http.HandlerFunc(app.accountView)))
	router.Handler(http.MethodPost, "/user/logout", protected(http.HandlerFunc(app.logoutUserPost)))
	router.Handler(http.MethodGet, "/ping", http.HandlerFunc(ping))

	return app.recoverPanic(app.logRequest(secureHeaders(router)))
}
