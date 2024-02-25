package lib

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/timeout"
	"github.com/sivukhin/gobughunt/lib/utils"
	"github.com/sivukhin/gobughunt/storage"
	"github.com/sivukhin/gobughunt/storage/db"
)

type Manager struct {
	Storage             *db.Queries
	DockerApi           DockerApi
	GitApi              GitApi
	FetchTimeout        time.Duration
	RefreshTimeout      time.Duration
	ScheduleTimeout     time.Duration
	ManagerFailDelay    time.Duration
	ManagerSuccessDelay time.Duration
}

func (m Manager) ManageForever(ctx context.Context) {
	logging.Logger.Infof(
		"manager started: fetchTimeout=%v, refreshTimeout=%v, scheduleTimeout=%v, managerFailDelay=%v, managerSuccessDelay=%v",
		m.FetchTimeout,
		m.RefreshTimeout,
		m.ScheduleTimeout,
		m.ManagerFailDelay,
		m.ManagerSuccessDelay,
	)
	periodic := timeout.Periodic(ctx, m.ManagerFailDelay, m.ManagerSuccessDelay)
	repos := timeout.Process("fetch-repos", periodic, m.FetchTimeout, func(ctx context.Context, _ struct{}, next func(result dto.Repo)) error {
		repos, err := m.Storage.ListRepos(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch all repos: %w", err)
		}
		for _, repo := range repos {
			commitHash, _ := repo.RepoLastGitCommitHash.Value()
			var instance *dto.RepoInstance
			if commitHash != nil {
				instance = &dto.RepoInstance{
					Id:            repo.RepoID,
					GitUrl:        repo.RepoGitUrl,
					GitCommitHash: commitHash.(string),
				}
			}
			next(dto.Repo{
				Meta: dto.RepoMeta{
					Id:        repo.RepoID,
					GitUrl:    repo.RepoGitUrl,
					GitBranch: repo.RepoGitBranch,
				},
				Instance: instance,
			})
		}
		return nil
	})
	refresh := timeout.Process("refresh-repos", repos, m.RefreshTimeout, func(ctx context.Context, repo dto.Repo, next func(result dto.Repo)) error {
		logging.Logger.Infof("starting refresh of the repo %+v", repo)
		updated, err := m.RefreshRepo(ctx, repo)
		if err != nil {
			return fmt.Errorf("failed refresh of repo %+v: %w", repo.Meta, err)
		}
		logging.Logger.Infof("succeeded with refresh of repo %+v", updated)
		next(updated)
		return nil
	})
	scheduler := timeout.Process("scheduler", refresh, m.ScheduleTimeout, func(ctx context.Context, repo dto.Repo, next func(result struct{})) error {
		allLinters, err := m.Storage.ListLinters(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch all linters: %w", err)
		}
		linters := make([]dto.Linter, 0)
		for _, linter := range allLinters {
			dockerImage, _ := linter.LinterLastDockerImage.Value()
			dockerImageSha, _ := linter.LinterLastDockerShaHash.Value()
			if dockerImage == nil || dockerImageSha == nil {
				continue
			}
			instance := &dto.LinterInstance{
				Id:                 linter.LinterID,
				DockerImage:        dockerImage.(string),
				DockerImageShaHash: dockerImageSha.(string),
			}
			linters = append(linters, dto.Linter{
				Meta: dto.LinterMeta{
					Id:        linter.LinterID,
					GitUrl:    linter.LinterGitUrl,
					GitBranch: linter.LinterGitBranch,
				},
				Instance: instance,
			})
		}
		logging.Logger.Infof("found %v linters, %v with instances", len(allLinters), len(linters))
		for _, linter := range linters {
			err := m.ManageOnce(ctx, repo, linter)
			if err != nil {
				logging.Logger.Errorf("failed single iteration: %v", err)
			} else {
				logging.Logger.Infof("succeed with single iteration")
			}
		}
		return nil
	})
	timeout.Close(scheduler)
}

func (m Manager) RefreshRepo(ctx context.Context, repo dto.Repo) (dto.Repo, error) {
	targetDir, err := os.MkdirTemp("", "repo_clone_*")
	if err != nil {
		return dto.Repo{}, fmt.Errorf("mkdir temp failed: %w", err)
	}
	err = os.Chmod(targetDir, 0660)
	if err != nil {
		return dto.Repo{}, fmt.Errorf("%w: chmod failed: %w", LintTempErr, err)
	}
	defer func() {
		err := os.RemoveAll(targetDir)
		if err != nil {
			logging.Logger.Errorf("failed to remove temp dir %v: %v", targetDir, err)
		}
	}()
	info, err := m.GitApi.Fetch(ctx, repo.Meta.GitUrl, dto.GitRef{Branch: repo.Meta.GitBranch}, targetDir)
	if err != nil {
		return dto.Repo{}, fmt.Errorf("failed to fetch repo %v: %w", repo, err)
	}
	updated := dto.Repo{
		Meta: repo.Meta,
		Instance: &dto.RepoInstance{
			Id:            repo.Meta.Id,
			GitUrl:        repo.Meta.GitUrl,
			GitCommitHash: info.CommitHash,
		},
	}
	err = m.Storage.UpsertRepo(ctx, db.UpsertRepoParams{
		RepoID:                repo.Meta.Id,
		RepoGitUrl:            repo.Meta.GitUrl,
		RepoGitBranch:         repo.Meta.GitBranch,
		RepoLastGitCommitHash: pgtype.Text{String: info.CommitHash, Valid: true},
		CreatedAt:             pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	return updated, err
}

func (m Manager) ManageOnce(ctx context.Context, repo dto.Repo, linter dto.Linter) error {
	lintId := utils.Must(guid.NewV4()).String()
	lintTask := dto.LintTask{Id: lintId, Linter: *linter.Instance, Repo: *repo.Instance}
	err := m.Storage.AddLintTask(ctx, db.AddLintTaskParams{
		LintID:              lintId,
		LintStatus:          db.LintStatusPending,
		LinterID:            linter.Meta.Id,
		LinterDockerImage:   linter.Instance.DockerImage,
		LinterDockerShaHash: linter.Instance.DockerImageShaHash,
		RepoID:              repo.Meta.Id,
		RepoGitUrl:          repo.Meta.GitUrl,
		RepoGitCommitHash:   repo.Instance.GitCommitHash,
		CreatedAt:           pgtype.Timestamp{Time: time.Now(), Valid: true},
	})

	if storage.ViolatesUniqueConstraint(err) {
		return nil
	} else if err != nil {
		logging.Logger.Errorf("failed to add task %+v: %v", lintTask, err)
		return err
	} else {
		logging.Logger.Infof("successfully added task %+v", lintTask)
	}
	return nil
}
