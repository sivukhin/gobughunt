package main

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"

	_ "golang.org/x/oauth2/endpoints"

	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/utils"
	"github.com/sivukhin/gobughunt/storage"
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
	//go:embed templates/dashboard.html
	dashboardTemplateString string
	//go:embed templates/lint-tasks.html
	lintTasksTemplateString string
	//go:embed templates/lint-highlights.html
	lintHighlightsTemplateString string
	//go:embed templates/about.html
	aboutTemplateString string
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

func wrap(handle func(request *http.Request, writer http.ResponseWriter) (string, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !basicAuth(writer, request) {
			return
		}
		result, err := handle(request, writer)
		if err != nil {
			// todo (sivukhin, 2024-02-25): add more errors
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		if serverLocal {
			writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5000")
			writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		}
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(result))
	}
}

func main() {
	templateFuncs := template.FuncMap{
		"DerefF64": func(f *float64) float64 { return *f },
		"DerefStr": func(s *string) string { return *s },
	}

	var (
		dashboardTemplate      = template.Must(template.New("dashboard").Funcs(templateFuncs).Parse(dashboardTemplateString))
		lintTasksTemplate      = template.Must(template.New("lint-tasks").Funcs(templateFuncs).Parse(lintTasksTemplateString))
		lintHighlightsTemplate = template.Must(template.New("lint-highlights").Funcs(templateFuncs).Parse(lintHighlightsTemplateString))
		aboutTemplate          = template.Must(template.New("about").Funcs(templateFuncs).Parse(aboutTemplateString))
	)

	connectCtx, cancel := context.WithTimeout(context.Background(), connectionDuration)
	defer cancel()

	pgStorage, err := storage.NewPgQueries(connectCtx, connectionString)
	if err != nil {
		logging.Logger.Fatalf("failed to create task storage: %v", err)
	}

	apiController := ApiController{Storage: pgStorage}

	server := http.NewServeMux()
	static := http.FileServer(http.Dir("./static"))
	server.Handle("/static/", http.StripPrefix("/static/", static))
	server.Handle("/lint-highlights", wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
		params := request.URL.Query()
		lintId := params.Get("lintId")
		repoId := params.Get("repoId")
		linterId := params.Get("linterId")
		if lintId == "" && repoId == "" && linterId == "" {
			return "", fmt.Errorf("one of three parameters should be set: lintId, repoId, linterId")
		}
		dtoHighlights, err := apiController.LintHighlights(request.Context(), lintId, repoId, linterId)
		if err != nil {
			return "", err
		}
		return RenderTemplate(lintHighlightsTemplate, dtoHighlights)
	}))
	server.HandleFunc("/lint-tasks", wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
		dtoTasks, err := apiController.LintTasks(request.Context(), 0, 10000)
		if err != nil {
			return "", err
		}
		return RenderTemplate(lintTasksTemplate, dtoTasks)
	}))
	server.HandleFunc("/about", wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
		return RenderTemplate(aboutTemplate, struct{}{})
	}))
	server.Handle("/lint-highlight/moderate", wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
		params := request.URL.Query()
		lintId := params.Get("lintId")
		if lintId == "" {
			return "", fmt.Errorf("lintId required")
		}
		path := params.Get("path")
		if path == "" {
			return "", fmt.Errorf("path required")
		}
		startLineString := params.Get("startLine")
		if startLineString == "" {
			return "", fmt.Errorf("startLine required")
		}
		startLine, err := strconv.Atoi(startLineString)
		if err != nil {
			return "", err
		}
		endLineString := params.Get("endLine")
		if endLineString == "" {
			return "", fmt.Errorf("endLine required")
		}
		endLine, err := strconv.Atoi(endLineString)
		if err != nil {
			return "", err
		}
		status := params.Get("status")
		if status != "accepted" && status != "rejected" {
			return "", fmt.Errorf("unexpected status: %v", status)
		}
		err = apiController.LintHighlightModerate(request.Context(), lintId, path, startLine, endLine, status)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`<div class="%v">%v</div>`, status, status), nil
	}))
	server.HandleFunc("/", wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
		dashboardDto, err := apiController.Dashboard(request.Context())
		if err != nil {
			return "", err
		}
		return RenderTemplate(dashboardTemplate, dashboardDto)
	}))
	//server.HandleFunc("/oauth/callback", func(writer http.ResponseWriter, request *http.Request) {
	//	code := request.URL.Query().Get("code")
	//	if code == "" {
	//		writer.WriteHeader(http.StatusUnauthorized)
	//		return
	//	}
	//	token, err := githubOauthConfig.Exchange(request.Context(), code)
	//	if err != nil {
	//		logging.Logger.Errorf("failed to get token from code: %v", err)
	//		writer.WriteHeader(http.StatusInternalServerError)
	//		return
	//	}
	//	writer.WriteHeader(http.StatusOK)
	//	body, _ := io.ReadAll(request.Body)
	//	logging.Logger.Infof("callback: %v %v, token: %v", request.RequestURI, string(body), token)
	//})
	err = http.ListenAndServe(serverListenAddr, server)
	if err != nil {
		logging.Logger.Errorf("exited server: %v", err)
	}
}
