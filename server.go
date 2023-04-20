package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	model  SQLModel
	logger Logger

	mux *chi.Mux
	//homeTmpl *template.Template
	//listTmpl *template.Template
}

type Model interface {
	FooBar() error
}

type Logger interface {
	Printf(format string, v ...interface{})
}

func NewServer(
	model SQLModel,
	logger Logger,
) (*Server, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	s := &Server{
		model:  model,
		logger: logger,
		mux:    r,
	}
	s.addRoutes()
	// s.addTemplates()
	return s, nil
}

func (s *Server) addRoutes() {
	s.mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Cache-Control", "no-cache")
	s.mux.ServeHTTP(w, r)
	s.logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(startTime))
}
