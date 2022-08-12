package proxy

import (
	"fmt"
	"github.com/habakke/auth-proxy/internal/auth"
	"github.com/habakke/auth-proxy/internal/auth/providers"
	"github.com/habakke/auth-proxy/internal/cookie"
	"github.com/habakke/auth-proxy/internal/session"
	"github.com/habakke/auth-proxy/pkg/util"
	"github.com/habakke/auth-proxy/pkg/util/logutils"
	"github.com/rs/zerolog/log"
	"html/template"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

type Proxy struct {
	Target string

	headers           map[string]string
	authorizedHeaders map[string]string

	localAuth *auth.LocalAuth
	provider  providers.Provider

	pathWhiteList   []*regexp.Regexp
	domainWhiteList []*regexp.Regexp
	errorPath       string
	loginPath       string
	logoutPath      string
	staticPath      string
	resetPath       string
	signupPath      string

	sessionManager *session.Manager
}

func NewProxy(target string, provider providers.Provider, sessionManager *session.Manager) *Proxy {
	return &Proxy{
		Target:            target,
		headers:           make(map[string]string),
		authorizedHeaders: make(map[string]string),
		provider:          provider,
		localAuth:         auth.NewAuthLocal(),
		errorPath:         "/auth/error",
		loginPath:         "/auth/login",
		logoutPath:        "/auth/logout",
		resetPath:         "/auth/reset",
		signupPath:        "/auth/signup",
		staticPath:        "/static",

		sessionManager: sessionManager,
	}
}

func (p *Proxy) SetErrorPath(errorPath string) {
	p.errorPath = errorPath
}

func (p *Proxy) SetLoginPath(loginPath string) {
	p.errorPath = loginPath
}

func (p *Proxy) SetLogoutPath(logoutPath string) {
	p.errorPath = logoutPath
}

func (p *Proxy) SetStaticPath(staticPath string) {
	p.errorPath = staticPath
}

func (p *Proxy) SetLocalAuth(localAuth *auth.LocalAuth) {
	p.localAuth = localAuth
}

func (p *Proxy) AddHeaderToUpstreamRequests(key string, value string) {
	p.headers[key] = value
}

func (p *Proxy) AddAuthenticatedHeaderToUpstreamRequests(key string, value string) {
	p.authorizedHeaders[key] = value
}

func (p *Proxy) AddBearingTokenToUpstreamRequests(token string) {
	p.AddAuthenticatedHeaderToUpstreamRequests("authorization", fmt.Sprintf("Bearer %s", token))
}

func (p *Proxy) AddXForwardedForToRequests(host string) {
	p.AddHeaderToUpstreamRequests("x-forwarded-for", host)
}

func (p *Proxy) getProxyURL() string {
	target := os.Getenv("TARGET")
	if target == "" {
		target = p.Target
	}
	return target
}

func (p *Proxy) Authenticate(req *http.Request) bool {
	s, err := p.sessionManager.ReadSession(req)
	if err != nil {
		return false
	}

	return p.provider.AuthenticateSession(s)
}

func (p Proxy) IsWhitelistRequest(req *http.Request) bool {
	return req.Method == "OPTIONS" || p.IsWhitelistedPath(req.URL.Path) || p.IsWhitelistedDomain(req.URL.Host)
}

func (p Proxy) IsWhitelistedPath(path string) bool {
	for _, u := range p.pathWhiteList {
		ok := u.MatchString(path)
		return ok
	}
	return false
}

func (p Proxy) IsWhitelistedDomain(domain string) bool {
	for _, d := range p.domainWhiteList {
		ok := d.MatchString(domain)
		return ok
	}
	return false
}

// handle error and redirect to error page
func errorHandler(res http.ResponseWriter, req *http.Request, errMsg string) {
	http.Redirect(res, req, fmt.Sprintf("/auth/error?error=%s", util.Base64Encode([]byte(errMsg))), http.StatusTemporaryRedirect)
}

// Serve a reverse proxy for a given url
func (p *Proxy) serveReverseProxy(target string, authenticated bool, res http.ResponseWriter, req *http.Request) {
	// parse the url
	u, _ := url.Parse(target)

	for k, v := range p.headers {
		req.Header.Add(k, v)
	}

	if authenticated {
		for k, v := range p.authorizedHeaders {
			req.Header.Add(k, v)
		}
	}

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		req.Header.Add("x-forwarded-for", clientIP)
	}

	// create the reverse proxy
	proxy := httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = u.Scheme
			r.URL.Host = u.Host
			r.URL.Path = u.Path + r.URL.Path
			r.Host = u.Host
		},
		Transport: logutils.NewLoggingRoundTripper(http.DefaultTransport),
	}
	proxy.ServeHTTP(res, req)
}

