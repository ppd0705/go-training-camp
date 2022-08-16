package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type App struct {
	servers         []*Server
	cbs             []ShutdownCallback
	cbTimeout       time.Duration
	shutdownTimeout time.Duration
	waitTime        time.Duration
}

func NewApp(servers []*Server, opts ...Option) *App {
	app := &App{
		servers:         servers,
		cbTimeout:       time.Second * 3,
		waitTime:        time.Second * 10,
		shutdownTimeout: time.Second * 30,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func (a *App) StartAndServe() {
	eg, ctx := errgroup.WithContext(context.Background())
	for _, s := range a.servers {
		s := s
		eg.Go(func() error {
			if err := s.Start(); !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		})
	}
	sg := NewSigHandler(a.shutdownTimeout, a.shutdown)
	eg.Go(func() error {
		sg.Run(ctx)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Printf("err: %v\n", err)
	}

}

func (a *App) shutdown() {
	log.Println("reject new request")
	for _, s := range a.servers {
		s.rejectReq()
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.waitTime)
	defer cancel()
	var wg sync.WaitGroup
	for _, s := range a.servers {
		wg.Add(1)
		s := s
		go func() {
			_ = s.Stop(ctx)
			wg.Done()
		}()
	}
	wg.Wait()
	log.Println("all servers closed")

	ctx, cancel = context.WithTimeout(context.Background(), a.cbTimeout)
	defer cancel()
	for _, cb := range a.cbs {
		wg.Add(1)
		cb := cb
		go func() {
			cb(ctx)
			wg.Done()
		}()
	}
	wg.Wait()
	log.Println("all cbs called")

	a.writeBack()
	log.Println("cache flushed")

	a.closeDB()
	log.Println("db closed")
}

func (a *App) writeBack() {
	time.Sleep(time.Second)
}

func (a *App) closeDB() {
	time.Sleep(time.Second * 1)
}

type Option func(app *App)

type ShutdownCallback func(ctx context.Context)

func WithShutdownCallbacks(cbs ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = cbs
	}
}

func WithShutDownTimeout(t time.Duration) Option {
	return func(app *App) {
		app.shutdownTimeout = t
	}
}

func WithWaitTime(t time.Duration) Option {
	return func(app *App) {
		app.waitTime = t
	}
}

func WithCallbackTimeout(t time.Duration) Option {
	return func(app *App) {
		app.cbTimeout = t
	}
}
