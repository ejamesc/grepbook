package main

import (
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/ejamesc/grepbook"
)

func (a *App) UserProfileHandler() HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		user := getUser(req)
		if user == nil {
			return newError(500, "User doesn't exist", nil)
		}

		fs := a.getFlashes(w, req)
		pp := &struct {
			Flashes []interface{}
			localPresenter
		}{
			Flashes: fs,
			localPresenter: localPresenter{
				PageTitle:       user.Email + " Profile",
				PageURL:         "/user",
				globalPresenter: a.gp,
				User:            user,
			},
		}
		err := a.rndr.HTML(w, http.StatusOK, "user", pp)
		if err != nil {
			a.logr.Log(newRenderErrMsg(err))
		}
		return nil
	}
}

func (a *App) UserEditHandler(db grepbook.UserDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		reloginTarget := "/user"
		email, name := req.FormValue("email"), req.FormValue("name")
		oldPass, newPass, newPass2 := req.FormValue("old-password"), req.FormValue("new-password"), req.FormValue("new-password2")

		email, name = strings.TrimSpace(email), strings.TrimSpace(name)
		oldPass, newPass, newPass2 = strings.TrimSpace(oldPass), strings.TrimSpace(newPass), strings.TrimSpace(newPass2)

		if !govalidator.IsEmail(email) {
			a.saveFlash(w, req, "That's not a valid email address")
			http.Redirect(w, req, reloginTarget, 302)
			return newError(400, "Invalid email provided", nil)
		}

		userDelta := grepbook.UserDelta{Email: email}

		if oldPass != "" || newPass != "" || newPass2 != "" {
			if oldPass == "" {
				a.saveFlash(w, req, "You need to provide your old password")
				http.Redirect(w, req, reloginTarget, 302)
				return newError(400, "No old password provided", nil)
			}
			if newPass != newPass2 {
				a.saveFlash(w, req, "Your new passwords do not match!")
				http.Redirect(w, req, reloginTarget, 302)
				return newError(400, "New passwords do not match", nil)
			}
			if newPass == "" || newPass2 == "" {
				a.saveFlash(w, req, "One of the new password slots was left empty")
				http.Redirect(w, req, reloginTarget, 302)
				return newError(400, "One of the new password slots was left empty", nil)
			}
		}

		user := getUser(req)
		if !db.IsUserPasswordCorrect(user.Email, oldPass) {
			a.saveFlash(w, req, "Old password wrong")
			http.Redirect(w, req, reloginTarget, 302)
			return newError(400, "Old password wrong", nil)

		}
		userDelta.Password = newPass
		userDelta.Name = name

		_, err := db.UpdateUser(user.Email, userDelta)
		if err != nil {
			a.saveFlash(w, req, "An error occured when saving your data")
			http.Redirect(w, req, reloginTarget, 302)
			return err
		}

		if userDelta.Name != "" {
			a.gp.Username = userDelta.Name
		}

		http.Redirect(w, req, reloginTarget, 302)
		return nil
	}
}