func (p *Proxy) LocalAuth(req *http.Request) (providers.User, bool) {
	if req.Method != "POST" || p.localAuth == nil {
		return nil, false
	}

	username := req.FormValue("username")
	password := req.FormValue("password")
	if username == "" || password == "" {
		return nil, false
	}

	return p.localAuth.Authenticate(username, password)
}

func (p *Proxy) Login(res http.ResponseWriter, req *http.Request) {
	// First try local authentication
	user, ok := p.LocalAuth(req)
	if ok {
		sd := session.Data{
			ID:         user.GetID(),
			Authorized: false,
		}
		_ = p.sessionManager.AttachSession(res, sd)
		http.Redirect(res, req, "/", http.StatusFound)
	}

	// Start Provider Oauth2 authentication
	u, err := p.provider.GetProviderLoginURL(res)
	if err != nil {
		errorHandler(res, req, "failed to generate Oauth2 authentication link")
		return
	}

	http.Redirect(res, req, u.String(), http.StatusFound)
}

func (p *Proxy) OauthCallback(res http.ResponseWriter, req *http.Request) {
	csrf, _ := req.Cookie(cookie.CSRFCookieName)

	// Do some sanity checking
	err := req.ParseForm()
	if err != nil {
		errorHandler(res, req, fmt.Sprintf("Internal error,  %s", err.Error()))
		return
	}
	errorString := req.Form.Get("error")
	if errorString != "" {
		errorHandler(res, req, fmt.Sprintf("Permission denied: %s", errorString))
		return
	}
	if req.FormValue("state") != csrf.Value {
		errMsg := "invalid csrf state"
		errorHandler(res, req, errMsg)
		return
	}

	// Exchange auth code for access/refresh token pair
	err = p.provider.Exchange(req.FormValue("code"))
	if err != nil {
		errMsg := fmt.Sprintf("failed to exchange authorization code with %s", p.provider.Data().Name)
		errorHandler(res, req, errMsg)
		return
	}

	// Get userinfo from provider
	user, err := p.provider.GetUser()
	if err != nil {
		errMsg := "failed to get userdata from Oauth provider"
		errorHandler(res, req, errMsg)
		return
	}

	log.Debug().Str("id", user.GetID()).Str("user", user.GetUsername()).Msg("user logged in")

	// Set session data
	s := session.Data{
		ID:         user.GetID(),
		Name:       user.GetName(),
		Email:      user.GetEmail(),
		Authorized: false,
	}
	_ = p.sessionManager.AttachSession(res, s)
	http.Redirect(res, req, "/?", http.StatusFound)
}

func disableCaching(res http.ResponseWriter) {
	res.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	res.Header().Set("Pragma", "no-cache")
	res.Header().Set("Expires", "0")
}

func (p Proxy) getTemplate(name string) *template.Template {
	tDir := os.Getenv("TEMPLATE_DIR")
	if tDir == "" {
		log.Fatal().Msg("The TEMPLATE_DIR environmental variable cannot be empty")
	}
	t, err := template.New(name).ParseFiles(path.Join(path.Clean(tDir), name))
	if err != nil {
		log.Fatal().AnErr("err", err).Msgf("failed parsing template %s", name)
	}
	return t
}

