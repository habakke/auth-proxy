package proxy

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type Proxy struct {
	Target string
	Token  string
}

func removeAllHeaders(header http.Header) {
	for k, _ := range header {
		header.Del(k)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func appendAuthhorizationHeader(header http.Header, authorization string) {
	header.Set("Authorization", fmt.Sprintf("Bearer %s", authorization))
}

func (p *Proxy) getProxyURL() string {
	target := os.Getenv("TARGET")
	if target == "" {
		target = p.Target
	}
	return target
}

func (p *Proxy) getToken() string {
	token := os.Getenv("TOKEN")
	if token == "" {
		token = p.Token
	}
	return token
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, token string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	u, _ := url.Parse(target)

	// tweak request
	if token == "" {
		log.Fatal().Msg("Auth token is missing, exiting...")
	}
	appendAuthhorizationHeader(req.Header, token)
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	// create the reverse proxy
	proxy := httputil.ReverseProxy{Director: func(r *http.Request) {
		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.URL.Path = u.Path + r.URL.Path
		r.Host = u.Host
	}}
	proxy.ServeHTTP(res, req)
}

func (p *Proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	serveReverseProxy(p.getProxyURL(), p.getToken(), res, req)
}
