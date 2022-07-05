package main

import (
	"net/http"

	"github.com/bmizerany/pat"
)

func (app *application) routes() http.Handler {
	mux := pat.New()
	dynamicmiddleware := func(fun http.Handler) http.Handler {
		return app.session.Enable(NoSurf(app.authenticate(fun)))
	}
	mux.Get("/", dynamicmiddleware(http.HandlerFunc(app.home)))
	mux.Get("/snippet/create", dynamicmiddleware(app.requireAuthUser(http.HandlerFunc(app.createSnippetForm))))
	mux.Post("/snippet/create", dynamicmiddleware(app.requireAuthUser(http.HandlerFunc(app.createSnippet))))
	mux.Get("/snippet/:id", dynamicmiddleware(http.HandlerFunc(app.showSnippet)))

	mux.Get("/user/signup", dynamicmiddleware(http.HandlerFunc(app.signupUserForm)))
	mux.Post("/user/signup", dynamicmiddleware(http.HandlerFunc(app.signupUser)))
	mux.Get("/user/login", dynamicmiddleware(http.HandlerFunc(app.loginUserForm)))
	mux.Post("/user/login", dynamicmiddleware(http.HandlerFunc(app.loginUser)))
	mux.Post("/user/logout", dynamicmiddleware(app.requireAuthUser(http.HandlerFunc(app.logoutUser))))

	fileServer := http.FileServer(http.Dir("../../ui/static/"))

	mux.Get("/static/", http.StripPrefix("/static", fileServer))
	return app.recoverPanic(app.logRequest(secureHeaders(mux)))
}
