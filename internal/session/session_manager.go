package session

import (
	"encoding/json"
	"fmt"
	"github.com/habakke/auth-proxy/internal/cookie"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

const MaxSessionDuration = 30
const SessionCookieName = "session"

type Manager struct {
	cookieSeed string
	cookieKey  string
}

func NewManager(cookieSeed string, cookieKey string) *Manager {
	return &Manager{
		cookieSeed: cookieSeed,
		cookieKey:  cookieKey,
	}
}

func (m *Manager) MakeSessionCookie(seed string, key string, payload string) (*http.Cookie, error) {
	encryptedPayload, err := cookie.EncryptCookieValue(key, payload)
	if err != nil {
		return nil, err
	}
	v := cookie.SignCookieValue(seed, SessionCookieName, encryptedPayload, time.Now())
	return cookie.MakeCookie(SessionCookieName, v), nil
}

func (m *Manager) ReadSessionCookie(c *http.Cookie, cookieSeed string, cookieKey string) (string, error) {
	if c.Name != SessionCookieName {
		return "", fmt.Errorf("cookie is not a session cookie")
	}

	encryptedValue, _, ok := cookie.Validate(c, cookieSeed, time.Hour*24*MaxSessionDuration)
	if !ok {
		return "", fmt.Errorf("failed to validate cookie")
	}
	data, err := cookie.DecryptCookieValue(cookieKey, encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt cookie payload")
	}

	return data, nil
}

func (m *Manager) ReadSession(req *http.Request) (*Data, error) {
	c, err := req.Cookie(SessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("cookie %q not present", SessionCookieName)
	}

	var data string
	data, err = m.ReadSessionCookie(c, m.cookieSeed, m.cookieKey)
	if err != nil {
		log.Error().AnErr("err", err).Str("data", c.Value).Msg("failed to read session cookie")
		return nil, err
	}

	d := Data{}
	if err = json.Unmarshal([]byte(data), &d); err != nil {
		log.Error().AnErr("err", err).Str("data", c.Value).Msg("failed to unmarshal auth cookie data")
		return nil, err
	}

	return &d, err
}

func (m *Manager) AttachSession(res http.ResponseWriter, session Data) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	c, err := m.MakeSessionCookie(m.cookieSeed, m.cookieKey, string(data))
	if err != nil {
		return err
	}

	http.SetCookie(res, c)
	return nil
}

func (m *Manager) RemoveSession(res http.ResponseWriter) {
	c := cookie.MakeInvalidationCookie(SessionCookieName)
	http.SetCookie(res, c)
}
