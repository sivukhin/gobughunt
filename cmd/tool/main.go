package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
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

	if len(os.Args) < 2 {
		fmt.Printf("usage: script.sql arg1 arg2 ... argn")
		os.Exit(1)
	}

	pgStorage, err := storage.NewPgStorage(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	script, err := os.ReadFile(os.Args[1])
	if err != nil {
		logging.Logger.Fatalf("failed to read SQL script from file %v: %v", os.Args[1], err)
	}
	args := make([]any, 0, len(os.Args)-2)
	for _, arg := range os.Args[2:] {
		args = append(args, arg)
	}
	queries := strings.Split(string(script), "--- DELIMITER ---")
	for i, query := range queries {
		rows, err := pgStorage.Query(signalsCtx, query, args...)
		if err != nil {
			logging.Logger.Errorf("script #%v failed: %v", i, err)
		} else {
			descriptions := rows.FieldDescriptions()
			columns := make([]any, 0, len(descriptions))
			for _, description := range descriptions {
				columns = append(columns, description.Name)
			}
			for rows.Next() {
				values, err := rows.Values()
				if err != nil {
					logging.Logger.Errorf("failed to extract values: %v", err)
					return
				}
				logging.Logger.Infof(strings.Repeat("%v\t", len(values)), values...)
			}
			err = rows.Err()
			if err != nil {
				logging.Logger.Errorf("script #%v failed: %v", i, err)
			} else {
				logging.Logger.Infof("script #%v succeeded", i)
			}
		}
	}
}
