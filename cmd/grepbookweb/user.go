package main

import (
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/ejamesc/grepbook"
)

var reloginTarget = "/user"

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
		email, name := req.FormValue("email"), req.FormValue("name")
		oldPass, newPass, newPass2 := req.FormValue("old-password"), req.FormValue("new-password"), req.FormValue("new-password2")

		email, name = strings.TrimSpace(email), strings.TrimSpace(name)
		oldPass, newPass, newPass2 = strings.TrimSpace(oldPass), strings.TrimSpace(newPass), strings.TrimSpace(newPass2)

		user := getUser(req)
		userDelta := grepbook.UserDelta{}

		isPasswordUpdated := oldPass != "" || newPass != "" || newPass2 != ""
		if isPasswordUpdated {
			if oldPass == "" {
				redirectToUserForm(a, w, req, "You need to provide your old password", 302)
				return newError(400, "No old password provided", nil)
			}
			if newPass != newPass2 {
				redirectToUserForm(a, w, req, "Your new passwords do not match!", 302)
				return newError(400, "New passwords do not match", nil)
			}
			if newPass == "" || newPass2 == "" {
				redirectToUserForm(a, w, req, "One of the new password slots was left empty", 302)
				return newError(400, "One of the new password slots was left empty", nil)
			}
			if !db.IsUserPasswordCorrect(user.Email, oldPass) {
				redirectToUserForm(a, w, req, "Old password wrong", 302)
				return newError(400, "Old password wrong", nil)
			}
			userDelta.Password = newPass
		} else {
			if email != "" {
				if govalidator.IsEmail(email) {
					userDelta.Email = email
				} else {
					redirectToUserForm(a, w, req, "That's not a valid email address", 302)
					return newError(400, "Invalid email provided", nil)
				}
			}
			userDelta.Name = name
		}

		_, err := db.UpdateUser(user.Email, userDelta)
		if err != nil {
			redirectToUserForm(a, w, req, "An error occurred when saving your data", 302)
			return err
		}

		if userDelta.Name != "" {
			a.gp.Username = userDelta.Name
		}

		redirectToUserForm(a, w, req, "", 302)
		return nil
	}
}

func redirectToUserForm(a *App, w http.ResponseWriter, req *http.Request, message string, code int) {
	if message != "" {
		a.saveFlash(w, req, message)
	}
	http.Redirect(w, req, reloginTarget, code)
}
