package lib

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
)

const ContainerBindPath = "/home/repo"

type NaiveLinting struct {
	TempDir   string
	DockerApi DockerApi
	GitApi    GitApi
}

type Linting interface {
	Run(ctx context.Context, repo dto.RepoInstance, linter dto.LinterInstance) ([]dto.LintHighlightSnippet, error)
}

var Lint = NaiveLinting{DockerApi: Docker, GitApi: Git}

var (
	LintTempErr    = errors.New("lint failed with temp error")
	LintFatalErr   = errors.New("lint failed with fatal error")
	LintCloneErr   = errors.New("lint clone failed")
	LintExecErr    = errors.New("lint exec failed")
	LintSkippedErr = errors.New("lint skipped")
)

func (l NaiveLinting) Run(
	ctx context.Context,
	repo dto.RepoInstance,
	linter dto.LinterInstance,
) ([]dto.LintHighlightSnippet, error) {
	logging.Logger.Infof("start linting repo %v with linter %v", repo, linter)
	lintStartTime := time.Now()

	targetDir, err := os.MkdirTemp(l.TempDir, "repo_clone_*")
	if err != nil {
		return nil, fmt.Errorf("%w: mkdir temp failed: %w", LintTempErr, err)
	}
	defer func() {
		err := os.RemoveAll(targetDir)
		if err != nil {
			logging.Logger.Errorf("failed to remove temp dir %v: %v", targetDir, err)
		}
	}()

	cloneStartTime := time.Now()
	logging.Logger.Infof("ready to clone repo %v to the directory %v", repo, targetDir)
	_, err = l.GitApi.Fetch(ctx, repo.GitUrl, dto.GitRef{CommitHash: repo.GitCommitHash}, targetDir)
	if err != nil {
		logging.Logger.Errorf("clone of repo %v to the directory %v failed: err=%v, elapsed=%v", repo, targetDir, err, time.Since(cloneStartTime))
		return nil, fmt.Errorf("%w: clone of repo %v failed: %w", LintCloneErr, repo, err)
	} else {
		logging.Logger.Infof("clone of repo %v to the directory %v succeeded: elapsed=%v", repo, targetDir, time.Since(cloneStartTime))
	}

	execStartTime := time.Now()
	targetDirAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to get absolute path for directory %v: %w", LintExecErr, targetDir, err)
	}
	logging.Logger.Infof("ready to lint repo %v with linter %v", repo, linter)
	lines, err := l.DockerApi.Exec(
		ctx,
		fmt.Sprintf("%v@sha256:%v", linter.DockerImage, linter.DockerImageShaHash),
		ContainerBindPath,
		targetDirAbs,
	)
	if err != nil {
		logging.Logger.Errorf("exec of the linter %v against repo %v failed: err=%v, lines=%v, elapsed=%v", linter, repo, err, lines, time.Since(execStartTime))

		if errors.Is(err, DockerNonZeroExitCodeErr) {
			return nil, fmt.Errorf("%w: linter %v exited with non-zero code: %w", LintExecErr, linter, err)
		}
		return nil, fmt.Errorf("%w: %w", LintExecErr, err)
	} else {
		logging.Logger.Infof("exec of the linter %v against repo %v succeed: elapsed=%v", linter, repo, time.Since(execStartTime))
	}
	highlights, skipped := ExtractHighlights(lines)
	logging.Logger.Infof("linting of repo %v with linter %v succeed: len(highlights)=%v, skipped=%v, elapsed=%v", repo, linter, len(highlights), skipped, time.Since(lintStartTime))
	if skipped {
		return nil, LintSkippedErr
	}
	highlightSnippets, err := ExtractHighlightSnippets(targetDir, highlights)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to extract snippets: %w", LintFatalErr, err)
	}
	return highlightSnippets, nil
}

var (
	// GitHub actions formatting
	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-a-warning-message
	ghDelimiter         = "::"
	ghSupportedPrefixes = [...]string{"::warning ", "::error "}
	ghMessageProp       = "message"
	ghTitleProp         = "title"
	ghFileProp          = "file"
	ghStartLineProp     = "line"
	ghEndLineProp       = "endLine"
)

var (
	skipPrefix = "::skip"
)

