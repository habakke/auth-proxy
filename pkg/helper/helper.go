package helper

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"runtime"
	"strconv"
)

func IsEnvSet(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func GetStringEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	return "", errors.New(fmt.Sprintf("environment variable %s not set", key))
}

func GetStringEnvWithDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetIntEnv(key string) (int, error) {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err != nil {
			return 0, err
		} else {
			return i, nil
		}
	}
	return 0, errors.New(fmt.Sprintf("environment variable %s not set", key))
}

func GetIntEnvWithDefault(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err != nil {
			return fallback
		} else {
			return i
		}
	}
	return fallback
}

func HandleError(err error, fatal bool, msg string, args ...interface{}) string {
	if err != nil {
		pc, filename, line, _ := runtime.Caller(1)
		logMessage := fmt.Sprintf("%s[%s:%d] %s", runtime.FuncForPC(pc).Name(), filename, line, fmt.Sprintf(msg, args...))
		if fatal {
			log.Fatal().Msg(logMessage)
		} else {
			log.Error().Msg(logMessage)
		}
		return logMessage
	}
	return ""
}
