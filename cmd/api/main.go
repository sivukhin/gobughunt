package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/storage"
	"github.com/sivukhin/gobughunt/lib/utils"
)

func wrap[T any](handle func(request *http.Request) (T, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		result, err := handle(request)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			writer.WriteHeader(http.StatusInsufficientStorage)
			_, _ = writer.Write([]byte(err.Error()))
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(response)
	}
}

func main() {
	var (
		connectionDuration = utils.EnvMustParseDurationSec("CONNECTION_DURATION_SEC")
		connectionString   = utils.EnvMustParseString("CONNECTION_STRING")
	)
	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()

	pgStorage, err := storage.NewPgStorage(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	bugHuntStorage := storage.PgBugHuntStorage(pgStorage)

	server := http.NewServeMux()
	server.Handle("/", http.FileServer(http.Dir("./static")))
	server.HandleFunc("/api/linters", wrap(func(request *http.Request) ([]storage.LinterDto, error) {
		return bugHuntStorage.Linters(request.Context())
	}))
	server.HandleFunc("/api/repos", wrap(func(request *http.Request) ([]storage.RepoDto, error) {
		return bugHuntStorage.Repos(request.Context())
	}))
	server.HandleFunc("/api/lint-tasks", wrap(func(request *http.Request) ([]storage.LintTaskDto, error) {
		params := request.URL.Query()
		skip, err := strconv.Atoi(params.Get("skip"))
		if err != nil {
			return nil, err
		}
		take, err := strconv.Atoi(params.Get("take"))
		if err != nil {
			return nil, err
		}
		take = max(1, min(take, skip+100))
		return bugHuntStorage.LintTasks(request.Context(), skip, take)
	}))
	server.Handle("/api/lint-highlights", wrap(func(request *http.Request) ([]storage.LintHighlightDto, error) {
		params := request.URL.Query()
		lintId := params.Get("lintId")
		if lintId == "" {
			return nil, fmt.Errorf("lintId required")
		}
		return bugHuntStorage.LintHighlights(request.Context(), lintId)
	}))

	err = http.ListenAndServe(":3000", server)
	if err != nil {
		logging.Logger.Errorf("exited server: %v", err)
	}
}
