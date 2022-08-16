package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var defaultSignals = []os.Signal{syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM}

type SigHandler struct {
	timeout  time.Duration
	callback func()
	c        chan struct{}
}

func NewSigHandler(timeout time.Duration, callback func(), sigs ...os.Signal) *SigHandler {
	s := &SigHandler{
		c:        make(chan struct{}, 1),
		callback: callback,
		timeout:  timeout,
	}
	if len(sigs) == 0 {
		sigs = defaultSignals
	}
	handleC := make(chan os.Signal, 2)
	signal.Notify(handleC, sigs...)
	go func() {
		<-handleC
		close(s.c)
		<-handleC
		os.Exit(1)
	}()
	return s
}

func (s *SigHandler) Run(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-s.c:
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	go func() {
		<-timeoutCtx.Done()
		log.Printf("shutdown timeout")
		os.Exit(1)
	}()
	s.callback()
}
