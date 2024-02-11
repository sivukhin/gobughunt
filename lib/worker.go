package lib

import (
	"context"
	"errors"
	"time"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/storage"
	"github.com/sivukhin/gobughunt/lib/timeout"
)

type Worker struct {
	LintStorage    storage.LintStorage
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
		lintTask, err := w.LintStorage.TryTake(ctx, now.Add(-w.LockDuration), now)
		if err != nil {
			return err
		}
		logging.Logger.Infof("took single lint task: %+v", lintTask)
		next(lintTask)
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
			return w.LintStorage.Set(ctx, item.task, dto.LintResult{
				Status:   dto.Skipped,
				Duration: item.duration,
			}, now)
		} else if errors.Is(item.err, LintTempErr) {
			return errors.Join(item.err, w.LintStorage.Set(ctx, item.task, dto.LintResult{
				Status:   dto.Pending,
				Duration: item.duration,
			}, now))
		} else if item.err != nil {
			return errors.Join(item.err, w.LintStorage.Set(ctx, item.task, dto.LintResult{
				Status:        dto.Failed,
				StatusComment: item.err.Error(),
				Duration:      item.duration,
			}, now))
		}
		return w.LintStorage.Set(ctx, item.task, dto.LintResult{
			Status:     dto.Succeed,
			Duration:   item.duration,
			Highlights: item.highlights,
		}, now)
	})
	timeout.Close(update)
}
