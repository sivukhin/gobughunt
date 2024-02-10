package lib

import (
	"bufio"
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

const ContainerBindPath = "/src"

type naiveLinting struct {
	DockerApi DockerApi
	GitApi    GitApi
}

type Linting interface {
	Run(ctx context.Context, repo dto.RepoInstance, linter dto.LinterInstance) ([]dto.LintHighlightSnippet, error)
}

var NaiveLinting = naiveLinting{DockerApi: NaiveDockerApi, GitApi: NaiveGitApi}

var (
	LintTempErr    = errors.New("lint failed with temp error")
	LintFatalErr   = errors.New("lint failed with fatal error")
	LintCloneErr   = errors.New("lint clone failed")
	LintExecErr    = errors.New("lint exec failed")
	LintSkippedErr = errors.New("lint skipped")
)

func (l naiveLinting) Run(
	ctx context.Context,
	repo dto.RepoInstance,
	linter dto.LinterInstance,
) ([]dto.LintHighlightSnippet, error) {
	logging.Logger.Infof("start linting repo %v with linter %v", repo, linter)
	lintStartTime := time.Now()
	_ = lintStartTime

	targetDir, err := os.MkdirTemp(".", "repo_clone_*")
	if err != nil {
		return nil, fmt.Errorf("%w: mkdir temp failed: %w", LintTempErr, err)
	}
	defer os.RemoveAll(targetDir)

	cloneStartTime := time.Now()
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
		return nil, fmt.Errorf("%w: unable to get absolute path for directory %v: %w", LintTempErr, targetDir, err)
	}
	lines, err := l.DockerApi.Exec(ctx, linter.DockerImage, ContainerBindPath, targetDirAbs)
	if err != nil {
		logging.Logger.Errorf("exec of the linter %v against repo %v failed: err=%v, elapsed=%v", linter, repo, err, time.Since(execStartTime))

		if errors.Is(err, DockerNonZeroExitCodeErr) {
			return nil, fmt.Errorf("%w: linter %v exited with non-zero code: %w", LintExecErr, linter, err)
		}
		return nil, fmt.Errorf("%w: %w", LintTempErr, err)
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
		scanner := bufio.NewScanner(bytes.NewReader(content))
		lines := make([][]byte, 0)
		for scanner.Scan() {
			lines = append(lines, scanner.Bytes())
		}
		for _, highlight := range fileHighlights {
			if highlight.StartLine < 1 {
				return nil, fmt.Errorf("invalid highlight lines: highlight=%+v", highlight)
			}
			if highlight.StartLine > len(lines) || highlight.EndLine > len(lines) {
				return nil, fmt.Errorf("invalid highlight lines: highlight=%+v", highlight)
			}
			highlightSnippets = append(highlightSnippets, dto.LintHighlightSnippet{
				LintHighlight: highlight,
				Snippet:       string(bytes.Join(lines[highlight.StartLine-1:highlight.EndLine], []byte(`\n`))),
			})
		}
	}
	return highlightSnippets, nil
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
