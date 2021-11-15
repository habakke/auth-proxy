package logutils

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ConfigureLogging() {
	var logging = getEnv("LOGLEVEL", zerolog.InfoLevel.String())
	l, err := zerolog.ParseLevel(logging)
	if err != nil {
		l = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(l)
	if env, _ := os.LookupEnv("ENV"); env == "local" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	}
	log.Logger = log.Logger.With().Caller().Logger()
}
