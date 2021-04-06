package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/habakke/auth-proxy/internal/session"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8081/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

/*
 {
  "id": "113699748174122511172",
  "email": "habakke@matrise.net",
  "verified_email": true,
  "name": "Håvard Bakke",
  "given_name": "Håvard",
  "family_name": "Bakke",
  "picture": "https://lh3.googleusercontent.com/a-/AOh14Gh7BJIdgVp-xw6RGWfUYVY5JfRRiROcZeS6vB1HmKU=s96-c",
  "hd": "matrise.net"
}
*/

type GoogleUserInfo struct {
	Id         string `json:"id,omitempty"`
	Email      string `json:"email,omitempty"`
	Verified   bool   `json:"verified_email"`
	Name       string `json:"name,omitempty"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
	Picture    string `json:"picture,omitempty"`
	HD         string `json:"hd,omitempty"`
}

type Oauth2Google struct {
	token string
}

func NewGoogleOauth2(token string) *Oauth2Google {
	return &Oauth2Google{
		token: token,
	}
}

func (o *Oauth2Google) LoginHandler(res http.ResponseWriter, req *http.Request) {
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(res)
	u := googleOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(res, req, u, http.StatusTemporaryRedirect)
}

func (o *Oauth2Google) CallbackHandler(res http.ResponseWriter, req *http.Request) {
	// Read oauthState from Cookie
	oauthState, _ := req.Cookie("oauthstate")

	if req.FormValue("state") != oauthState.Value {
		log.Info().Msg("invalid oauth google state")
		http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := getUserDataFromGoogle(req.FormValue("code"))
	if err != nil {
		log.Info().AnErr("err", err).Msg("failed to get userdata from Google")
		http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
		return
	}

	user := GoogleUserInfo{}
	if err := json.Unmarshal(data, &user); err != nil {
		log.Info().AnErr("err", err).Str("data", string(data)).Msg("failed to unmarshal auth cookie data")
		return
	}

	log.Debug().Str("id", user.Id).Str("user", user.Email).Msg("user logged in")

	s := session.SessionData{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}
	_ = session.AttachSession(res, s)
	http.Redirect(res, req, "/?", http.StatusTemporaryRedirect)
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}
