package app

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
)

type Server struct {
	srv  *http.Server
	name string
	mux  *serverMux
}

func NewServer(name, add string) *Server {
	mux := &serverMux{ServeMux: http.NewServeMux()}
	return &Server{
		name: name,
		mux:  mux,
		srv: &http.Server{
			Addr:    add,
			Handler: mux,
		},
	}
}

func (s *Server) Handle(patten string, handler http.Handler) {
	s.mux.Handle(patten, handler)
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) rejectReq() {
	s.mux.RejectNewReq()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Printf("server %s is stopping", s.name)
	return s.srv.Shutdown(ctx)
}

type serverMux struct {
	reject int64
	*http.ServeMux
}

func (s *serverMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt64(&s.reject) == 1 {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("service is closed"))
		return
	}
	s.ServeMux.ServeHTTP(w, r)
}

func (s *serverMux) RejectNewReq() {
	atomic.StoreInt64(&s.reject, 1)
}