func ExtractHighlightSnippets(targetDir string, highlights []dto.LintHighlight) ([]dto.LintHighlightSnippet, error) {
	highlightSnippets := make([]dto.LintHighlightSnippet, 0, len(highlights))

	files := make(map[string][]dto.LintHighlight)
	for _, highlight := range highlights {
		fullPath := path.Join(targetDir, highlight.Path)
		files[fullPath] = append(files[fullPath], highlight)
	}
	for fullPath, fileHighlights := range files {
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read file '%v': %w", fullPath, err)
		}
		fileHighlightSnippets, err := ExtractHighlightSnippetsForFile(content, fileHighlights)
		if err != nil {
			return nil, fmt.Errorf("unable to extract snippets for file '%v': %w", fullPath, err)
		}
		highlightSnippets = append(highlightSnippets, fileHighlightSnippets...)
	}
	return highlightSnippets, nil
}

const snippetSurroundingLines = 2

func ExtractHighlightSnippetsForFile(content []byte, highlights []dto.LintHighlight) ([]dto.LintHighlightSnippet, error) {
	snippets := make([]dto.LintHighlightSnippet, 0, len(highlights))
	lines := bytes.Split(content, []byte("\n"))
	for _, highlight := range highlights {
		if highlight.StartLine < 1 {
			return nil, fmt.Errorf("invalid highlight lines: highlight=%+v", highlight)
		}
		if highlight.StartLine > len(lines) || highlight.EndLine > len(lines) {
			return nil, fmt.Errorf("invalid highlight lines: highlight=%+v", highlight)
		}
		startLine := max(0, highlight.StartLine-1-snippetSurroundingLines)
		endLine := min(len(lines), highlight.EndLine+snippetSurroundingLines)
		snippets = append(snippets, dto.LintHighlightSnippet{
			LintHighlight: highlight,
			Snippet: dto.HighlightSnippet{
				StartLine: startLine + 1,
				EndLine:   endLine,
				Code:      string(bytes.Join(ReduceIndentation(lines[startLine:endLine]), []byte("\n"))),
			},
		})
	}
	return snippets, nil
}

func isSpace(c byte) bool { return c == '\t' || c == ' ' }

func ReduceIndentation(lines [][]byte) [][]byte {
	if len(lines) == 0 {
		return lines
	}
	toCut := len(lines[0])
	for i := 1; i < len(lines); i++ {
		commonPrefix := 0
		for commonPrefix < len(lines[i-1]) && commonPrefix < len(lines[i]) && lines[i][commonPrefix] == lines[i-1][commonPrefix] && isSpace(lines[i][commonPrefix]) {
			commonPrefix++
		}
		toCut = min(toCut, commonPrefix)
	}
	reduced := make([][]byte, 0, len(lines))
	for _, line := range lines {
		reduced = append(reduced, line[toCut:])
	}
	return reduced
}

func ExtractHighlights(rawLines []string) (highlights []dto.LintHighlight, skipped bool) {
	for _, line := range rawLines {
		if strings.HasPrefix(line, skipPrefix) {
			return nil, true
		}
		var suffix string
		var ok bool
		for _, prefix := range ghSupportedPrefixes {
			if suffix, ok = strings.CutPrefix(line, prefix); ok {
				break
			}
		}
		if !ok {
			continue
		}
		tokens := strings.SplitN(suffix, ghDelimiter, 2)
		attributeStrings := strings.Split(tokens[0], ",")
		attributes := make(map[string]string)
		for _, attribute := range attributeStrings {
			keyValue := strings.SplitN(attribute, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			attributes[keyValue[0]] = keyValue[1]
		}
		if len(tokens) == 2 {
			attributes[ghMessageProp] = tokens[1]
		}
		startLine, err := strconv.Atoi(attributes[ghStartLineProp])
		if err != nil {
			logging.Logger.Debugf("failed to convert line attribute for output string: err=%v, line='%v'", err, line)
			continue
		}
		var endLine int
		if endLine, err = strconv.Atoi(attributes[ghEndLineProp]); err != nil {
			endLine = startLine
		}
		highlightPath := attributes[ghFileProp]
		if highlightPath == "" {
			logging.Logger.Debugf("file attribute absent in output string: line='%v'", line)
			continue
		}
		var explanation string
		title := attributes[ghTitleProp]
		message := attributes[ghMessageProp]
		if title != "" && message != "" {
			explanation = title + ": " + message
		} else if title != "" {
			explanation = title
		} else if message != "" {
			explanation = message
		}

		highlights = append(highlights, dto.LintHighlight{
			Path:        highlightPath,
			StartLine:   startLine,
			EndLine:     endLine,
			Explanation: explanation,
		})
	}
	return highlights, false
}
