package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/storage"
	"github.com/sivukhin/gobughunt/lib/utils"
)

var (
	connectionDuration = utils.EnvMustParseDurationSec("CONNECTION_DURATION_SEC")
	connectionString   = utils.EnvMustParseString("CONNECTION_STRING")
	serverLocal        = utils.EnvTryParseBool("SERVER_LOCAL")
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
		if serverLocal {
			writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5000")
			writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(response)
	}
}

func main() {
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
		repoId := params.Get("repoId")
		linterId := params.Get("linterId")
		if lintId == "" && repoId == "" && linterId == "" {
			return nil, fmt.Errorf("one of three parameters should be set: lintId, repoId, linterId")
		}
		return bugHuntStorage.LintHighlights(request.Context(), storage.LintHighlightsFilter{LintId: lintId, RepoId: repoId, LinterId: linterId})
	}))
	server.Handle("/api/lint-highlights/moderate", wrap(func(request *http.Request) (struct{}, error) {
		params := request.URL.Query()
		lintId := params.Get("lintId")
		if lintId == "" {
			return struct{}{}, fmt.Errorf("lintId required")
		}
		path := params.Get("path")
		if path == "" {
			return struct{}{}, fmt.Errorf("path required")
		}
		startLineString := params.Get("startLine")
		if startLineString == "" {
			return struct{}{}, fmt.Errorf("startLine required")
		}
		startLine, err := strconv.Atoi(startLineString)
		if err != nil {
			return struct{}{}, err
		}
		endLineString := params.Get("endLine")
		if endLineString == "" {
			return struct{}{}, fmt.Errorf("endLine required")
		}
		endLine, err := strconv.Atoi(endLineString)
		if err != nil {
			return struct{}{}, err
		}
		status := params.Get("status")
		if status != "accepted" && status != "rejected" {
			return struct{}{}, fmt.Errorf("unexpected status: %v", status)
		}
		highlight := dto.LintHighlight{
			Path:      path,
			StartLine: startLine,
			EndLine:   endLine,
		}
		return struct{}{}, bugHuntStorage.ModerateHighlight(request.Context(), lintId, highlight, status)
	}))

	err = http.ListenAndServe(":3000", server)
	if err != nil {
		logging.Logger.Errorf("exited server: %v", err)
	}
}
