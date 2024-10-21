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

func ping(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("OK"))
}

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
	form := &snippetCreateFormData{Expires: 365}
	data.Form = form

	app.render(w, http.StatusOK, "create.tmpl", data)
}

// NOTE: not having a property instead embedding the validator here i.e. the snippetCreateFormData struct inherits from the validator
type snippetCreateFormData struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var formData snippetCreateFormData
	err := app.decodePostForm(r, &formData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// NOTE: form data validation
	formData.CheckField(validator.NotBlank(formData.Title), "title", "This field cannot be blank")
	formData.CheckField(validator.MaxChars(formData.Title, 100), "title", "This field cannot be more than 100 characters long")
	formData.CheckField(validator.NotBlank(formData.Content), "content", "This field cannot be blank")
	formData.CheckField(validator.PermittedValue(formData.Expires, 1, 7, 365), "expires", "This field can only be 1, 7 or 365")

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

type userSignupFormData struct {
	Name                string `form:"name"` // NOTE: when decoding form data the name="name" decodes to this field
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupFormData{}

	app.render(w, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var formData userSignupFormData
	err := app.decodePostForm(r, &formData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	formData.CheckField(validator.NotBlank(formData.Name), "name", "This field cannot be blank")
	formData.CheckField(validator.NotBlank(formData.Email), "email", "This field cannot be blank")
	formData.CheckField(validator.Matches(formData.Email, validator.EmailRX), "email", "This field must be a valid email address")
	formData.CheckField(validator.NotBlank(formData.Password), "password", "This field cannot be blank")
	formData.CheckField(validator.MinChars(formData.Password, 8), "password", "password must be at least 8 characters long")

	if !formData.Valid() {
		data := app.newTemplateData(r)
		data.Form = formData
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	// NOTE: create new user...
	err = app.userModel.Insert(formData.Name, formData.Email, formData.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			formData.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = formData
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please login.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userLoginData struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginData{}

	app.render(w, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var formData userLoginData
	app.decodePostForm(r, &formData)

	formData.CheckField(validator.NotBlank(formData.Email), "email", "This field cannot be blank")
	formData.CheckField(validator.Matches(formData.Email, validator.EmailRX), "email", "This field must be a valid email")
	formData.CheckField(validator.NotBlank(formData.Password), "password", "This field cannot be blank")

	if !formData.Valid() {
		data := app.newTemplateData(r)
		data.Form = formData
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	userID, err := app.userModel.Authenticate(formData.Email, formData.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			formData.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = formData

			app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
			return
		}

		app.serverError(w, err)
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// NOTE: adding the authenticatedUserID key in the sessionData for future authnetication checks
	app.sessionManager.Put(r.Context(), "authenticatedUserID", userID)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogout(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	app.sessionManager.Put(r.Context(), "flash", "You've been logged out sucessfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
