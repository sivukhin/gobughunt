package lib

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
)

type GitRepo struct {
	CommitHash string
}

type GitApi interface {
	Fetch(ctx context.Context, gitUrl string, gitRef dto.GitRef, targetDir string) (GitRepo, error)
}

type naiveGitApi struct{}

var NaiveGitApi GitApi = naiveGitApi{}

func (_ naiveGitApi) Fetch(ctx context.Context, gitUrl string, gitRef dto.GitRef, targetDir string) (GitRepo, error) {
	_, err := runCommand(ctx, targetDir, "git", "init")
	if err != nil {
		return GitRepo{}, fmt.Errorf("unable to init git in the target dir %v: %w", targetDir, err)
	}
	_, err = runCommand(ctx, targetDir, "git", "remote", "add", "origin", gitUrl)
	if err != nil {
		return GitRepo{}, fmt.Errorf("unable to add origin remote %v: %w", gitUrl, err)
	}
	if gitRef.Branch != "" {
		_, err = runCommand(ctx, targetDir, "git", "fetch", "origin", gitRef.Branch)
		if err != nil {
			return GitRepo{}, fmt.Errorf("unable to fetch branch %v from repo %v: %w", gitRef.Branch, gitUrl, err)
		}
		_, err = runCommand(ctx, targetDir, "git", "checkout", gitRef.Branch)
		if err != nil {
			return GitRepo{}, fmt.Errorf("unable to checkout branch %v from repo %v: %w", gitRef.Branch, gitUrl, err)
		}
	} else if gitRef.CommitHash != "" {
		_, err = runCommand(ctx, targetDir, "git", "fetch", "origin", gitRef.CommitHash)
		if err != nil {
			return GitRepo{}, fmt.Errorf("unable to fetch revision %v from repo %v: %w", gitRef.CommitHash, gitUrl, err)
		}
		_, err = runCommand(ctx, targetDir, "git", "reset", "--hard", gitRef.CommitHash)
		if err != nil {
			return GitRepo{}, fmt.Errorf("unable to hard reset to revision %v in repo %v: %w", gitRef.CommitHash, gitUrl, err)
		}
	} else {
		logging.Logger.Fatalf("gitRef is empty for repo %v", gitUrl)
	}
	commitHash, err := runCommand(ctx, targetDir, "git", "rev-parse", "HEAD")
	if err != nil {
		return GitRepo{}, err
	}
	return GitRepo{CommitHash: strings.TrimSpace(commitHash)}, nil
}

func runCommand(ctx context.Context, targetDir, name string, args ...string) (string, error) {
	stderr := bytes.NewBuffer(nil)
	stdout := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = targetDir
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		errs, _ := io.ReadAll(stderr)
		return "", fmt.Errorf("command [%v %v] failed: %w (%v)", name, args, err, strings.TrimSpace(string(errs)))
	}
	output, _ := io.ReadAll(stdout)
	return string(output), nil
}
