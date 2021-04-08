package main

import (
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/habakke/auth-proxy/internal/auth"
	"github.com/habakke/auth-proxy/internal/healthz"
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

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}

func main() {
	profileStart()
	defer profileStop()

	var port = os.Getenv("PORT")
	var target = os.Getenv("TARGET")
	var token = os.Getenv("TOKEN")

	addr := fmt.Sprintf(":%s", port)
	log.Info().Msgf("Starting proxy server for %s on %s", target, addr)
	defer log.Info().Msg("shutting down...")

	p := &proxy.Proxy{
		Target: target,
		Token:  token,
	}

	lmw := util.LoggingMiddleware(log.Logger)
	o := auth.NewGoogleOauth2(token)
	a := auth.NewAuthLocal(token)

	r := mux.NewRouter()
	r.Use(basicPromMetricsHandler)
	r.PathPrefix("/auth/login").Methods("GET").Handler(lmw(http.FileServer(http.Dir("./static"))))
	r.PathPrefix("/auth/login").Methods("POST").Handler(lmw(http.HandlerFunc(a.LoginHandler)))
	r.Handle("/auth/google/login", lmw(http.HandlerFunc(o.LoginHandler)))
	r.Handle("/auth/google/callback", lmw(http.HandlerFunc(o.CallbackHandler)))
	r.Handle("/healthz", healthz.Handler())
	r.Handle("/metrics", promhttp.Handler())
	r.PathPrefix("/").Handler(lmw(p))

	srv := http.Server{Addr: addr, Handler: r}
	go func() { log.Fatal().Err(srv.ListenAndServe()) }()
	_ = waitForSignal()
}
