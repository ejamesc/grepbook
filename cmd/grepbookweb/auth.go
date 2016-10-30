package main

import (
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/ejamesc/grepbook"
)

func (a *App) LoginPageHandler() HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		user := getUser(req)
		if user != nil {
			http.Redirect(w, req, "/", 302)
			return nil
		}
		fs := a.getFlashes(w, req)
		p := &struct {
			Flashes []interface{}
			localPresenter
		}{
			Flashes:        fs,
			localPresenter: localPresenter{PageTitle: "Login", PageURL: "/login", globalPresenter: a.gp}}

		err := a.rndr.HTML(w, http.StatusOK, "login", p)
		if err != nil {
			a.logr.Log(newRenderErrMsg(err))
		}
		return nil
	}
}

func (a *App) LoginPostHandler(db grepbook.UserDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		user := getUser(req)
		if user != nil {
			http.Redirect(w, req, "/", 302)
			return nil
		}
		email, pass := req.FormValue("email"), req.FormValue("password")
		if !govalidator.IsEmail(email) {
			a.saveFlash(w, req, "That's not a valid email address")
			http.Redirect(w, req, "/login", 302)
			return newError(400, "Invalid email provided", nil)
		}

		if strings.TrimSpace(pass) == "" {
			a.saveFlash(w, req, "You need to provide a password")
			http.Redirect(w, req, "/login", 302)
			return newError(400, "No password provided", nil)
		}

		user, err := db.GetUser(email)
		if err != nil {
			a.saveFlash(w, req, "Whoops, your email or password is incorrect!")
			a.logr.Log("Error getting user by email: %s", err)
			http.Redirect(w, req, "/login", 302)
			return nil
		}

		ss, err := a.store.Get(req, SessionName)
		if err != nil {
			return newError(500, "error getting store", err)
		}

		if db.IsUserPasswordCorrect(user.Email, pass) {
			sess, err := db.CreateSessionForUser(user.Email)
			if err != nil {
				a.logr.Log("error creating user session %s", err)
				http.Redirect(w, req, "/login", 302)
				return newSessionSaveError(err)
			}

			ss.Values[SessionKeyName] = sess.Key
			ss.Save(req, w)
			http.Redirect(w, req, "/", 302)
		} else {
			a.saveFlash(w, req, "Wrong email or password!")
			ss.Save(req, w)
			http.Redirect(w, req, "/login", 302)
		}

		return nil
	}
}

func (a *App) LogoutHandler() HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		session, _ := a.store.Get(req, SessionName)
		delete(session.Values, SessionKeyName)
		session.Save(req, w)
		http.Redirect(w, req, "/", 302)
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

		user, err := db.CreateUser(email, pass)
		if err != nil {
			return err
		}

		sess, err := db.CreateSessionForUser(user.Email)
		if err != nil {
			return err
		}
		ss, _ := a.store.Get(req, SessionName)
		ss.Values[SessionKeyName] = sess.Key
		ss.Save(req, w)
		http.Redirect(w, req, "/", 302)
		return nil
	}
}
