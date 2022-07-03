package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/xyedo/snippetbox/pkg/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, r, "home.page.tmpl", &templateData{Snippets: s})

}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	snippet, err := app.snippets.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, r, "show.page.tmpl", &templateData{Snippet: snippet})
}
func (app *application) createSnippetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		createSnippet(app, w, r)
	case "GET":
		createSnippetForm(app, w, r)
	default:
		w.Header().Set("Allow", "POST,GET")
		app.clientError(w, http.StatusMethodNotAllowed)
	}

}

func createSnippetForm(app *application, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new snippet..."))
}
func createSnippet(app *application, w http.ResponseWriter, r *http.Request) {
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut Slowly, slowly!\n\n -Kobayashi Issa"
	expires := "7"
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)

}
