package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/habakke/auth-proxy/internal/auth"
	"github.com/habakke/auth-proxy/internal/healthz"
	"github.com/habakke/auth-proxy/pkg/proxy"
	"github.com/habakke/auth-proxy/pkg/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"runtime"
	"time"
)

func init() {
	ConfigureMaxProcs()
	ConfigureLogging()
}

func ConfigureMaxProcs() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func ConfigureLogging() {
	var logging = os.Getenv("LOGGING")
	l, err := zerolog.ParseLevel(logging)
	if err != nil {
		l = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(l)

	if env, _ := os.LookupEnv("ENV"); env == "local" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	}
	log.Logger = log.Logger.With().Caller().Logger()
	log.Info().Msgf("Setting default loglevel to %s", l.String())
}

func main() {
	var port = os.Getenv("PORT")
	var target = os.Getenv("TARGET")
	var token = os.Getenv("TOKEN")

	addr := fmt.Sprintf(":%s", port)

	p := &proxy.Proxy{
		Target: target,
		Token:  token,
	}

	lmw := util.LoggingMiddleware(log.Logger)
	o := auth.NewGoogleOauth2(token)
	a := auth.NewAuthLocal(token)

	r := mux.NewRouter()
	r.PathPrefix("/auth/login").Methods("GET").Handler(lmw(http.FileServer(http.Dir("./static"))))
	r.PathPrefix("/auth/login").Methods("POST").Handler(lmw(http.HandlerFunc(a.LoginHandler)))
	r.Handle("/auth/google/login", lmw(http.HandlerFunc(o.LoginHandler)))
	r.Handle("/auth/google/callback", lmw(http.HandlerFunc(o.CallbackHandler)))
	r.Handle("/healthz", healthz.Handler())
	r.Handle("/metrics", promhttp.Handler())
	r.PathPrefix("/").Handler(lmw(p))

	log.Info().Msgf("Starting proxy server for %s on %s", target, addr)
	log.Fatal().Err(http.ListenAndServe(addr, r))
}
