package main

import (
	"os"
	"os/signal"
	"syscall"
)

func SignalHandler(hooks ...func(os.Signal)) <-chan any {
	s := make(chan os.Signal, 1)
	f := make(chan any, 1)

	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer close(f)

		sig := <-s
		for _, hook := range hooks {
			hook(sig)
		}
	}()

	return f
}
