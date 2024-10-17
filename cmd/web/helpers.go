package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	data := &templateData{
		CurrentYear:     time.Now().Year(),
		IsAuthenticated: false,
		NotFound:        true,
	}

	app.render(w, http.StatusNotFound, "notFound.tmpl", data)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := errors.New(fmt.Sprintf("template %s doesn't exists", page))
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	// NOTE: executing the template to make sure we are don't encounter any error
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	// NOTE: ok is returned when doing type assersions. true if assertion was successful else false
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool) // NOTE: returns nil if no value is associated with the key
	if !ok {
		return false
	}

	return isAuthenticated
}
