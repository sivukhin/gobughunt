package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"

	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/utils"
	"github.com/sivukhin/gobughunt/storage"
	"github.com/sivukhin/gobughunt/storage/db"

	_ "golang.org/x/oauth2/endpoints"
)

var (
	connectionDuration       = utils.EnvMustParseDurationSec("CONNECTION_DURATION_SEC")
	connectionString         = utils.EnvMustParseString("CONNECTION_STRING")
	serverLocal              = utils.EnvTryParseBool("SERVER_LOCAL")
	serverUser               = utils.EnvMustParseString("SERVER_USER")
	serverPass               = utils.EnvMustParseString("SERVER_PASS")
	serverListenAddr         = utils.EnvMustParseString("SERVER_LISTEN_ADDR")
	githubOauth2ClientId     = utils.EnvMustParseString("GITHUB_OAUTH_CLIENT_ID")
	githubOauth2ClientSecret = utils.EnvMustParseString("GITHUB_OAUTH_CLIENT_SECRET")
)

var (
	githubOauthConfig = &oauth2.Config{
		ClientID:     githubOauth2ClientId,
		ClientSecret: githubOauth2ClientSecret,
		Scopes:       []string{"(no scope)"},
		Endpoint:     endpoints.GitHub,
	}
)

func basicAuth(writer http.ResponseWriter, request *http.Request) bool {
	if serverLocal {
		return true
	}
	verifier := oauth2.GenerateVerifier()
	redirect := githubOauthConfig.AuthCodeURL("", oauth2.S256ChallengeOption(verifier))
	http.Redirect(writer, request, redirect, http.StatusTemporaryRedirect)
	return false

	//user, pass, ok := request.BasicAuth()
	//if !ok || user != serverUser || pass != serverPass {
	//	writer.Header().Set("WWW-Authenticate", "Basic realm=\"Access to the gobughunt\"")
	//	writer.WriteHeader(http.StatusUnauthorized)
	//	return false
	//}
	//return true
}

