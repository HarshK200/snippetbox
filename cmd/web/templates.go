package main

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/harshk200/snippetbox/internal/models"
	"github.com/harshk200/snippetbox/ui"
	"github.com/justinas/nosurf"
)

type templateData struct {
	CurrentYear     int
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r), // NOTE: checking if this request is authenticated
		CSRFToken:       nosurf.Token(r),
	}
}

func humanDate(t time.Time) string {
    if t.IsZero(){
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
			"html/partials/*.tmpl",
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
