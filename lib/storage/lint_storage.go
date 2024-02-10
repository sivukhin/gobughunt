package storage

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/sivukhin/gobughunt/lib/dto"
)

var (
	NoTasksErr       = errors.New("ready tasks not found")
	DuplicateTaskErr = errors.New("duplicate task")
)

type LintStorage interface {
	TryAdd(ctx context.Context, lintTask dto.LintTask, createdAt time.Time) error
	TryTake(ctx context.Context, lockTimeLowerBound time.Time, lockedAt time.Time) (dto.LintTask, error)
	Set(ctx context.Context, lintTask dto.LintTask, lintResult dto.LintResult, lintedAt time.Time) error
}

type PgLintStorage PgStorage

var _ LinterStorage = PgLinterStorage{}

//go:embed queries/lint_task_insert.sql
var lintTaskInsertSql string

//go:embed queries/lint_task_take.sql
var lintTaskTakeSql string

//go:embed queries/lint_task_set.sql
var lintTaskSetSql string

//go:embed queries/lint_highlights_add.sql
var lintHighlightsAddSql string

func (p PgLintStorage) TryAdd(ctx context.Context, lintTask dto.LintTask, createdAt time.Time) error {
	_, err := p.Exec(
		ctx,
		lintTaskInsertSql,
		lintTask.Id,
		"pending",
		lintTask.Linter.Id,
		lintTask.Linter.DockerImage,
		lintTask.Linter.DockerImageShaHash,
		lintTask.Repo.Id,
		lintTask.Repo.GitUrl,
		lintTask.Repo.GitCommitHash,
		createdAt,
	)
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		if pgError.Code == "23505" /* unique_violation */ {
			return fmt.Errorf("%w: lint task %+v is duplicate: %w", DuplicateTaskErr, lintTask, err)
		}
	}
	return err
}

func (p PgLintStorage) TryTake(ctx context.Context, lockTimeLowerBound time.Time, lockedAt time.Time) (dto.LintTask, error) {
	rows, err := p.Query(ctx, lintTaskTakeSql, lockTimeLowerBound, lockedAt)
	if err != nil {
		return dto.LintTask{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return dto.LintTask{}, NoTasksErr
	}
	if rows.Err() != nil {
		return dto.LintTask{}, rows.Err()
	}
	var (
		lintId              string
		linterId            string
		linterDockerImage   string
		linterDockerShaHash string
		repoId              string
		repoGitUrl          string
		repoGitCommitHash   string
	)
	err = rows.Scan(&lintId, &linterId, &linterDockerImage, &linterDockerShaHash, &repoId, &repoGitUrl, &repoGitCommitHash)
	if err != nil {
		return dto.LintTask{}, err
	}
	lintTask := dto.LintTask{
		Id: lintId,
		Linter: dto.LinterInstance{
			Id:                 linterId,
			DockerImage:        linterDockerImage,
			DockerImageShaHash: linterDockerShaHash,
		},
		Repo: dto.RepoInstance{
			Id:            repoId,
			GitUrl:        repoGitUrl,
			GitCommitHash: repoGitCommitHash,
		},
	}
	return lintTask, nil
}

func (p PgLintStorage) Set(ctx context.Context, lintTask dto.LintTask, lintResult dto.LintResult, lintedAt time.Time) error {
	if len(lintResult.Highlights) > 0 {
		batch := &pgx.Batch{}
		for _, highlight := range lintResult.Highlights {
			batch.Queue(
				lintHighlightsAddSql,
				lintTask.Id,
				highlight.Path,
				highlight.StartLine,
				highlight.EndLine,
				highlight.Explanation,
				highlight.Snippet.StartLine,
				highlight.Snippet.EndLine,
				highlight.Snippet.Code,
			)
		}
		result := p.SendBatch(ctx, batch)
		_, err := result.Exec()
		_ = result.Close()
		if err != nil {
			return fmt.Errorf("failed to insert batch of %v highlights: %w", len(lintResult.Highlights), err)
		}
	}
	_, err := p.Exec(
		ctx,
		lintTaskSetSql,
		lintTask.Id,
		lintResult.Status,
		lintResult.StatusComment,
		lintResult.Duration,
		lintedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update lint result state: %w", err)
	}
	return nil
}
