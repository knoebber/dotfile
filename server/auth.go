package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/usererror"
)

const passwordResetSubject = "Dotfilehub Password Reset"

func redirectWhenLoggedIn(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if p.Session != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return true
	}
	return false
}

func handleLogin(secure bool) pageBuilder {
	return func(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
		return login(w, r, p, secure)
	}
}

func handleAccountRecovery(config Config) pageBuilder {
	return func(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
		if config.SMTP == nil {
			return p.setError(w, usererror.Invalid("This server isn't configured for SMTP."))
		}
		email := r.Form.Get("email")

		// Don't let people see if emails exist or not by checking how long the page takes to load.
		go func() {
			token, err := db.SetPasswordResetToken(db.Connection, r.Form.Get("email"))
			if err != nil {
				// Log the error but don't tell the user.
				log.Print("reset password form: ", err)
				return
			}

			resetURL := fmt.Sprintf("%s/reset_password?token=%s", config.URL(r), token)
			body := "Please click the following link to reset your dotfilehub account:\n" + resetURL
			if err := mail(config.SMTP, email, passwordResetSubject, body); err != nil {
				log.Print(err)
			} else {
				log.Printf("sent mail to %q for password reset", email)
			}
		}()

		p.flashSuccess("Request processed.")
		return
	}
}

func loadAccountRecovery(config Config) pageBuilder {
	return func(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
		if config.SMTP == nil {
			p.Data["disabled"] = true
			return
		}

		p.Data["sender"] = config.SMTP.Sender
		p.Data["subject"] = passwordResetSubject
		return
	}
}

func handlePasswordReset(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	token := r.URL.Query().Get("token")
	password := r.Form.Get("password")
	confirm := r.Form.Get("confirm")

	if password != confirm {
		return p.setError(w, usererror.Invalid("Passwords do not match."))
	}

	username, err := db.ResetPassword(db.Connection, token, password)
	if err != nil {
		return p.setError(w, err)
	}

	log.Printf("reset password for %q", username)
	p.flashSuccess(fmt.Sprintf("Password updated. Your username is %q", username))
	return
}

func login(w http.ResponseWriter, r *http.Request, p *Page, secure bool) (done bool) {
	s, err := db.UserLogin(db.Connection, r.Form.Get("username"), r.Form.Get("password"), r.RemoteAddr)
	if err != nil {
		// Print the real error and show the user a generic catch all.
		log.Print(err)
		return p.setError(w, usererror.Invalid("Username or password is incorrect."))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    s.Session,
		Secure:   secure,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return true
}

func handleSignup(secure bool) pageBuilder {
	return func(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		confirm := r.Form.Get("confirm")

		if password != confirm {
			p.setError(w, usererror.Invalid("Passwords do not match."))
			return false
		}

		_, err := db.CreateUser(db.Connection, username, password)
		if err != nil {
			return p.setError(w, err)
		}

		return login(w, r, p, secure)
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if err := db.Logout(db.Connection, p.session()); err != nil {
		log.Print(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return true
}

func loginHandler(secure bool) http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "login.tmpl",
		title:        loginTitle,
		loadData:     redirectWhenLoggedIn,
		handleForm:   handleLogin(secure),
	})
}

func signupHandler(secure bool) http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "signup.tmpl",
		title:        "Signup",
		loadData:     redirectWhenLoggedIn,
		handleForm:   handleSignup(secure),
	})
}

func logoutHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		handleForm: handleLogout,
		protected:  true,
	})
}

func accountRecoveryHandler(config Config) http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "account_recovery.tmpl",
		title:        "Account Recovery",
		loadData:     loadAccountRecovery(config),
		handleForm:   handleAccountRecovery(config),
	})
}

func resetPasswordHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "reset_password.tmpl",
		title:        "Reset Password",
		handleForm:   handlePasswordReset,
	})
}
