package main

import (
	"context"
	"log"
	"time"

	"github.com/ppd0705/go-training-camp/week1/homework/app"
)

func main() {
	svc1 := app.NewServer("biz", "localhost:8082")
	svc2 := app.NewServer("admin", "localhost:8081")
	a := app.NewApp(
		[]*app.Server{svc1, svc2},
		app.WithShutDownTimeout(time.Second*30),
		app.WithWaitTime(time.Second*3),
		app.WithCallbackTimeout(time.Second*10),
		app.WithShutdownCallbacks(
			func(ctx context.Context) {
				log.Printf("callback1 called")
			},
			func(ctx context.Context) {
				log.Printf("callback2 called")
			},
		),
	)
	a.StartAndServe()
	log.Println("app exit...")
}
