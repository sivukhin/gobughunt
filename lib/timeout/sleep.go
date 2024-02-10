package timeout

import (
	"context"
	"time"

	"github.com/sivukhin/gobughunt/lib/logging"
)

func SleepOrDone(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.NewTimer(duration).C:
	}
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
