package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/habakke/auth-proxy/internal/auth/providers"
	"github.com/habakke/auth-proxy/internal/healthz"
	"github.com/habakke/auth-proxy/internal/metrics"
	"github.com/habakke/auth-proxy/internal/session"
	"github.com/habakke/auth-proxy/pkg/proxy"
	"github.com/habakke/auth-proxy/pkg/util/logutils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"
)

// Global variables
var (
	port       = os.Getenv("PORT")
	target     = os.Getenv("TARGET")
	token      = os.Getenv("TOKEN")
	cookieSeed = os.Getenv("COOKIE_SEED")
	cookieKey  = os.Getenv("COOKIE_KEY")
)

func init() {
	ConfigureMaxProcs()
	metrics.ConfigurePrometheusMetrics()
	logutils.ConfigureLogging()
}

func ConfigureMaxProcs() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func profileStart() {
	if len(os.Getenv("PROFILE")) == 0 {
		return
	}
	f, err := os.Create("cpuprofile")
	if err != nil {
		log.Error().AnErr("err", err).Msg("could not create CPU profile")
		os.Exit(1)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Error().AnErr("err", err).Msg("could not start CPU profile")
		os.Exit(1)
	}
}

func profileStop() {
	if len(os.Getenv("PROFILE")) == 0 {
		return
	}
	pprof.StopCPUProfile()
}

func main() {
	profileStart()
	defer profileStop()

	ctx := context.Background()

	addr := fmt.Sprintf(":%s", port)
	log.Info().Msgf("starting proxy server for %s on %s", target, addr)

	oauthProvider := providers.New("Google", &providers.ProviderData{})
	sm := session.NewManager(cookieSeed, cookieKey)
	p := proxy.New(
		target,
		oauthProvider,
		sm)

	p.AddBearingTokenToUpstreamRequests(token)

	r := mux.NewRouter()
	r.Use(metrics.CreatePrometheusHTTPMetricsHandler)
	r.Handle("/healthz", healthz.Handler())
	r.Handle("/metrics", promhttp.Handler())
	r.PathPrefix("/").Handler(p)

	srv := http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}
	go func() { log.Fatal().Err(srv.ListenAndServe()) }()
	_ = waitForSignal()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Info().Err(err).Msg("shutting down...")
	} else {
		log.Info().Msg("shutting down...")
	}
	os.Exit(0)
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}
