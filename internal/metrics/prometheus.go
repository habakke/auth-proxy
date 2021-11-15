package metrics

import (
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"io"
	"net/http"
)

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

func ConfigurePrometheusMetrics() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestLength)
	prometheus.MustRegister(httpResponseLength)
}

func ParseMetricResponse(metrics io.Reader) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(metrics)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func CreatePrometheusHTTPMetricsHandler(next http.Handler) http.Handler {
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
