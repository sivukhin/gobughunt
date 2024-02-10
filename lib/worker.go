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
	LintStorage         storage.LintStorage
	Linting             Linting
	IterationTimeout    time.Duration
	IterationShortDelay time.Duration
	IterationLongDelay  time.Duration
	LockDuration        time.Duration
}

func (w Worker) RunForever(ctx context.Context) {
	logging.Logger.Infof(
		"worker started: timeout=%v, shortDelay=%v, longDelay=%v, lockDuration=%v",
		w.IterationTimeout,
		w.IterationShortDelay,
		w.IterationLongDelay,
		w.LockDuration,
	)
	<-timeout.RunForeverAsync("worker", ctx, w.IterationTimeout, func(ctx context.Context) time.Duration {
		err := w.RunOnce(ctx)
		if errors.Is(err, storage.NoTasksErr) {
			logging.Logger.Infof("no ready tasks found, sleeping for %v", w.IterationLongDelay)
			return w.IterationLongDelay
		} else if err != nil {
			logging.Logger.Errorf("failed single iteration, sleeping for %v: %v", w.IterationShortDelay, err)
			return w.IterationShortDelay
		} else {
			logging.Logger.Infof("succeed with single iteration, sleeping for %v", w.IterationShortDelay)
			return w.IterationShortDelay
		}
	})
}

func (w Worker) RunOnce(ctx context.Context) error {
	now := time.Now()
	lintTask, err := w.LintStorage.TryTake(ctx, now.Add(-w.LockDuration), now)
	if err != nil {
		return err
	}

	startLintTime := time.Now()
	highlights, err := w.Linting.Run(ctx, lintTask.Repo, lintTask.Linter)
	if errors.Is(err, LintSkippedErr) {
		return w.LintStorage.Set(ctx, lintTask, dto.LintResult{
			LintStatus:   dto.Skipped,
			LintDuration: time.Since(startLintTime),
		}, time.Now())
	} else if errors.Is(err, LintTempErr) {
		return errors.Join(err, w.LintStorage.Set(ctx, lintTask, dto.LintResult{
			LintStatus:   dto.Pending,
			LintDuration: time.Since(startLintTime),
		}, time.Now()))
	} else if err != nil {
		return errors.Join(err, w.LintStorage.Set(ctx, lintTask, dto.LintResult{
			LintStatus:        dto.Failed,
			LintStatusComment: err.Error(),
			LintDuration:      time.Since(startLintTime),
		}, time.Now()))
	}
	return w.LintStorage.Set(ctx, lintTask, dto.LintResult{
		LintStatus:   dto.Succeed,
		LintDuration: time.Since(startLintTime),
		Highlights:   highlights,
	}, now)
}
