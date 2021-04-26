package main

import (
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/habakke/auth-proxy/internal/auth/providers"
	"github.com/habakke/auth-proxy/internal/healthz"
	"github.com/habakke/auth-proxy/internal/session"
	"github.com/habakke/auth-proxy/pkg/proxy"
	"github.com/habakke/auth-proxy/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
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
	psb = prometheus.ExponentialBuckets(1, 10, 6)

	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	}, []string{"path", "method", "code"})

	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of all HTTP requests",
	}, []string{"path", "method"})

	httpRequestLength = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_length_bytes",
		Help:    "Length of all HTTP requests",
		Buckets: psb,
	}, []string{"path", "method"})

	httpResponseLength = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_length_bytes",
		Help:    "Length of all HTTP responses",
		Buckets: psb,
	}, []string{"path", "method"})
)

func init() {
	ConfigurePrometheus()
	ConfigureMaxProcs()
	ConfigureLogging()
}

func basicPromMetricsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		route := mux.CurrentRoute(req)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(path, req.Method))
		m := httpsnoop.CaptureMetrics(next, res, req)
		timer.ObserveDuration()
		httpRequestsTotal.WithLabelValues(path, req.Method, fmt.Sprint(m.Code)).Inc()
		httpRequestLength.WithLabelValues(path, req.Method).Observe(float64(req.ContentLength))
		httpResponseLength.WithLabelValues(path, req.Method).Observe(float64(m.Written))
	})
}

func ConfigurePrometheus() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestLength)
	prometheus.MustRegister(httpResponseLength)
}

func ConfigureMaxProcs() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func ConfigureLogging() {
	var logging = os.Getenv("LOGLEVEL")
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

	var port = os.Getenv("PORT")
	var target = os.Getenv("TARGET")
	var token = os.Getenv("TOKEN")
	var cookieSeed = os.Getenv("COOKIE_SEED")
	var cookieKey = os.Getenv("COOKIE_KEY")

	addr := fmt.Sprintf(":%s", port)
	log.Info().Msgf("starting proxy server for %s on %s", target, addr)
	defer log.Info().Msg("shutting down...")

	oauthProvider := providers.New("Google", &providers.ProviderData{})
	lmw := util.NewLoggingMiddleware(log.Logger)
	sm := session.NewManager(cookieSeed, cookieKey)
	p := proxy.New(
		target,
		oauthProvider,
		sm)

	p.AddBearingTokenToUpstreamRequests(token)

	r := mux.NewRouter()
	r.Use(basicPromMetricsHandler)
	r.Handle("/healthz", healthz.Handler())
	r.Handle("/metrics", promhttp.Handler())
	r.PathPrefix("/").Handler(lmw(p))

	srv := http.Server{Addr: addr, Handler: r}
	go func() { log.Fatal().Err(srv.ListenAndServe()) }()
	_ = waitForSignal()
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}
