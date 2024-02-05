package lib

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type progress struct{ repo string }

func (p progress) Write(data []byte) (int, error) {
	logLine := fmt.Sprintf("clone progress(%v): %v", p.repo, string(data))
	logLine = strings.ReplaceAll(logLine, "\r", "")
	if !strings.HasSuffix(logLine, "\n") {
		logLine += "\n"
	}
	return os.Stderr.Write([]byte(logLine))
}

func GitCloneRepository(ctx context.Context, repoUrl, repoBranch, targetDir string) error {
	_, err := git.PlainCloneContext(ctx, targetDir, true, &git.CloneOptions{
		URL:           repoUrl,
		ReferenceName: plumbing.NewBranchReferenceName(repoBranch),
		SingleBranch:  true,
		Depth:         1,
		Progress:      progress{repo: repoUrl},
	})
	if err != nil {
		return fmt.Errorf("failed to clone branch %v of repo %v: %w", repoBranch, repoUrl, err)
	}
	return nil
}
