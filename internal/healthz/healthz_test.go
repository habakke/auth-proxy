package healthz

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHealthz(t *testing.T) {
	_ = os.Setenv("PORT", "8080")
	_ = os.Setenv("LOGGING", "TRUE")
	_ = os.Setenv("TOKEN", "token_goes_here")
	_ = os.Setenv("TARGET", "http://kubernetes-dashboard.k8s.matrise.net")

	req, err := http.NewRequest("GET", "http://localhost:8080/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := Handler()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
