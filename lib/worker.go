package lib

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/timeout"
	"github.com/sivukhin/gobughunt/storage/db"
)

type Worker struct {
	Storage        *db.Queries
	DockerApi      DockerApi
	Linting        Linting
	IterationDelay time.Duration
	CleanupTimeout time.Duration
	TakeTimeout    time.Duration
	LintTimeout    time.Duration
	UpdateTimeout  time.Duration
	LockDuration   time.Duration
}

func (w Worker) RunForever(ctx context.Context) {
	logging.Logger.Infof(
		"worker started: iterationDelay=%v, cleanupTimeout=%v, takeTimeout=%v, lintTimeout=%v, updateTimeout=%v, lockDuration=%v",
		w.IterationDelay,
		w.CleanupTimeout,
		w.TakeTimeout,
		w.LintTimeout,
		w.UpdateTimeout,
		w.LockDuration,
	)
	periodic := timeout.Periodic(ctx, w.IterationDelay, w.IterationDelay)
	cleanup := timeout.Process("cleanup", periodic, w.CleanupTimeout, func(ctx context.Context, item struct{}, next func(struct{})) error {
		err := w.DockerApi.Cleanup(ctx)
		if err != nil {
			logging.Logger.Errorf("failed to cleanup docker: %v", err)
		} else {
			next(struct{}{})
		}
		return err
	})
	take := timeout.Process("take", cleanup, w.TakeTimeout, func(ctx context.Context, _ struct{}, next func(task dto.LintTask)) error {
		logging.Logger.Infof("worker run single iteration")
		now := time.Now()
		lintTask, err := w.Storage.TryTakeLintTask(ctx, db.TryTakeLintTaskParams{
			LockTimeLowerBound: pgtype.Timestamp{Time: now.Add(-w.LockDuration), Valid: true},
			LockedAt:           pgtype.Timestamp{Time: now, Valid: true},
		})
		if err != nil {
			return err
		}
		logging.Logger.Infof("took single lint task: %+v", lintTask)
		next(dto.LintTask{
			Id: lintTask.LintID,
			Linter: dto.LinterInstance{
				Id:                 lintTask.LinterID,
				DockerImage:        lintTask.LinterDockerImage,
				DockerImageShaHash: lintTask.LinterDockerShaHash,
			},
			Repo: dto.RepoInstance{
				Id:            lintTask.RepoID,
				GitUrl:        lintTask.RepoGitUrl,
				GitCommitHash: lintTask.RepoGitCommitHash,
			},
		})
		return nil
	})
	type lintResult struct {
		task       dto.LintTask
		highlights []dto.LintHighlightSnippet
		duration   time.Duration
		err        error
	}
	lint := timeout.Process("lint", take, w.LintTimeout, func(ctx context.Context, item dto.LintTask, next func(result lintResult)) error {
		startTime := time.Now()
		highlights, err := w.Linting.Run(ctx, item.Repo, item.Linter)
		next(lintResult{task: item, highlights: highlights, err: err, duration: time.Since(startTime)})
		return nil
	})
	update := timeout.Process("update", lint, w.UpdateTimeout, func(ctx context.Context, item lintResult, next func(struct{})) error {
		now := time.Now()
		if errors.Is(item.err, LintSkippedErr) {
			return w.Storage.SetLintTask(ctx, db.SetLintTaskParams{
				LintID:       item.task.Id,
				LintStatus:   db.LintStatusSkipped,
				LintDuration: pgtype.Interval{Microseconds: item.duration.Microseconds(), Valid: true},
				LintedAt:     pgtype.Timestamp{Time: now, Valid: true},
			})
		} else if errors.Is(item.err, LintTempErr) {
			return errors.Join(item.err, w.Storage.SetLintTask(ctx, db.SetLintTaskParams{
				LintID:       item.task.Id,
				LintStatus:   db.LintStatusPending,
				LintDuration: pgtype.Interval{Microseconds: item.duration.Microseconds(), Valid: true},
				LintedAt:     pgtype.Timestamp{Time: now, Valid: true},
			}))
		} else if item.err != nil {
			return errors.Join(item.err, w.Storage.SetLintTask(ctx, db.SetLintTaskParams{
				LintID:            item.task.Id,
				LintStatus:        db.LintStatusFailed,
				LintStatusComment: pgtype.Text{String: item.err.Error(), Valid: true},
				LintDuration:      pgtype.Interval{Microseconds: item.duration.Microseconds(), Valid: true},
				LintedAt:          pgtype.Timestamp{Time: now, Valid: true},
			}))
		}

		params := make([]db.AddLintHighlightParams, 0, len(item.highlights))
		for _, highlight := range item.highlights {
			params = append(params, db.AddLintHighlightParams{
				LintID:           item.task.Id,
				Path:             highlight.Path,
				StartLine:        int32(highlight.StartLine),
				EndLine:          int32(highlight.EndLine),
				Explanation:      highlight.Explanation,
				SnippetStartLine: int32(highlight.Snippet.StartLine),
				SnippetEndLine:   int32(highlight.Snippet.EndLine),
				SnippetCode:      highlight.Snippet.Code,
			})
		}
		var batchErrs []error
		results := w.Storage.AddLintHighlight(ctx, params)
		results.Exec(func(i int, err error) { batchErrs = append(batchErrs, err) })
		if batchErr := errors.Join(batchErrs...); batchErr != nil {
			return fmt.Errorf("failed to update lint highlights: %w", batchErr)
		}

		return w.Storage.SetLintTask(ctx, db.SetLintTaskParams{
			LintID:       item.task.Id,
			LintStatus:   db.LintStatusSucceed,
			LintDuration: pgtype.Interval{Microseconds: item.duration.Microseconds(), Valid: true},
			LintedAt:     pgtype.Timestamp{Time: now, Valid: true},
		})
	})
	timeout.Close(update)
}
