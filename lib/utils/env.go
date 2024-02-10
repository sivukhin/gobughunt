package utils

import (
	"os"
	"strconv"
	"time"

	"github.com/sivukhin/gobughunt/lib/logging"
)

func EnvMustParseDurationSec(key string) time.Duration {
	value := os.Getenv(key)
	seconds, err := strconv.Atoi(value)
	if err != nil {
		logging.Logger.Fatalf("unexpected format of duration env var: key=%v, value=%v, err=%v", key, value, err)
	}
	return time.Duration(seconds) * time.Second
}

func EnvMustParseString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		logging.Logger.Fatalf("empty string found for required env var: key=%v", key)
	}
	return value
}
