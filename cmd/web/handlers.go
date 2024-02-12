package main

import (
	"aidoskanatbay.net/snippetbox/pkg/forms"
	"aidoskanatbay.net/snippetbox/pkg/models"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil)})
}
func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MaxLength("name", 255)
	form.MaxLength("email", 255)
	form.MatchesPattern("email", forms.EmailRx)
	form.MinLength("password", 10)

	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}
	taken, err := app.users.EmailTaken(form.Get("email"))
	if err != nil {
		// Handle the error
		app.serverError(w, err)
		return
	}
	if taken {
		// If email is taken, add an error message for the email field
		form.Errors.Add("email", "Address is already in use")
		// Render the signup page again with the error message
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}
	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Errors.Add("email", "Address is already in use")
			app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "flash", "Your signup successfully. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil)})
}
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		return
	}

	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or password incorrect")
			app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "authenticatedUserID", id)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "authenticatedUserID")
	app.session.Put(r, "flash", "You have been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	departments, err := app.departments.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := templateData{
		Snippets:    snippets,
		Departments: departments,
	}

	app.render(w, r, "home.page.tmpl", &data)
}

func (app *application) showDepartments(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	s, err := app.departments.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.render(w, r, "dep.show.page.tmpl", &templateData{
		Department: s,
	})

}
func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	// Pat doesn't strip the colon from the named capture key, so we need to
	// get the value of ":id" from the query string instead of "id".
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	s, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		Snippet: s,
	})
}

func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		// Pass a new empty forms.Form object to the template.
		Form: forms.New(nil),
	})
}
func (app *application) createDepartmentForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "dep.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}
	id, err := app.snippets.Insert(form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}
func (app *application) createDepartment(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("dep_name", "staff_quantity")
	form.MaxLength("dep_name", 100)
	form.PermittedValues("staff_quantity", "100")

	if !form.Valid() {
		app.render(w, r, "dep.page.tmpl", &templateData{Form: form})
		return
	}

	staffQuantity, err := strconv.Atoi(form.Get("staff_quantity"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	id, err := app.departments.Insert(form.Get("dep_name"), staffQuantity)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Department successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/department/%d", id), http.StatusSeeOther)
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle form submission

		// Access form values
		name := r.FormValue("name")
		email := r.FormValue("email")
		message := r.FormValue("message")

		// Do something with the form data, e.g., send an email, save to database, etc.

		// For now, let's just print the form data
		fmt.Printf("Name: %s\nEmail: %s\nMessage: %s\n", name, email, message)

		// Redirect to a thank you page or display a success message
		http.Redirect(w, r, "/thank-you", http.StatusSeeOther)
		return
	}

	// Render the contact form page for GET requests
	app.render(w, r, "contact.page.tmpl", nil)
}
func (app *application) students(w http.ResponseWriter, r *http.Request) {
	// Retrieve snippets with the category "students"
	snippets, err := app.snippets.Student("students")
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the new render helper.
	app.render(w, r, "student.page.tmpl", &templateData{
		Snippets: snippets,
	})
}

func (app *application) staff(w http.ResponseWriter, r *http.Request) {
	// Retrieve snippets with the category "students"
	snippets, err := app.snippets.Staff("staff")
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the new render helper.
	app.render(w, r, "staff.page.tmpl", &templateData{
		Snippets: snippets,
	})
}
func (app *application) applicant(w http.ResponseWriter, r *http.Request) {
	// Retrieve snippets with the category "students"
	snippets, err := app.snippets.Applicant("staff")
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the new render helper.
	app.render(w, r, "applicant.page.tmpl", &templateData{
		Snippets: snippets,
	})
}

func (app *application) researcher(w http.ResponseWriter, r *http.Request) {
	// Retrieve snippets with the category "students"
	snippets, err := app.snippets.Researcher("researcher")
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the new render helper.
	app.render(w, r, "researcher.page.tmpl", &templateData{
		Snippets: snippets,
	})
}
