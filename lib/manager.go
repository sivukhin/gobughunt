package lib

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Microsoft/go-winio/pkg/guid"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/storage"
	"github.com/sivukhin/gobughunt/lib/timeout"
	"github.com/sivukhin/gobughunt/lib/utils"
)

type Manager struct {
	LinterStorage   storage.LinterStorage
	RepoStorage     storage.RepoStorage
	LintStorage     storage.LintStorage
	DockerApi       DockerApi
	GitApi          GitApi
	ScheduleTimeout time.Duration
	ScheduleDelay   time.Duration
	ShortDelay      time.Duration
}

func (m Manager) ManageForever(ctx context.Context) {
	logging.Logger.Infof("manager started: scheduleTimeout=%v, scheduleDelay=%v, shortDelay=%v", m.ScheduleTimeout, m.ScheduleDelay, m.ShortDelay)
	<-timeout.RunForeverAsync("scheduler", ctx, m.ScheduleTimeout, func(ctx context.Context) time.Duration {
		allLinters, err := m.LinterStorage.List(ctx)
		if err != nil {
			logging.Logger.Errorf("failed to fetch all linters: %v", err)
			return m.ShortDelay
		}
		linters := allLinters.SelectWithInstances()
		logging.Logger.Infof("found %v linters, %v with instances", len(allLinters), len(linters))

		allRepos, err := m.RepoStorage.List(ctx)
		if err != nil {
			logging.Logger.Errorf("failed to fetch all repos: %v", err)
			return m.ShortDelay
		}
		repos := allRepos.SelectWithInstances()
		logging.Logger.Infof("found %v repos, %v with instances", len(allRepos), len(repos))

		for _, repo := range allRepos {
			err := m.RefreshRepo(ctx, repo)
			if err != nil {
				logging.Logger.Errorf("failed refresh of repo %+v: %v", repo.Meta, err)
			} else {
				logging.Logger.Infof("succeeded with refresh of repo %+v", repo.Meta)
			}
		}

		for _, repo := range repos {
			for _, linter := range linters {
				err := m.ManageOnce(ctx, repo, linter)
				if err != nil {
					logging.Logger.Errorf("failed single iteration: %v", err)
				} else {
					logging.Logger.Infof("succeed with single iteration")
				}
			}
		}
		return m.ScheduleDelay
	})
}

func (m Manager) RefreshRepo(ctx context.Context, repo dto.Repo) error {
	targetDir, err := os.MkdirTemp(".", "repo_clone_*")
	if err != nil {
		return fmt.Errorf("mkdir temp failed: %w", err)
	}
	defer os.Remove(targetDir)
	info, err := m.GitApi.Fetch(ctx, repo.Meta.RepoGitUrl, dto.GitRef{Branch: repo.Meta.RepoGitBranch}, targetDir)
	if err != nil {
		return fmt.Errorf("failed to fetch repo %v: %w", repo, err)
	}
	return m.RepoStorage.AddOrUpdate(ctx, dto.Repo{
		Meta: repo.Meta,
		Instance: &dto.RepoInstance{
			RepoId:        repo.Meta.RepoId,
			GitUrl:        repo.Meta.RepoGitUrl,
			GitCommitHash: info.CommitHash,
		},
	}, time.Now())
}

func (m Manager) ManageOnce(ctx context.Context, repo dto.Repo, linter dto.Linter) error {
	lintId := utils.Must(guid.NewV4()).String()
	lintTask := dto.LintTask{LintId: lintId, Linter: *linter.Instance, Repo: *repo.Instance}
	err := m.LintStorage.TryAdd(ctx, lintTask, time.Now())
	if errors.Is(err, storage.DuplicateTaskErr) {
		return nil
	} else if err != nil {
		logging.Logger.Errorf("failed to add task %+v: %v", lintTask, err)
		return err
	} else {
		logging.Logger.Infof("successfully added task %+v", lintTask)
	}
	return nil
}
