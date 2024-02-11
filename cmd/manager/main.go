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
		fetchTimeout        = utils.EnvMustParseDurationSec("MANAGER_FETCH_TIMEOUT_SEC")
		refreshTimeout      = utils.EnvMustParseDurationSec("MANAGER_REFRESH_TIMEOUT_SEC")
		scheduleTimeout     = utils.EnvMustParseDurationSec("MANAGER_SCHEDULE_TIMEOUT_SEC")
		managerFailDelay    = utils.EnvMustParseDurationSec("MANAGER_FAIL_DELAY_SEC")
		managerSuccessDelay = utils.EnvMustParseDurationSec("MANAGER_SUCCESS_DELAY_SEC")
	)
	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()
	signalsCtx := timeout.SignalsCtx(syscall.SIGTERM, syscall.SIGKILL)

	pgStorage, err := storage.NewPgStorage(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	manager := lib.Manager{
		LinterStorage:       storage.PgLinterStorage(pgStorage),
		RepoStorage:         storage.PgRepoStorage(pgStorage),
		LintStorage:         storage.PgLintStorage(pgStorage),
		DockerApi:           lib.Docker,
		GitApi:              lib.Git,
		FetchTimeout:        fetchTimeout,
		RefreshTimeout:      refreshTimeout,
		ScheduleTimeout:     scheduleTimeout,
		ManagerFailDelay:    managerFailDelay,
		ManagerSuccessDelay: managerSuccessDelay,
	}
	manager.ManageForever(signalsCtx)
}
