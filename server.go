package main

import (
	"net/http"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	model  SQLModel
	logger Logger

	mux      *chi.Mux
	homeTmpl *template.Template
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

	s.mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		s.home(w, r)
	})

	s.addRoutes()
	s.addTemplates()
	return s, nil
}

func (s *Server) addRoutes() {
	s.mux.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("foo"))
	})
}

func (s *Server) addTemplates() {
	s.homeTmpl = template.Must(template.New("home").Parse(homeTmpl))
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	list := List{
		&Token{
			"1",
			"foo",
			false,
			300,
			false,
			false,
		},
	}

	var data = struct {
		Name       string
		SayHi      bool
		ListTokens List
	}{
		Name:       "david",
		SayHi:      true,
		ListTokens: list,
	}

	err := s.homeTmpl.Execute(w, data)
	if err != nil {
		s.internalError(w, "rendering template", err)
		return
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Cache-Control", "no-cache")
	s.mux.ServeHTTP(w, r)
	s.logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(startTime))
}

func (s *Server) internalError(w http.ResponseWriter, msg string, err error) {
	s.logger.Printf("error %s: %v", msg, err)
	http.Error(w, "error "+msg, http.StatusInternalServerError)
}
