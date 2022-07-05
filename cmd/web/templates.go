package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/xyedo/snippetbox/internal/models"
	"github.com/xyedo/snippetbox/ui"
)

type templateData struct {
	IsAuthenticated bool
	CSRFToken       string
	CurrentYear     int
	Flash           string
	Form            any
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	User            *models.User
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		pattern := []string{
			"html/base.tmpl",
			"html/partials/*tmpl",
			page,
		}
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, pattern...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
