package utils

import (
	"github.com/sivukhin/gobughunt/lib/logging"
)

func Must[T any](value T, err error) T {
	if err != nil {
		logging.Logger.Fatal(err)
	}
	return value
}
