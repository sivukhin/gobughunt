package lib

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/lib/storage"
	"github.com/sivukhin/gobughunt/lib/timeout"
)

type Worker struct {
	LintStorage           storage.LintStorage
	DockerApi             DockerApi
	Linting               Linting
	IterationTimeout      time.Duration
	IterationFailDelay    time.Duration
	IterationSuccessDelay time.Duration
	LockDuration          time.Duration
}

func (w Worker) RunForever(ctx context.Context) {
	logging.Logger.Infof(
		"worker started: timeout=%v, shortDelay=%v, longDelay=%v, lockDuration=%v",
		w.IterationTimeout,
		w.IterationFailDelay,
		w.IterationSuccessDelay,
		w.LockDuration,
	)
	periodic := timeout.Periodic(ctx, w.IterationFailDelay, w.IterationSuccessDelay)
	worker := timeout.Process("worker", periodic, w.IterationTimeout, func(ctx context.Context, _ struct{}, next func(struct{})) error {
		err := w.RunOnce(ctx)
		if err != nil && !errors.Is(err, storage.NoTasksErr) {
			return fmt.Errorf("failed single iteration: %w", err)
		} else if errors.Is(err, storage.NoTasksErr) {
			logging.Logger.Infof("no ready tasks found")
		} else {
			logging.Logger.Infof("succeed with single iteration")
		}
		return nil
	})
	timeout.Close(worker)
}

func (w Worker) RunOnce(ctx context.Context) error {
	logging.Logger.Infof("worker run single iteration")
	now := time.Now()
	lintTask, err := w.LintStorage.TryTake(ctx, now.Add(-w.LockDuration), now)
	if err != nil {
		return err
	}
	logging.Logger.Infof("took single lint task: %+v", lintTask)

	err = w.DockerApi.Cleanup(ctx)
	if err != nil {
		logging.Logger.Errorf("failed to cleanup docker: %v", err)
	}

	startLintTime := time.Now()
	highlights, err := w.Linting.Run(ctx, lintTask.Repo, lintTask.Linter)
	if errors.Is(err, LintSkippedErr) {
		return w.LintStorage.Set(ctx, lintTask, dto.LintResult{
			Status:   dto.Skipped,
			Duration: time.Since(startLintTime),
		}, time.Now())
	} else if errors.Is(err, LintTempErr) {
		return errors.Join(err, w.LintStorage.Set(ctx, lintTask, dto.LintResult{
			Status:   dto.Pending,
			Duration: time.Since(startLintTime),
		}, time.Now()))
	} else if err != nil {
		return errors.Join(err, w.LintStorage.Set(ctx, lintTask, dto.LintResult{
			Status:        dto.Failed,
			StatusComment: err.Error(),
			Duration:      time.Since(startLintTime),
		}, time.Now()))
	}
	return w.LintStorage.Set(ctx, lintTask, dto.LintResult{
		Status:     dto.Succeed,
		Duration:   time.Since(startLintTime),
		Highlights: highlights,
	}, now)
}
