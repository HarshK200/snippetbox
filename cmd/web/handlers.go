package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/harshk200/snippetbox/internal/models"
	"github.com/harshk200/snippetbox/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippetModel.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets
	app.render(w, http.StatusOK, "home.tmpl", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
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

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, http.StatusOK, "view.tmpl", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	form := &snippetCreateForm{Expires: 365}
	data.Form = form

	app.render(w, http.StatusOK, "create.tmpl", data)
}

// NOTE: not having a property instead embedding the validator here i.e. the snippetCreateForm struct inherits from the validator
type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var formData snippetCreateForm
	err := app.decodePostForm(r, &formData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// NOTE: form data validation
	formData.CheckField(validator.NotBlank(formData.Title), "title", "This field cannot be blank")
	formData.CheckField(validator.MaxChars(formData.Title, 100), "title", "This field cannot be more than 100 characters long")
	formData.CheckField(validator.NotBlank(formData.Content), "content", "This field cannot be blank")
	formData.CheckField(validator.PermittedInt(formData.Expires, 1, 7, 365), "expires", "This field can only be 1, 7 or 365")

	// NOTE: if any errors re-render the form
	if !formData.Valid() {
		data := app.newTemplateData(r)
		data.Form = formData
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	id, err := app.snippetModel.Insert(formData.Title, formData.Content, formData.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// NOTE: after the session data is created succesfully we put that info in the session manager
	app.sessionManager.Put(r.Context(), "flash", "Snippet created succesfully!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "show html form for user signup")
}
func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "create a new user POST route...")
}
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "show html form for user login")
}
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "authenticate the email and password && login the user...")
}
func (app *application) userLogout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Logout the user...")
}
