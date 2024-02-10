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
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		iterationCtx, cancel := context.WithTimeout(ctx, w.IterationTimeout)
		err := w.RunOnce(iterationCtx)
		cancel()

		if errors.Is(err, storage.NoTasksErr) {
			logging.Logger.Infof("no ready tasks found, sleeping for %v", w.IterationLongDelay)
			timeout.SleepOrDone(ctx, w.IterationLongDelay)
		} else if err != nil {
			logging.Logger.Errorf("failed single iteration, sleeping for %v: %v", w.IterationShortDelay, err)
			timeout.SleepOrDone(ctx, w.IterationShortDelay)
		} else {
			logging.Logger.Infof("succeed with single iteration, sleeping for %v", w.IterationShortDelay)
			timeout.SleepOrDone(ctx, w.IterationShortDelay)
		}
	}
}

func (w Worker) RunOnce(ctx context.Context) error {
	now := time.Now()
	lintTask, err := w.LintStorage.TryTake(ctx, now.Add(-w.LockDuration), now)
	if err != nil {
		return err
	}

	startLintTime := now
	highlights, err := w.Linting.Run(ctx, lintTask.Repo, lintTask.Linter)
	if errors.Is(err, LintSkippedErr) {
		return w.LintStorage.Set(ctx, lintTask, dto.LintResult{
			LintStatus:   dto.Skipped,
			LintDuration: time.Since(startLintTime),
		}, now)
	} else if err != nil {
		return errors.Join(err, w.LintStorage.Set(ctx, lintTask, dto.LintResult{
			LintStatus:        dto.Failed,
			LintStatusComment: err.Error(),
			LintDuration:      time.Since(startLintTime),
		}, now))
	}
	return w.LintStorage.Set(ctx, lintTask, dto.LintResult{
		LintStatus:   dto.Succeed,
		LintDuration: time.Since(startLintTime),
		Highlights:   highlights,
	}, now)
}
