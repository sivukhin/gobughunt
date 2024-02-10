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
		scheduleTimeout    = utils.EnvMustParseDurationSec("MANAGER_SCHEDULE_TIMEOUT_SEC")
		scheduleDelay      = utils.EnvMustParseDurationSec("MANAGER_SCHEDULE_DELAY_SEC")
		shortDelay         = utils.EnvMustParseDurationSec("MANAGER_SHORT_DELAY_SEC")
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
		ScheduleTimeout: scheduleTimeout,
		ScheduleDelay:   scheduleDelay,
		ShortDelay:      shortDelay,
	}
	manager.ManageForever(signalsCtx)
}
