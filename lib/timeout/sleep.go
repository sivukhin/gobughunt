package timeout

import (
	"context"
	"time"
)

func SleepOrDone(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.NewTimer(duration).C:
	}
}
