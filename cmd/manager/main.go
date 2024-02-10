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
		connectionDuration = utils.EnvMustParseDurationSec("CONNECTION_DURATION_SEC")
		connectionString   = utils.EnvMustParseString("CONNECTION_STRING")
		iterationTimeout   = utils.EnvMustParseDurationSec("MANAGER_ITERATION_TIMEOUT_SEC")
		iterationDelay     = utils.EnvMustParseDurationSec("MANAGER_ITERATION_DELAY_SEC")
	)
	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()
	signalsCtx := timeout.SignalsCtx(syscall.SIGTERM, syscall.SIGKILL)

	pgStorage, err := storage.NewPgStorage(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	manager := lib.Manager{
		LinterStorage:   storage.PgLinterStorage(pgStorage),
		RepoStorage:     storage.PgRepoStorage(pgStorage),
		LintStorage:     storage.PgLintStorage(pgStorage),
		DockerApi:       lib.NaiveDockerApi,
		GitApi:          lib.NaiveGitApi,
		ScheduleTimeout: iterationTimeout,
		ScheduleDelay:   iterationDelay,
	}
	manager.ManageForever(signalsCtx)
}
