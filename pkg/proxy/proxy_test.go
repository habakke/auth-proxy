package proxy

import (
	"github.com/habakke/auth-proxy/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProxyServeHTTP(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/", nil)
	if err != nil {
		t.Fatal(err)
	}

	_ = os.Setenv("TOKEN", "token_goes_here")
	_ = os.Setenv("TARGET", "http://kubernetes-dashboard.k8s.matrise.net")
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	p := &Proxy{}
	lmw := util.LoggingMiddleware(log.Logger)
	lp := lmw(p)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(lp.ServeHTTP)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var keyword = "Kubernetes Dashboard"
	if !strings.Contains(rr.Body.String(), keyword) {
		t.Errorf("handler returned wrong body: missing keyword %v", keyword)
	}
}
