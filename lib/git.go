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

type GitApi interface {
	Clone(ctx context.Context, gitUrl string, gitRef dto.GitRef, targetDir string) error
}

type naiveGitApi struct{}

var NaiveGitApi GitApi = naiveGitApi{}

func (_ naiveGitApi) Clone(ctx context.Context, gitUrl string, gitRef dto.GitRef, targetDir string) error {
	err := runCommand(ctx, targetDir, "git", "init")
	if err != nil {
		return fmt.Errorf("unable to init git in the target dir %v: %w", targetDir, err)
	}
	err = runCommand(ctx, targetDir, "git", "remote", "add", "origin", gitUrl)
	if err != nil {
		return fmt.Errorf("unable to add origin remote %v: %w", gitUrl, err)
	}
	if gitRef.Branch != "" {
		err = runCommand(ctx, targetDir, "git", "fetch", "origin", gitRef.Branch)
		if err != nil {
			return fmt.Errorf("unable to fetch branch %v from repo %v: %w", gitRef.Branch, gitUrl, err)
		}
		err = runCommand(ctx, targetDir, "git", "checkout", gitRef.Branch)
		if err != nil {
			return fmt.Errorf("unable to checkout branch %v from repo %v: %w", gitRef.Branch, gitUrl, err)
		}
	} else if gitRef.CommitHash != "" {
		err = runCommand(ctx, targetDir, "git", "fetch", "origin", gitRef.CommitHash)
		if err != nil {
			return fmt.Errorf("unable to fetch revision %v from repo %v: %w", gitRef.CommitHash, gitUrl, err)
		}
		err = runCommand(ctx, targetDir, "git", "reset", "--hard", gitRef.CommitHash)
		if err != nil {
			return fmt.Errorf("unable to hard reset to revision %v in repo %v: %w", gitRef.CommitHash, gitUrl, err)
		}
	} else {
		logging.Logger.Fatalf("gitRef is empty for repo %v", gitUrl)
	}
	return nil
}

func runCommand(ctx context.Context, targetDir, name string, args ...string) error {
	stderr := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = targetDir
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		errs, _ := io.ReadAll(stderr)
		return fmt.Errorf("command [%v %v] failed: %w (%v)", name, args, err, strings.TrimSpace(string(errs)))
	}
	return nil
}
