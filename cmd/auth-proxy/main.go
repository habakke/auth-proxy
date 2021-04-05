package main

import (
	"flag"
	"fmt"
	"github.com/habakke/auth-proxy/pkg/util"
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
	var port = flag.String("port", "8080", "The port number of the application")
	var target = flag.String("target", "", "The host to proxy requests towards")
	var token = flag.String("token", "", "The auth bearer token to add to requests")
	flag.Parse()

	p := &proxy.Proxy{
		Target: *target,
		Token:  *token,
	}
	lmw := util.LoggingMiddleware(log.Logger)
	lp := lmw(p)

	log.Info().Str("interface", "0.0.0.0").Str("port", *port).Msg("Starting proxy server")
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", *port), lp); err != nil {
		log.Fatal().Msgf("ListenAndServe: %v", err)
	}
}