func (p Proxy) ErrorPage(res http.ResponseWriter, req *http.Request) {
	disableCaching(res)

	msg := "No error message found"
	if e, ok := req.URL.Query()["error"]; ok {
		if b, err := util.Base64Decode(e[0]); err != nil {
			log.Debug().AnErr("err", err).Msg("failed to decode error message")
		} else {
			msg = string(b)
		}
	}

	name := "error.tpl"
	data := struct {
		ErrorMessage string
		StaticPath   string
		HomePageURL  string
		ContactEmail string
	}{
		ErrorMessage: msg,
		StaticPath:   p.staticPath,
		HomePageURL:  os.Getenv("HOMEPAGE_URL"),
		ContactEmail: os.Getenv("CONTACT_EMAIL"),
	}
	_ = p.getTemplate(name).ExecuteTemplate(res, name, data)
}

func (p *Proxy) LoginPage(res http.ResponseWriter, req *http.Request) {
	p.sessionManager.RemoveSession(res)
	disableCaching(res)

	name := "login.tpl"
	data := struct {
		LoginPath         string
		ProviderLoginPath string
		StaticPath        string
	}{
		LoginPath:         p.loginPath,
		ProviderLoginPath: p.provider.GetLoginPath(),
		StaticPath:        p.staticPath,
	}
	_ = p.getTemplate(name).ExecuteTemplate(res, name, data)
}

func (p *Proxy) ResetPage(res http.ResponseWriter, req *http.Request) {
	p.sessionManager.RemoveSession(res)
	disableCaching(res)

	name := "reset.tpl"
	data := struct {
		StaticPath string
	}{
		StaticPath: p.staticPath,
	}
	_ = p.getTemplate(name).ExecuteTemplate(res, name, data)
}

func (p *Proxy) SignupPage(res http.ResponseWriter, req *http.Request) {
	p.sessionManager.RemoveSession(res)
	disableCaching(res)

	name := "signup.tpl"
	data := struct {
		StaticPath  string
		HomePageURL string
	}{
		StaticPath:  p.staticPath,
		HomePageURL: os.Getenv("HOMEPAGE_URL"),
	}
	_ = p.getTemplate(name).ExecuteTemplate(res, name, data)
}

func (p *Proxy) StaticFolder(res http.ResponseWriter, req *http.Request) {
	sDir := os.Getenv("STATIC_DIR")
	if sDir == "" {
		log.Fatal().Msg("The STATIC_DIR environmental variable cannot be empty")
	}
	http.StripPrefix(p.staticPath, http.FileServer(http.Dir(path.Clean(sDir)))).ServeHTTP(res, req)
}

func (p *Proxy) Logout(res http.ResponseWriter, req *http.Request) {
	p.sessionManager.RemoveSession(res)
	http.Redirect(res, req, "/", http.StatusFound)
}

func (p *Proxy) Proxy(res http.ResponseWriter, req *http.Request) {
	if !p.Authenticate(req) {
		p.sessionManager.RemoveSession(res)
		http.Redirect(res, req, fmt.Sprintf("%s?p=%s", p.loginPath, req.URL.Path), http.StatusFound)
	} else {
		p.serveReverseProxy(p.getProxyURL(), true, res, req)
	}
}

func (p *Proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	switch cleanPath := strings.TrimSuffix(req.URL.Path, "/"); {
	case cleanPath == p.errorPath && req.Method == "GET":
		p.ErrorPage(res, req)
	case cleanPath == p.resetPath && req.Method == "GET":
		p.ResetPage(res, req)
	case cleanPath == p.signupPath && req.Method == "GET":
		p.SignupPage(res, req)
	case cleanPath == p.loginPath && req.Method == "GET":
		p.LoginPage(res, req)
	case cleanPath == p.loginPath && req.Method == "POST":
		p.Login(res, req)
	case strings.HasPrefix(cleanPath, p.staticPath) && req.Method == "GET":
		p.StaticFolder(res, req)
	case cleanPath == p.provider.GetLoginPath():
		p.Login(res, req)
	case cleanPath == p.logoutPath:
		p.Logout(res, req)
	case p.IsWhitelistRequest(req):
		p.serveReverseProxy(p.getProxyURL(), true, res, req)
	case cleanPath == p.provider.GetCallbackPath():
		p.OauthCallback(res, req)
	default:
		p.Proxy(res, req)
	}
}
