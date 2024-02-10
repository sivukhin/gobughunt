package main

import (
	"context"
	_ "embed"
	"syscall"

	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/storage"
	"github.com/sivukhin/gobughunt/lib/timeout"
	"github.com/sivukhin/gobughunt/lib/utils"
)

func main() {
	var (
		connectionDuration = utils.EnvMustParseDurationSec("CONNECTION_DURATION_SEC")
		connectionString   = utils.EnvMustParseString("CONNECTION_STRING")
	)
	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()
	signalsCtx := timeout.SignalsCtx(syscall.SIGTERM, syscall.SIGKILL)

	pgStorage, err := storage.NewPgStorage(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	if err := storage.PgLintStorage(pgStorage).InitTables(signalsCtx); err != nil {
		logging.Logger.Errorf("unable to init tables: %v", err)
	}
	if err := storage.PgLinterStorage(pgStorage).InitTables(signalsCtx); err != nil {
		logging.Logger.Errorf("unable to init tables: %v", err)
	}
	if err := storage.PgRepoStorage(pgStorage).InitTables(signalsCtx); err != nil {
		logging.Logger.Errorf("unable to init tables: %v", err)
	}
}
