package session

import (
	"fmt"
	"github.com/habakke/auth-proxy/internal/cookie"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func bootstrapTestServer(sm *Manager) (srv *httptest.Server) {

	srv = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/attach":
			_ = sm.AttachSession(res, Data{ID: "test", Name: "test", Authorized: false})
		default:
			hostname, _, _ := net.SplitHostPort(req.Host)
			_, _ = res.Write([]byte(fmt.Sprintf("hostname=%s\n", hostname)))
			_, _ = res.Write([]byte(fmt.Sprintf("path=%s\n", req.URL.Path)))

		}
		res.WriteHeader(200)
	}))
	return srv
}

func TestMakeSessionCookie(t *testing.T) {
	cookieSeed := "0123456789abcdefghijklmnopqrstuv"
	cookieKey := "2345asdYDS!2012L"
	cookiePayload := "payload goes here"

	sm := NewManager(cookieSeed, cookieKey)
	c, err := sm.MakeSessionCookie(cookieSeed, cookieKey, cookiePayload)
	if err != nil {
		t.Errorf("failed to make session cookie %e", err)
	}

	data, err := sm.ReadSessionCookie(c, cookieSeed, cookieKey)
	if err != nil {
		t.Errorf("faile to read session cookie %e", err)
	}

	assert.Equal(t, cookiePayload, data)
}

func TestAuthenticatedSession(t *testing.T) {
	cookieSeed := "0123456789abcdefghijklmnopqrstuv"
	cookieKey := "2345asdYDS!2012L"
	cookiePayload := "payload goes here"

	sm := NewManager(cookieSeed, cookieKey)
	c, err := sm.MakeSessionCookie(cookieSeed, cookieKey, cookiePayload)
	if err != nil {
		t.Errorf("failed to create session cookie")
	}

	encryptedData, _, ok := cookie.Validate(c, cookieSeed, time.Hour*24)
	if !ok {
		t.Errorf("failed to validate session cookie")
	}

	data, err := cookie.DecryptCookieValue(cookieKey, encryptedData)
	if err != nil {
		t.Errorf("failed to decrypt cookie data")
	}

	assert.Equal(t, cookiePayload, data, "payload is not as expected")
}

func TestAttachSession(t *testing.T) {
	cookieSeed := "0123456789abcdefghijklmnopqrstuv"
	cookieKey := "2345asdYDS!2012L"
	sm := NewManager(cookieSeed, cookieKey)
	srv := bootstrapTestServer(sm)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/attach", srv.URL), nil)
	res, _ := http.DefaultClient.Do(req)
	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))

	// Hokey pokey need to adapt to the session interface
	for _, c := range res.Cookies() {
		req.AddCookie(c)
	}

	data, err := sm.ReadSession(req)
	if err != nil {
		t.Errorf("failed to read session: %e", err)
	}

	assert.Equal(t, "test", data.Name)
}
