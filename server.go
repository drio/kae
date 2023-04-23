package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type ServerOpts struct {
	model          Model
	logger         Logger
	authMiddleware func(next http.Handler) http.Handler
}

type Server struct {
	model  Model
	logger Logger

	mux            *chi.Mux
	homeTmpl       *template.Template
	authMiddleware func(next http.Handler) http.Handler
}

type Logger interface {
	Printf(format string, v ...interface{})
}

type Model interface {
	CreateToken(string, string, int) (string, error)
	GetTokens() (ListTokens, error)
	GetIdFromToken(string) (int, error)
	InsertHeartBeat(int) error
	LastHeartBeat(int) (time.Time, error)
	Fire(int, bool) error
	Disable(int, bool) error
	Remove(int) error
}

func NewServer(opts ServerOpts) (*Server, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	s := &Server{
		model:          opts.model,
		logger:         opts.logger,
		mux:            r,
		authMiddleware: opts.authMiddleware,
	}

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(r, "/assets", filesDir)

	s.addRoutes()
	s.addTemplates()
	return s, nil
}

func (s *Server) addRoutes() {
	s.mux.Get("/hb/{token}", s.hbToken)

	// These have to be protected
	m := s.authMiddleware
	s.mux.Method("get", "/", m(http.HandlerFunc(s.home)))
	s.mux.Method("post", "/newtoken", m(http.HandlerFunc(s.createToken)))
	s.mux.Method("get", "/{action:enable|disable}/{id}", m(http.HandlerFunc(s.updateDisable)))
	s.mux.Method("get", "/delete/{id}", m(http.HandlerFunc(s.remove)))
}

func (s *Server) remove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		s.badRequestError(w, "token id not provided", nil)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		s.internalError(w, "converting token id to int", err)
		return
	}

	err = s.model.Remove(intID)
	if err != nil {
		s.internalError(w, "deleting token", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) updateDisable(w http.ResponseWriter, r *http.Request) {
	action := chi.URLParam(r, "action")

	id := chi.URLParam(r, "id")
	if id == "" {
		s.badRequestError(w, "token id not provided", nil)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		s.internalError(w, "converting token id to int", err)
		return
	}

	if action == "enable" {
		err = s.model.Disable(intID, false)
	} else {
		err = s.model.Disable(intID, true)
	}
	if err != nil {
		s.internalError(w, "disabling token", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) createToken(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	desc := r.FormValue("description")
	if desc == "" {
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

	_, err = s.model.CreateToken(name, desc, intInterval)
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
		_, err = w.Write([]byte(fmt.Sprintf("ok t=%s (nd)", token)))
		if err != nil {
			s.internalError(w, "writing back to the user", err)
		}
		return
	}

	err = s.model.InsertHeartBeat(id)
	if err != nil {
		s.internalError(w, "heartbeat", err)
		return
	}

	// respond to the client
	_, err = w.Write([]byte(fmt.Sprintf("ok t=%s", token)))
	if err != nil {
		s.internalError(w, "writing back to the user", err)
	}

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

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
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

func noAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
