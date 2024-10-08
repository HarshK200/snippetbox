package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/harshk200/snippetbox/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

    snippets, err := app.snippetModel.Latest()
    if err != nil {
        app.serverError(w, err)
        return
    }

    for _, snippet := range snippets {
        fmt.Fprintf(w, "%+v\n", snippet)
    }


	// filesPath := []string{
	// 	"./ui/html/base.tmpl",
    // 	"./ui/html/partials/nav.tmpl",
	// 	"./ui/html/pages/home.tmpl",
	// }
	// ts, err := template.ParseFiles(filesPath...)
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }
	//
	// err = ts.ExecuteTemplate(w, "base", nil)
	// if err != nil {
	// 	app.errorLog.Printf(err.Error())
	// 	app.serverError(w, err)
	// }
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippetModel.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
            return
		} else {
			app.serverError(w, err)
			return
		}
	}

	fmt.Fprintf(w, "%+v", snippet)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7 // expires in 7 days i.e. days

	id, err := app.snippetModel.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}
