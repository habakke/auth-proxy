package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/habakke/auth-proxy/internal/auth/providers"
	"github.com/habakke/auth-proxy/internal/session"
	"github.com/habakke/auth-proxy/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func makeProxyRequest(proxy *Proxy, req *http.Request) (rr *httptest.ResponseRecorder) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	lmw := util.NewLoggingMiddleware(log.Logger)
	lp := lmw(proxy)

	rr = httptest.NewRecorder()
	loggingProxyHandler := http.HandlerFunc(lp.ServeHTTP)
	loggingProxyHandler.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, res *httptest.ResponseRecorder, code int) {
	if status := res.Code; status != code {
		t.Errorf("response returned incorrect status code: got %v want %v", status, code)
	}
}

func checkResponseBody(t *testing.T, res *httptest.ResponseRecorder, keyword string) {
	if !strings.Contains(res.Body.String(), keyword) {
		t.Errorf("response body is missing keyword '%s'", keyword)
	}
}

func bootstrapProxy(provider providers.Provider, cookieSeed string, cookieKey string) (target *httptest.Server, proxy *Proxy, sm *session.Manager) {
	sm = session.NewManager(cookieSeed, cookieKey)
	target = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		hostname, _, _ := net.SplitHostPort(req.Host)
		_, _ = res.Write([]byte(fmt.Sprintf("hostname=%s\n", hostname)))
		_, _ = res.Write([]byte(fmt.Sprintf("path=%s\n", req.URL.Path)))
	}))
	proxy = New(target.URL, provider, sm)
	return target, proxy, sm
}

func TestMakeUnauthenticatedProxyRequest(t *testing.T) {
	cookieSeed := "0123456789abcdefghijklmnopqrstuv"
	cookieKey := "2345asdYDS!2012L"
	provider := providers.New("Google", &providers.ProviderData{})
	target, proxy, _ := bootstrapProxy(provider, cookieSeed, cookieKey)
	defer target.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/", target.URL), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := makeProxyRequest(proxy, req)
	checkResponseCode(t, rr, http.StatusFound)
	checkResponseBody(t, rr, "<a href=\"/auth/login?p=/\">Found</a>.\n\n")
}

func TestMakeAuthenticatedProxyRequest(t *testing.T) {
	cookieSeed := "0123456789abcdefghijklmnopqrstuv"
	cookieKey := "2345asdYDS!2012L"
	provider := providers.New("Google", &providers.ProviderData{})
	target, proxy, sm := bootstrapProxy(provider, cookieSeed, cookieKey)
	defer target.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/test1234", target.URL), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Fake a logged in session by attaching a session cookie to the request
	payload, _ := json.Marshal(session.Data{ID: "test", Name: "Test"})
	c, err := sm.MakeSessionCookie(cookieSeed, cookieKey, string(payload))
	if err != nil {
		t.Errorf("failed to create session cookie")
	}
	req.AddCookie(c)

	// Make request
	rr := makeProxyRequest(proxy, req)
	checkResponseCode(t, rr, http.StatusOK)
	checkResponseBody(t, rr, "test1234")
}
