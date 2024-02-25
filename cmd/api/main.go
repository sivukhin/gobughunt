package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"

	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/utils"
	"github.com/sivukhin/gobughunt/storage"
)

var (
	connectionDuration       = utils.EnvMustParseDurationSec("CONNECTION_DURATION_SEC")
	connectionString         = utils.EnvMustParseString("CONNECTION_STRING")
	serverLocal              = utils.EnvTryParseBool("SERVER_LOCAL")
	serverListenAddr         = utils.EnvMustParseString("SERVER_LISTEN_ADDR")
	serverJwtSecretKey       = []byte(utils.EnvMustParseString("SERVER_JWT_SECRET_KEY"))
	serverModeratorLogins    = utils.EnvMustParseStringArray("SERVER_MODERATOR_LOGINS")
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
		RedirectURL:  utils.Ternary(serverLocal, "http://localhost:3000", ""),
	}
)

func basicAuth(writer http.ResponseWriter, request *http.Request) (*http.Request, bool) {
	if serverLocal {
		return request, true
	}
	signedJwt, err := request.Cookie("GobughuntJwt")
	if err != nil {
		return request, true
	}
	var claims jwt.MapClaims
	token, err := jwt.ParseWithClaims(signedJwt.Value, &claims, func(token *jwt.Token) (interface{}, error) { return serverJwtSecretKey, nil })
	if err != nil {
		logging.Logger.Errorf("failed to parse JWT token: %v", err)
		return request, true
	}
	if !token.Valid {
		logging.Logger.Errorf("JWT token is invalid")
		return request, true
	}
	user, ok := claims["user"]
	if !ok {
		return request, true
	}
	return request.WithContext(context.WithValue(request.Context(), "user", user)), true
	//verifier := oauth2.GenerateVerifier()
	//redirect := githubOauthConfig.AuthCodeURL("", oauth2.S256ChallengeOption(verifier))
	//http.Redirect(writer, request, redirect, http.StatusTemporaryRedirect)
	//return false

	//user, pass, ok := request.BasicAuth()
	//if !ok || user != serverUser || pass != serverPass {
	//	writer.Header().Set("WWW-Authenticate", "Basic realm=\"Access to the gobughunt\"")
	//	writer.WriteHeader(http.StatusUnauthorized)
	//	return false
	//}
	//return true
}

func log(handle http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()
		handle.ServeHTTP(writer, request)
		logging.Logger.Infof("%v %v - %v", request.Method, request.RequestURI, time.Since(startTime))
	}
}

func wrap(handle func(request *http.Request, writer http.ResponseWriter) (string, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		update, ok := basicAuth(writer, request)
		if !ok {
			return
		}
		request = update

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

	apiController := ApiController{
		Storage:         pgStorage,
		ModeratorLogins: serverModeratorLogins,
	}

	server := http.NewServeMux()
	static := http.FileServer(http.Dir("./static"))
	server.Handle("/static/", log(http.StripPrefix("/static/", static)))
	server.Handle("/lint-highlights", log(wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
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
	})))
	server.HandleFunc("/lint-tasks", log(wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
		dtoTasks, err := apiController.LintTasks(request.Context(), 0, 10000)
		if err != nil {
			return "", err
		}
		return RenderTemplate(lintTasksTemplate, dtoTasks)
	})))
	server.HandleFunc("/about", log(wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
		login, _ := request.Context().Value("user").(string)
		return RenderTemplate(aboutTemplate, struct{ Login string }{Login: login})
	})))
	server.HandleFunc("/login", log(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		verifier := oauth2.GenerateVerifier()
		redirect := githubOauthConfig.AuthCodeURL("", oauth2.S256ChallengeOption(verifier))
		http.Redirect(writer, request, redirect, http.StatusTemporaryRedirect)
	})))
	server.HandleFunc("/logout", log(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Set-Cookie", "GobughuntJwt=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
	})))
	server.Handle("/lint-highlight/moderate", log(wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
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
	})))
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
		githubRequest, err := http.NewRequestWithContext(request.Context(), "GET", "https://api.github.com/user", nil)
		if err != nil {
			logging.Logger.Errorf("failed to create request to GitHub: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		githubRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
		response, err := http.DefaultClient.Do(githubRequest)
		if err != nil {
			logging.Logger.Errorf("failed to exec request to GitHub: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		data, err := io.ReadAll(response.Body)
		if err != nil {
			logging.Logger.Errorf("failed to read response from GitHub: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		var user struct {
			Login string `json:"login"`
		}
		err = json.Unmarshal(data, &user)
		if err != nil {
			logging.Logger.Errorf("failed to unmarshal response from GitHub: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user": user.Login})
		signedJwtToken, err := jwtToken.SignedString(serverJwtSecretKey)
		if err != nil {
			logging.Logger.Errorf("failed sign JWT token: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Set-Cookie", "GobughuntJwt="+signedJwtToken+"; path=/")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
	})
	server.HandleFunc("/", log(wrap(func(request *http.Request, writer http.ResponseWriter) (string, error) {
		dashboardDto, err := apiController.Dashboard(request.Context())
		if err != nil {
			return "", err
		}
		return RenderTemplate(dashboardTemplate, dashboardDto)
	})))
	err = http.ListenAndServe(serverListenAddr, server)
	if err != nil {
		logging.Logger.Errorf("exited server: %v", err)
	}
}
