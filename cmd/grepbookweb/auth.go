package main

import (
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/ejamesc/grepbook"
)

func (a *App) LoginPageHandler() HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		p := &localPresenter{PageTitle: "Login", PageURL: "/login", globalPresenter: a.gp}
		a.rndr.HTML(w, http.StatusOK, "login", p)
		return nil
	}
}

func (a *App) LoginPostHandler(db grepbook.UserDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		//email, pass := req.FormValue("email"), req.FormValue("password")

		return nil
	}
}

func (a *App) LogoutHandler(db grepbook.UserDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

func (a *App) SignupPageHandler(db grepbook.UserDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		if db.DoesAnyUserExist() {
			http.Redirect(w, req, "/login", 302)
			return nil
		}
		p := &localPresenter{PageTitle: "Sign Up", PageURL: "/signup", globalPresenter: a.gp}
		err := a.rndr.HTML(w, http.StatusOK, "signup", p)
		if err != nil {
			a.logr.Log(newRenderErrMsg(err))
		}
		return nil
	}
}

func (a *App) SignupPostHandler(db grepbook.UserDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		if db.DoesAnyUserExist() {
			http.Redirect(w, req, "/login", 302)
			return nil
		}
		email, pass := strings.TrimSpace(req.FormValue("email")), req.FormValue("password")
		if !govalidator.IsEmail(email) || strings.TrimSpace(pass) == "" {
			http.Redirect(w, req, "/signup", 302)
			return nil
		}

		_, err := db.CreateUser(email, pass)
		if err != nil {
			return err
		}
		http.Redirect(w, req, "/", 302)
		return nil
	}
}
