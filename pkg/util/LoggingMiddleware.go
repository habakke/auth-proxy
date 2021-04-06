package util

import (
	"bytes"
	"github.com/rs/zerolog"
	"net/http"
	"runtime/debug"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true

	return
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.body.Write(data)
	return rw.ResponseWriter.Write(data)
}

func errorHandler(logger zerolog.Logger, w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error().
			Err(err.(error)).
			Bytes("trace", debug.Stack()).
			Msg("Internal Server Error")
	}
}

func LoggingMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer errorHandler(logger, w, r)

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			defer wrapped.body.Reset()
			next.ServeHTTP(wrapped, r)
			/*
				log.Print("=== REQUEST HEADERS ===")
				for k, v := range r.Header {
					log.Print(fmt.Sprintf(" << %s: %s", k, v))
				}

				log.Print("=== RESPONSE HEADERS ===")
				for k, v := range wrapped.Header() {
					log.Print(fmt.Sprintf(" << %s: %s", k, v))
				}
				logger.Print(fmt.Sprintf("=== BODY ===\n%s", string(wrapped.body.Bytes())))
			*/
			logger.Debug().
				Int("status", wrapped.status).
				Dur("duration", time.Since(start)).
				Str("host", r.Host).
				Str("path", r.URL.EscapedPath()).
				Str("method", r.Method).
				Str("params", r.URL.RawQuery).
				Msg("Request")
		}

		return http.HandlerFunc(fn)
	}
}
