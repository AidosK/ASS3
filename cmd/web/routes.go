package main

import (
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet))

	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser))

	mux.Get("/department/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createDepartmentForm))
	mux.Post("/department/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createDepartment))
	mux.Get("/department/:id", dynamicMiddleware.ThenFunc(app.showDepartments))

	mux.Get("/contact", dynamicMiddleware.ThenFunc(app.contact))
	mux.Get("/student", dynamicMiddleware.ThenFunc(app.students))
	mux.Get("/staff", dynamicMiddleware.ThenFunc(app.staff))
	mux.Get("/researcher", dynamicMiddleware.ThenFunc(app.researcher))
	mux.Get("/applicant", dynamicMiddleware.ThenFunc(app.applicant))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))
	return standardMiddleware.Then(mux)
}
