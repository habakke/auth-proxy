package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/habakke/auth-proxy/internal/auth/providers"
	"github.com/habakke/auth-proxy/internal/healthz"
	"github.com/habakke/auth-proxy/internal/metrics"
	"github.com/habakke/auth-proxy/internal/session"
	"github.com/habakke/auth-proxy/pkg/util"
	"github.com/habakke/auth-proxy/pkg/util/logutils"
	"github.com/habakke/auth-proxy/pkg/util/testutils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"os"
	"testing"
)

var (
	cookieSeed = "0123456789abcdefghijklmnopqrstuv"
	cookieKey  = "2345asdYDS!2012L"
)

func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	_ = os.Setenv("TEMPLATE_DIR", fmt.Sprintf("%s/../../templates", cwd))
	_ = os.Setenv("STATIC_DIR", fmt.Sprintf("%s/../../static", cwd))
	_ = os.Setenv("ENV", "local")
	_ = os.Setenv("LOGLEVEL", "trace")
	logutils.ConfigureLogging()
	metrics.ConfigurePrometheusMetrics()
	code := m.Run()
	os.Exit(code)
}

func createDefaultHandlerFunc() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		hostname, _, _ := net.SplitHostPort(req.Host)
		_, _ = w.Write([]byte(fmt.Sprintf("hostname=%s\n", hostname)))
		_, _ = w.Write([]byte(fmt.Sprintf("path=%s\n", req.URL.Path)))
	}
}

func createDefaultHandler() http.Handler {
	return http.HandlerFunc(createDefaultHandlerFunc())
}

func TestMakeUnauthenticatedProxyRequest(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and request
	client := testutils.CreateHTTPClient(true)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/", proxyURL), nil)
	require.NoError(t, err)

	// Make request
	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)
	testutils.CheckResponseBody(t, res, "<title>Login</title>")
}

func TestMakeAuthenticatedProxyRequest(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and request
	client := testutils.CreateHTTPClient(true)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/test1234", proxyURL), nil)
	require.NoError(t, err)

	// Fake a logged-in session by attaching a session cookie to the request
	payload, _ := json.Marshal(session.Data{ID: "test", Name: "Test"})
	c, err := sm.MakeSessionCookie(cookieSeed, cookieKey, string(payload))
	require.NoError(t, err)
	req.AddCookie(c)

	// Make request
	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)
	testutils.CheckResponseBody(t, res, "path=/test1234")
}

func TestMetricsEndpoint(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.Use(metrics.CreatePrometheusHTTPMetricsHandler)
	pr.Handle("/metrics", promhttp.Handler())
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Fake a logged-in session by attaching a session cookie to the request
	payload, _ := json.Marshal(session.Data{ID: "test", Name: "Test"})
	c, err := sm.MakeSessionCookie(cookieSeed, cookieKey, string(payload))
	require.NoError(t, err)

	// Create client and loop some authenticated requests
	client := testutils.CreateHTTPClient(true)
	var loopRequest *http.Request
	for i := 1; i <= 10; i++ {
		loopRequest, err = http.NewRequest("GET", fmt.Sprintf("%s/test%d", proxyURL, i), nil)
		require.NoError(t, err)
		loopRequest.AddCookie(c)
		_, err = client.Do(loopRequest)
		require.NoError(t, err)
	}

	// Make metric request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/metrics", proxyURL), nil)
	require.NoError(t, err)
	req.AddCookie(c)
	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)

	// Parse metric data
	var m map[string]*dto.MetricFamily
	m, err = metrics.ParseMetricResponse(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)

	_, ok := m["http_requests_total"]
	require.True(t, ok)
	_, ok = m["http_request_duration_seconds"]
	require.True(t, ok)
	_, ok = m["http_request_length_bytes"]
	require.True(t, ok)
	_, ok = m["http_response_length_bytes"]
	require.True(t, ok)
}

func TestHealthzEndpoint(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.Handle("/healthz", healthz.Handler())
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and healthz request
	client := testutils.CreateHTTPClient(true)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/healthz", proxyURL), nil)
	require.NoError(t, err)
	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)
	testutils.CheckResponseBody(t, res, "OK")
}

func TestStaticEndpoint(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and static resource request
	client := testutils.CreateHTTPClient(true)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/favicon.png", proxyURL, proxy.staticPath), nil)
	require.NoError(t, err)
	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)
	testutils.CheckResponseContentType(t, res, "image/png")
}

func TestErrorEndpoint(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and static resource request
	errorMessage := "some unexpected error message"
	client := testutils.CreateHTTPClient(true)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", proxyURL, proxy.errorPath), nil)
	require.NoError(t, err)
	q := req.URL.Query()
	q.Add("error", util.Base64Encode([]byte(errorMessage)))
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)
	testutils.CheckResponseBody(t, res, errorMessage)
}

func TestLoginEndpoint(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and request
	client := testutils.CreateHTTPClient(true)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", proxyURL, proxy.loginPath), nil)
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)
	testutils.CheckResponseBody(t, res, "Forgot your password?")
}

func TestLogoutEndpoint(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and request
	client := testutils.CreateHTTPClient(false)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", proxyURL, proxy.logoutPath), nil)
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusFound)
}

func TestResetEndpoint(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and request
	client := testutils.CreateHTTPClient(true)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", proxyURL, proxy.resetPath), nil)
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)
	testutils.CheckResponseBody(t, res, "Reset your password")
}

func TestSignupEndpoint(t *testing.T) {
	// Create mock service
	sr := mux.NewRouter()
	sr.PathPrefix("/").HandlerFunc(createDefaultHandlerFunc())
	serverURL := testutils.StartTestServer(sr)

	// Create proxy
	provider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	proxy := New(serverURL, provider, sm)
	pr := mux.NewRouter()
	pr.PathPrefix("/").Handler(proxy)
	proxyURL := testutils.StartProxy(pr)

	// Create client and request
	client := testutils.CreateHTTPClient(true)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", proxyURL, proxy.signupPath), nil)
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)
	testutils.CheckResponseCode(t, res, http.StatusOK)
	testutils.CheckResponseBody(t, res, "Signup and become a ninja")
}
