package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/habakke/auth-proxy/internal/cookie"
	"github.com/habakke/auth-proxy/internal/session"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  os.Getenv("GOOGLE_OAUTH_CALLBACK_URL"), // Ex. https://<domain>/auth/google/callback
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleURLAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

type GoogleUserInfo struct {
	ID         string `json:"id,omitempty"`
	Email      string `json:"email,omitempty"`
	Verified   bool   `json:"verified_email"`
	Name       string `json:"name,omitempty"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
	Picture    string `json:"picture,omitempty"`
	HD         string `json:"hd,omitempty"`
}

func (u GoogleUserInfo) GetID() string {
	return u.ID
}

func (u GoogleUserInfo) GetUsername() string {
	return u.Email
}

func (u GoogleUserInfo) GetName() string {
	return u.Name
}

func (u GoogleUserInfo) GetEmail() string {
	return u.Email
}

type GoogleProvider struct {
	*ProviderData
	Token *Token
}

func NewGoogleProvider(p *ProviderData) *GoogleProvider {
	p.Name = "Google"
	return &GoogleProvider{
		ProviderData: p,
	}
}

func (p *GoogleProvider) Data() *ProviderData {
	return p.ProviderData
}

func (p *GoogleProvider) GetUser() (User, error) {
	response, err := http.Get(oauthGoogleURLAPI + p.Token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}

	var user = GoogleUserInfo{}
	if err := json.Unmarshal(data, &user); err != nil {
		log.Info().AnErr("err", err).Str("data", string(data)).Msg("failed to unmarshal Google userinfo")
		return nil, fmt.Errorf("failed to unmarshal Google userinfo")
	}
	return user, nil
}

func (p *GoogleProvider) Exchange(code string) error {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("code exchange failed: %s", err.Error())
	}

	p.Token = &Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
	return nil
}

func (p *GoogleProvider) GetCallbackPath() string {
	return fmt.Sprintf("/auth/%s/callback", strings.ToLower(p.Name))
}

func (p *GoogleProvider) GetLoginPath() string {
	return fmt.Sprintf("/auth/%s/login", strings.ToLower(p.Name))
}

func (p *GoogleProvider) GetProviderLoginURL(res http.ResponseWriter) (*url.URL, error) {
	nonce, err := cookie.Nonce()
	if err != nil {
		log.Error().AnErr("err", err).Msg("failed to generate nonce")
		return nil, err
	}

	http.SetCookie(res, cookie.MakeCSRFCookie(nonce))
	urlString := googleOauthConfig.AuthCodeURL(nonce)
	u, err := url.Parse(urlString)
	if err != nil {
		log.Error().AnErr("err", err).Msg("failed to generate an authentication url")
		return nil, err
	}

	return u, nil
}

func (p *GoogleProvider) AuthenticateSession(data *session.Data) bool {
	return true
}
