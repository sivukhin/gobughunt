package lib

import (
	"context"
	"errors"
	"time"

	"github.com/Microsoft/go-winio/pkg/guid"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/storage"
	"github.com/sivukhin/gobughunt/lib/timeout"
	"github.com/sivukhin/gobughunt/lib/utils"
)

type Manager struct {
	LinterStorage    storage.LinterStorage
	RepoStorage      storage.RepoStorage
	LintStorage      storage.LintStorage
	DockerApi        DockerApi
	GitApi           GitApi
	IterationTimeout time.Duration
	IterationDelay   time.Duration
}

func (m Manager) ManageForever(ctx context.Context) {
	logging.Logger.Infof("manager started: timeout=%v, delay=%v", m.IterationTimeout, m.IterationDelay)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		iterationCtx, cancel := context.WithTimeout(ctx, m.IterationTimeout)
		err := m.ManageOnce(iterationCtx)
		cancel()

		if err != nil {
			logging.Logger.Errorf("failed single iteration, sleeping for %v: %v", m.IterationDelay, err)
		} else {
			logging.Logger.Infof("succeed with single iteration, sleeping for %v", m.IterationDelay)
		}
		timeout.SleepOrDone(ctx, m.IterationDelay)
	}
}

func (m Manager) ManageOnce(ctx context.Context) error {
	allLinters, err := m.LinterStorage.List(ctx)
	if err != nil {
		return err
	}
	linters := allLinters.SelectWithInstances()
	logging.Logger.Infof("found %v linters, %v with instances", len(allLinters), len(linters))

	allRepos, err := m.RepoStorage.List(ctx)
	if err != nil {
		return err
	}
	repos := allRepos.SelectWithInstances()
	logging.Logger.Infof("found %v repos, %v with instances", len(allRepos), len(repos))

	for _, linter := range linters {
		if linter.Instance == nil {
			continue
		}
		for _, repo := range repos {
			lintId := utils.Must(guid.NewV4()).String()
			lintTask := dto.LintTask{LintId: lintId, Linter: *linter.Instance, Repo: *repo.Instance}
			err := m.LintStorage.TryAdd(ctx, lintTask, time.Now())
			if errors.Is(err, storage.DuplicateTaskErr) {
				continue
			} else if err != nil {
				logging.Logger.Errorf("failed to add task %+v: %v", lintTask, err)
			} else {
				logging.Logger.Infof("successfully added task %+v", lintTask)
			}
		}
	}
	return nil
}
