package testutils

import (
	"crypto/tls"
	"fmt"
	"github.com/habakke/auth-proxy/pkg/util/logutils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

var (
	ServerAddr string
	ServerHost string
	ServerPort int

	ProxyServerAddr string
	ProxyServerHost string
	ProxyServerPort int
)

func StartTestServer(mux http.Handler) string {
	server := httptest.NewServer(mux)
	ServerAddr = server.Listener.Addr().String()
	h, p, err := net.SplitHostPort(ServerAddr)
	if err != nil {
		log.Fatal().Msgf("failed to parse ServerAddr")
	}
	ServerHost = h
	ServerPort, err = strconv.Atoi(p)
	if err != nil {
		log.Fatal().Msgf("failed to parse server port")
	}
	log.Info().Msgf("Server available at http://%s:%d", ServerHost, ServerPort)
	return fmt.Sprintf("http://%s", ServerAddr)
}

func StartProxy(mux http.Handler) string {
	proxyServer := httptest.NewServer(mux)
	ProxyServerAddr = proxyServer.Listener.Addr().String()
	h, p, err := net.SplitHostPort(ProxyServerAddr)
	if err != nil {
		log.Fatal().Msgf("failed to parse ServerAddr")
	}
	ProxyServerHost = h
	ProxyServerPort, err = strconv.Atoi(p)
	if err != nil {
		log.Fatal().Msgf("failed to parse server port")
	}
	log.Info().Msgf("Proxy available at http://%s:%d", ProxyServerHost, ProxyServerPort)
	return fmt.Sprintf("http://%s", ProxyServerAddr)
}

func CreateHTTPClient(followRedirects bool) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		//#nosec
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: logutils.NewLoggingRoundTripper(transport),
		Timeout:   5 * time.Second,
	}

	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func CheckResponseCode(t *testing.T, res *http.Response, code int) {
	if res.StatusCode != code {
		t.Errorf("response returned incorrect status code: got %v want %v", res.StatusCode, code)
	}
}

func CheckResponseContentType(t *testing.T, res *http.Response, contentType string) {
	c := res.Header.Get("Content-Type")
	if c != contentType {
		t.Errorf("response returned incorrect content type: got %v want %v", c, contentType)
	}
}

func CheckResponseBody(t *testing.T, res *http.Response, keyword string) {
	body, err := io.ReadAll(res.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), keyword)
}
