package timeout

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/sivukhin/gobughunt/lib/logging"
)

func SleepOrDone(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.NewTimer(duration).C:
	}
}

type Trigger[T any] struct {
	Ctx  context.Context
	Data T
	Done chan<- error
}

func Close[T any](sequence <-chan Trigger[T]) {
	done := make(chan struct{})
	go func() {
		for item := range sequence {
			item.Done <- nil
		}
	}()
	<-done
	return
}

func Periodic(ctx context.Context, failDelay time.Duration, successDelay time.Duration) <-chan Trigger[struct{}] {
	ticks := make(chan Trigger[struct{}])
	go func() {
		defer close(ticks)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			done := make(chan error)
			ticks <- Trigger[struct{}]{Ctx: ctx, Done: done}
			err := <-done
			delay := successDelay
			if err != nil {
				delay = failDelay
				logging.Logger.Infof("sleeping for %v after failed iteration: %v", delay, err)
			} else {
				delay = successDelay
				logging.Logger.Infof("sleeping for %v after successful iteration", delay)
			}
			select {
			case <-ctx.Done():
			case <-time.NewTimer(delay).C:
			}
		}
	}()
	return ticks
}

func Process[T, Q any](
	name string,
	sequence <-chan Trigger[T],
	runTimeout time.Duration,
	run func(ctx context.Context, item T, next func(result Q)) error,
) <-chan Trigger[Q] {
	processed := make(chan Trigger[Q])
	go func() {
		defer close(processed)
		for item := range sequence {
			var wg errgroup.Group
			runCtx, cancel := context.WithTimeout(item.Ctx, runTimeout)
			err := run(runCtx, item.Data, func(result Q) {
				wg.Go(func() error {
					done := make(chan error)
					processed <- Trigger[Q]{Ctx: item.Ctx, Data: result, Done: done}
					err := <-done
					if err != nil {
						logging.Logger.Errorf("%v: upstream failed: %v", name, err)
					}
					return err
				})
			})
			cancel()
			item.Done <- errors.Join(err, wg.Wait())
		}
	}()
	return processed
}

func RunForeverAsync(name string, globalCtx context.Context, iterationTimeout time.Duration, run func(ctx context.Context) time.Duration) <-chan struct{} {
	finish := make(chan struct{})
	go func() {
		defer func() { finish <- struct{}{} }()
		for {
			select {
			case <-globalCtx.Done():
				return
			default:
			}
			iterationCtx, cancel := context.WithTimeout(globalCtx, iterationTimeout)
			sleep := run(iterationCtx)
			cancel()
			logging.Logger.Infof("%v: sleeping for %v", name, sleep)
			SleepOrDone(globalCtx, sleep)
		}
	}()
	return finish
}
