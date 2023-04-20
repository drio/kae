package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	model  Model
	logger Logger

	mux      *chi.Mux
	homeTmpl *template.Template
}

type Logger interface {
	Printf(format string, v ...interface{})
}

func NewServer(
	model Model,
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
	s.addTemplates()
	return s, nil
}

func (s *Server) addRoutes() {
	s.mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		s.home(w, r)
	})

	s.mux.Post("/newtoken", s.createToken)
	s.mux.Get("/hb/{token}", s.hbToken)
}

func (s *Server) createToken(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		//just reload home page
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	interval := strings.TrimSpace(r.FormValue("interval"))
	if interval == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	intInterval, err := strconv.Atoi(interval)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	_, err = s.model.CreateToken(name, intInterval)
	if err != nil {
		s.internalError(w, "creating new token", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) hbToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		s.badRequestError(w, "token not provided", nil)
		return
	}

	id, err := s.model.GetIdFromToken(token)
	if err != nil {
		s.internalError(w, "checking for token", err)
		return
	}

	if id == 0 {
		w.Write([]byte(fmt.Sprintf("ok t=%s (nd)", token)))
		return
	}

	err = s.model.InsertHeartBeat(id)
	if err != nil {
		s.internalError(w, "heartbeat", err)
		return
	}

	// respond to the client
	w.Write([]byte(fmt.Sprintf("ok t=%s", token)))
}

func (s *Server) addTemplates() {
	s.homeTmpl = template.Must(template.New("home").Parse(homeTmpl))
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	list, err := s.model.GetTokens()
	if err != nil {
		s.internalError(w, "rendering home template", err)
		return
	}

	var data = struct {
		Name   string
		SayHi  bool
		Tokens ListTokens
	}{
		Name:   "david",
		SayHi:  false,
		Tokens: list,
	}

	err = s.homeTmpl.Execute(w, data)
	if err != nil {
		s.internalError(w, "rendering home template", err)
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

func (s *Server) badRequestError(w http.ResponseWriter, msg string, err error) {
	http.Error(w, "error "+msg, http.StatusBadRequest)
}
