package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/habakke/auth-proxy/internal/healthz"
	"github.com/habakke/auth-proxy/pkg/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/habakke/auth-proxy/pkg/proxy"
)

func init() {
	ConfigureMaxProcs()
	ConfigureLogging()
}

func ConfigureMaxProcs() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func ConfigureLogging() {
	if env, _ := os.LookupEnv("ENV"); env == "local" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	}
	log.Logger = log.Logger.With().Caller().Logger()
}

func main() {
	var port = os.Getenv("PORT")
	var target = os.Getenv("TARGET")
	var token = os.Getenv("TOKEN")
	var logging = os.Getenv("LOGGING") != ""

	addr := fmt.Sprintf(":%s", port)

	p := &proxy.Proxy{
		Target: target,
		Token:  token,
	}
	lmw := util.LoggingMiddleware(log.Logger)
	lp := lmw(p)

	r := mux.NewRouter()
	r.Handle("/healthz", healthz.Handler())
	r.Handle("/metrics", promhttp.Handler())

	if logging {
		r.PathPrefix("/").Handler(lp)
	} else {
		r.PathPrefix("/").Handler(p)
	}

	log.Info().Msgf("Starting proxy server on %s", addr)
	log.Fatal().Err(http.ListenAndServe(addr, r))
}
