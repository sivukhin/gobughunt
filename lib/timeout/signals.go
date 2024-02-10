package timeout

import (
	"context"
	"os"
	"os/signal"
)

func SignalsCtx(signals ...os.Signal) context.Context {
	notification := make(chan os.Signal, 1)
	signal.Notify(notification, signals...)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-notification
		cancel()
	}()
	return ctx
}
