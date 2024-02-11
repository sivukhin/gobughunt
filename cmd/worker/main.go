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
		connectionDuration    = utils.EnvMustParseDurationSec("CONNECTION_DURATION_SEC")
		connectionString      = utils.EnvMustParseString("CONNECTION_STRING")
		lockDuration          = utils.EnvMustParseDurationSec("WORKER_TASK_LOCK_DURATION_SEC")
		iterationTimeout      = utils.EnvMustParseDurationSec("WORKER_ITERATION_TIMEOUT_SEC")
		iterationFailDelay    = utils.EnvMustParseDurationSec("WORKER_ITERATION_FAIL_DELAY_SEC")
		iterationSuccessDelay = utils.EnvMustParseDurationSec("WORKER_ITERATION_SUCCESS_DELAY_SEC")
		dockerMemoryGb        = utils.EnvMustParseInt("DOCKER_MEMORY_GB")
		dockerCpuMillis       = utils.EnvMustParseInt("DOCKER_CPU_MILLIS")
		dockerTempDir         = utils.EnvMustParseString("DOCKER_TEMP_DIR")
	)
	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()
	signalsCtx := timeout.SignalsCtx(syscall.SIGTERM, syscall.SIGKILL)

	pgStorage, err := storage.NewPgStorage(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	worker := lib.Worker{
		DockerApi: lib.NaiveDockerApi{
			MemoryBytes: dockerMemoryGb * 1024 * 1024 * 1024,
			CpuNanos:    dockerCpuMillis * 1000 * 1000,
		},
		LintStorage:           storage.PgLintStorage(pgStorage),
		Linting:               lib.NaiveLinting{TempDir: dockerTempDir, DockerApi: lib.Docker, GitApi: lib.Git},
		IterationTimeout:      iterationTimeout,
		IterationFailDelay:    iterationFailDelay,
		IterationSuccessDelay: iterationSuccessDelay,
		LockDuration:          lockDuration,
	}
	worker.RunForever(signalsCtx)
}
