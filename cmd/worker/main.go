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
		iterationDelay     = utils.EnvMustParseDurationSec("WORKER_ITERATION_DELAY_SEC")
		cleanupTimeout     = utils.EnvMustParseDurationSec("WORKER_CLEANUP_TIMEOUT_SEC")
		takeTimeout        = utils.EnvMustParseDurationSec("WORKER_TAKE_TIMEOUT_SEC")
		lintTimeout        = utils.EnvMustParseDurationSec("WORKER_LINT_TIMEOUT_SEC")
		updateTimeout      = utils.EnvMustParseDurationSec("WORKER_UPDATE_TIMEOUT_SEC")
		lockDuration       = utils.EnvMustParseDurationSec("WORKER_TASK_LOCK_DURATION_SEC")
		dockerMemoryGb     = utils.EnvMustParseInt("DOCKER_MEMORY_GB")
		dockerCpuMillis    = utils.EnvMustParseInt("DOCKER_CPU_MILLIS")
		dockerTempDir      = utils.EnvMustParseString("DOCKER_TEMP_DIR")
	)
	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()
	signalsCtx := timeout.SignalsCtx(syscall.SIGTERM, syscall.SIGKILL)

	pgStorage, err := storage.NewPgStorage(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	worker := lib.Worker{
		LintStorage: storage.PgLintStorage(pgStorage),
		DockerApi: lib.NaiveDockerApi{
			MemoryBytes: dockerMemoryGb * 1024 * 1024 * 1024,
			CpuMilli:    dockerCpuMillis,
			PidLimit:    16 * 1024,
		},
		Linting:        lib.NaiveLinting{TempDir: dockerTempDir, DockerApi: lib.Docker, GitApi: lib.Git},
		IterationDelay: iterationDelay,
		CleanupTimeout: cleanupTimeout,
		TakeTimeout:    takeTimeout,
		LintTimeout:    lintTimeout,
		UpdateTimeout:  updateTimeout,
		LockDuration:   lockDuration,
	}
	worker.RunForever(signalsCtx)
}