func wrap[T any](handle func(request *http.Request) (T, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !basicAuth(writer, request) {
			return
		}
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
		if serverLocal {
			writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5000")
			writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(response)
	}
}

func main() {
	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()

	pgStorage, err := storage.NewPgQueries(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	server := http.NewServeMux()
	server.HandleFunc("/api/linters", wrap(func(request *http.Request) ([]LinterDto, error) {
		linters, err := pgStorage.ListBugHuntLinters(request.Context())
		if err != nil {
			return nil, err
		}
		dtoLinters := make([]LinterDto, 0, len(linters))
		for _, linter := range linters {
			dtoLinters = append(dtoLinters, LinterDto{
				Id:                 linter.LinterID,
				GitUrl:             linter.LinterGitUrl,
				GitBranch:          linter.LinterGitBranch,
				DockerImage:        storage.TryGetText(linter.LinterLastDockerImage),
				DockerImageShaHash: storage.TryGetText(linter.LinterLastDockerShaHash),
				StatDto: &StatDto{
					TotalHighlightDedup:    int(linter.TotalHighlightDedup),
					PendingHighlightDedup:  int(linter.PendingHighlightDedup),
					RejectedHighlightDedup: int(linter.RejectedHighlightDedup),
					AcceptedHighlightDedup: int(linter.AcceptedHighlightDedup),
				},
			})
		}
		return dtoLinters, nil
	}))
	server.HandleFunc("/api/repos", wrap(func(request *http.Request) ([]RepoDto, error) {
		repos, err := pgStorage.ListBugHuntRepos(request.Context())
		if err != nil {
			return nil, err
		}
		dtoRepos := make([]RepoDto, 0, len(repos))
		for _, repo := range repos {
			dtoRepos = append(dtoRepos, RepoDto{
				Id:            repo.RepoID,
				GitUrl:        repo.RepoGitUrl,
				GitBranch:     repo.RepoGitBranch,
				GitCommitHash: storage.TryGetText(repo.RepoLastGitCommitHash),
				StatDto: &StatDto{
					TotalHighlightDedup:    int(repo.TotalHighlightDedup),
					PendingHighlightDedup:  int(repo.PendingHighlightDedup),
					RejectedHighlightDedup: int(repo.RejectedHighlightDedup),
					AcceptedHighlightDedup: int(repo.AcceptedHighlightDedup),
				},
			})
		}
		return dtoRepos, nil
	}))
	server.HandleFunc("/api/lint-tasks", wrap(func(request *http.Request) ([]LintTaskDto, error) {
		params := request.URL.Query()
		skip, err := strconv.Atoi(params.Get("skip"))
		if err != nil {
			return nil, err
		}
		take, err := strconv.Atoi(params.Get("take"))
		if err != nil {
			return nil, err
		}
		take = max(1, min(take, skip+1000))
		args := db.ListBugHuntLintTasksParams{Offset: int32(skip), Limit: int32(take)}
		tasks, err := pgStorage.ListBugHuntLintTasks(request.Context(), args)
		if err != nil {
			return nil, err
		}
		dtoTasks := make([]LintTaskDto, 0, len(tasks))
		for _, task := range tasks {
			dtoTasks = append(dtoTasks, LintTaskDto{
				Id:              task.LintID,
				Status:          string(task.LintStatus),
				StatusComment:   storage.TryGetText(task.LintStatusComment),
				LintDurationSec: storage.TryGetDurationSec(task.LintDuration),
				Linter: LinterDto{
					Id:                 task.LinterID,
					GitUrl:             task.LinterGitUrl,
					GitBranch:          task.LinterGitBranch,
					DockerImage:        &task.LinterDockerImage,
					DockerImageShaHash: &task.LinterDockerShaHash,
				},
				Repo: RepoDto{
					Id:            task.RepoID,
					GitUrl:        task.RepoGitUrl,
					GitBranch:     task.RepoGitBranch,
					GitCommitHash: &task.RepoGitCommitHash,
				},
			})
		}
		return dtoTasks, nil
	}))
	server.Handle("/api/lint-highlights", wrap(func(request *http.Request) ([]LintHighlightDto, error) {
		params := request.URL.Query()
		lintId := params.Get("lintId")
		repoId := params.Get("repoId")
		linterId := params.Get("linterId")
		if lintId == "" && repoId == "" && linterId == "" {
			return nil, fmt.Errorf("one of three parameters should be set: lintId, repoId, linterId")
		}
		args := db.ListBugHuntHighlightsParams{LintID: lintId, RepoID: repoId, LinterID: linterId}
		highlights, err := pgStorage.ListBugHuntHighlights(request.Context(), args)
		if err != nil {
			return nil, err
		}
		dtoHighlights := make([]LintHighlightDto, 0, len(highlights))
		for _, highlight := range highlights {
			dtoHighlights = append(dtoHighlights, LintHighlightDto{
				LintId: highlight.LintID,
				Linter: LinterDto{
					Id:                 highlight.LinterID,
					GitUrl:             highlight.LinterGitUrl,
					GitBranch:          highlight.LinterGitBranch,
					DockerImage:        &highlight.LinterDockerImage,
					DockerImageShaHash: &highlight.LinterDockerShaHash,
				},
				Repo: RepoDto{
					Id:            highlight.RepoID,
					GitUrl:        highlight.RepoGitUrl,
					GitBranch:     highlight.RepoGitBranch,
					GitCommitHash: &highlight.RepoGitCommitHash,
				},
				Status:      string(highlight.ModerationStatus),
				Path:        highlight.Path,
				StartLine:   int(highlight.StartLine),
				EndLine:     int(highlight.EndLine),
				Explanation: highlight.Explanation,
				Snippet: HighlightSnippetDto{
					StartLine: int(highlight.SnippetStartLine),
					EndLine:   int(highlight.SnippetEndLine),
					Code:      highlight.SnippetCode,
				},
			})
		}
		return dtoHighlights, nil
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
		arg := db.ModerateBugHuntHighlightParams{
			LintID:           lintId,
			Path:             path,
			StartLine:        int32(startLine),
			EndLine:          int32(endLine),
			ModerationStatus: db.HighlightStatus(status),
		}
		return struct{}{}, pgStorage.ModerateBugHuntHighlight(request.Context(), arg)
	}))
	server.HandleFunc("/oauth/callback", func(writer http.ResponseWriter, request *http.Request) {
		code := request.URL.Query().Get("code")
		if code == "" {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		token, err := githubOauthConfig.Exchange(request.Context(), code)
		if err != nil {
			logging.Logger.Errorf("failed to get token from code: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		body, _ := io.ReadAll(request.Body)
		logging.Logger.Infof("callback: %v %v, token: %v", request.RequestURI, string(body), token)
	})
	server.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if !basicAuth(writer, request) {
			return
		}
		if !strings.HasPrefix(request.URL.Path, "/assets") {
			request.URL.Path = "/"
		}
		http.FileServer(http.Dir("./static")).ServeHTTP(writer, request)
	})
	err = http.ListenAndServe(serverListenAddr, server)
	if err != nil {
		logging.Logger.Errorf("exited server: %v", err)
	}
}
