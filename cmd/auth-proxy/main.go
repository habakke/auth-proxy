package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
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

func healthzHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
}

func main() {
	var port = flag.String("port", "8080", "The port number of the application")
	var target = flag.String("target", "", "The host to proxy requests towards")
	var token = flag.String("token", "", "The auth bearer token to add to requests")
	flag.Parse()

	addr := fmt.Sprintf(":%s", *port)

	p := &proxy.Proxy{
		Target: *target,
		Token:  *token,
	}
	lmw := util.LoggingMiddleware(log.Logger)
	lp := lmw(p)

	r := mux.NewRouter()
	r.Handle("/healthz", http.HandlerFunc(healthzHandler))
	r.Handle("/metrics", promhttp.Handler())
	r.PathPrefix("/").Handler(lp)

	log.Info().Msgf("Starting proxy server on %s", addr)
	log.Fatal().Err(http.ListenAndServe(addr, r))
}
