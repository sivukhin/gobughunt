package main

import (
	"context"
	"syscall"

	"github.com/sivukhin/gobughunt/lib"
	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/storage"
	"github.com/sivukhin/gobughunt/lib/timeout"
	"github.com/sivukhin/gobughunt/lib/utils"
)

func main() {
	var (
		connectionDuration  = utils.EnvMustParseDurationSec("CONNECTION_DURATION_SEC")
		connectionString    = utils.EnvMustParseString("CONNECTION_STRING")
		lockDuration        = utils.EnvMustParseDurationSec("WORKER_TASK_LOCK_DURATION_SEC")
		iterationTimeout    = utils.EnvMustParseDurationSec("WORKER_ITERATION_TIMEOUT_SEC")
		iterationShortDelay = utils.EnvMustParseDurationSec("WORKER_ITERATION_SHORT_DELAY_SEC")
		iterationLongDelay  = utils.EnvMustParseDurationSec("WORKER_ITERATION_LONG_DELAY_SEC")
	)
	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()
	signalsCtx := timeout.SignalsCtx(syscall.SIGTERM, syscall.SIGKILL)

	pgStorage, err := storage.NewPgStorage(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	worker := lib.Worker{
		LintStorage:         storage.PgLintStorage(pgStorage),
		Linting:             lib.NaiveLinting,
		IterationTimeout:    iterationTimeout,
		IterationShortDelay: iterationShortDelay,
		IterationLongDelay:  iterationLongDelay,
		LockDuration:        lockDuration,
	}
	worker.RunForever(signalsCtx)
}
