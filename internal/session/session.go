package session

import (
	"encoding/json"
	"github.com/habakke/auth-proxy/pkg/util"
	"github.com/rs/zerolog/log"
	"net/http"
)

const sessionCookieName = "session"

type SessionData struct {
	Id    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

func ReadSession(req *http.Request) (*SessionData, error) {
	c, err := req.Cookie("session")
	if err != nil {
		return nil, err
	}

	data, err := util.Base64Decode(c.Value)
	if err != nil {
		log.Debug().AnErr("err", err).Msg("failed to base64 decode auth cookie data")
		return nil, err
	}
	s := SessionData{}
	if err := json.Unmarshal(data, &s); err != nil {
		log.Debug().AnErr("err", err).Str("data", c.Value).Msg("failed to unmarshal auth cookie data")
		return nil, err
	}

	return &s, nil
}

func AttachSession(w http.ResponseWriter, session SessionData) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:   sessionCookieName,
		Value:  util.Base64Encode(data),
		Path:   "/",
		MaxAge: 86400,
	}
	http.SetCookie(w, &cookie)
	return nil
}

func RemoveSession(res http.ResponseWriter) {
	c := &http.Cookie{
		Name:   sessionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(res, c)
}
