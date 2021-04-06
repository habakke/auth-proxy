package auth

import (
	"fmt"
	"github.com/habakke/auth-proxy/internal/session"
	"github.com/rs/zerolog/log"
	"net/http"
)

type localAuth struct {
	token string
}

func NewAuthLocal(token string) *localAuth {
	return &localAuth{
		token: token,
	}
}

func (a *localAuth) Authenticate(res http.ResponseWriter, req *http.Request) bool {
	email := req.FormValue("email")
	password := req.FormValue("password")

	if len(email) == 0 || len(password) == 0 {
		log.Trace().Msg("email or password is missing")
		return false
	}

	// TODO implement better logic here
	if email == "habakke@matrise.net" && password == "1234" {
		s := session.SessionData{
			Id:    "0",
			Email: email,
			Name:  email,
		}
		err := session.AttachSession(res, s)
		if err != nil {
			log.Info().AnErr("err", err).Msg("failed to create session")
			return false
		}
		return true
	}

	return false
}

func (a *localAuth) LoginHandler(res http.ResponseWriter, req *http.Request) {
	if !a.Authenticate(res, req) {
		session.RemoveSession(res)
		http.Redirect(res, req, fmt.Sprintf("/auth/login?path=%s", req.URL.Query()["path"]), http.StatusTemporaryRedirect)
	} else {
		http.Redirect(res, req, "/?", http.StatusTemporaryRedirect)
	}
}
