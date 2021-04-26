package providers

import (
	"github.com/habakke/auth-proxy/internal/session"
	"net/http"
	"net/url"
	"time"
)

type Provider interface {
	Data() *ProviderData
	GetUser() (User, error)
	Exchange(code string) error
	GetLoginPath() string
	GetCallbackPath() string
	GetProviderLoginURL(res http.ResponseWriter) (*url.URL, error)

	AuthenticateSession(data *session.Data) bool
}

type User interface {
	GetID() string
	GetUsername() string
	GetName() string
	GetEmail() string
}

type Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func New(provider string, p *ProviderData) Provider {
	switch provider {
	case "google":
		return NewGoogleProvider(p)
	default:
		return NewGoogleProvider(p)
	}
}
