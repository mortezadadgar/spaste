package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/modules"
)

type Server struct {
	config    config.Config
	template  template
	paste     paste
	validator validator
	*http.Server
}

type paste interface {
	Get(addr string) (*modules.Paste, error)
	Create(text string, lang string, lineCount int) (string, error)
	Render(m *modules.Paste) (string, error)
}

type template interface {
	Render(w io.Writer, name string, data any) error
}

type validator interface {
	IsBlank(name string, value string)
	IsEqual(name string, value int, expectedValue int)
	Valid() error
}

// New returns a new instance of server.
func New(config config.Config, template template, paste paste, validator validator) *Server {
	s := Server{
		config:    config,
		template:  template,
		paste:     paste,
		validator: validator,
		Server:    prepareServer(config),
	}

	r := chi.NewMux()

	r.Use(s.logger)
	r.Use(s.recoverer)

	fs := http.FileServer(http.Dir("./static"))
	r.With(s.disallowRootFS).
		Handle("/static/*", http.StripPrefix("/static", fs))

	r.Get("/", s.renderIndex)
	r.Post("/paste", s.createPaste)
	r.Get("/{addr:[Aa-zZ.]+}", s.renderPaste)
	r.NotFound(s.notFoundHandler)

	s.Handler = r

	return &s
}

var ErrEmptyCreatePasteBody = errors.New("empty reponse body not allowed for createPaste")

func (s *Server) disallowRootFS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		defer func() {
			log.Printf("[%s %s %s from %s in %dµs]",
				r.Method,
				r.URL.Path,
				r.Proto,
				r.Host,
				time.Since(now).Microseconds())
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("%s\nSTACK: %s", err, debug.Stack())
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) renderIndex(w http.ResponseWriter, r *http.Request) {
	err := s.template.Render(w, "index.page.tmpl", nil)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
	}
}

func (s *Server) createPaste(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		s.serverError(w, r, ErrEmptyCreatePasteBody, http.StatusNotFound)
		return
	}
	defer r.Body.Close()

	var pasteData modules.Paste
	err := json.NewDecoder(r.Body).Decode(&pasteData)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
		return
	}

	s.validator.IsBlank("Text", pasteData.Text)
	s.validator.IsBlank("Lang", pasteData.Lang)
	s.validator.IsEqual("LineCount", pasteData.LineCount, 0)
	err = s.validator.Valid()
	if err != nil {
		s.serverError(w, r, err, http.StatusBadRequest)
		return
	}

	address, err := s.paste.Create(pasteData.Text, pasteData.Lang, pasteData.LineCount)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
		return
	}

	s.validator.IsBlank("Address", address)
	err = s.validator.Valid()
	if err != nil {
		s.serverError(w, r, err, http.StatusBadRequest)
		return
	}

	var addressData modules.Paste
	addressData.Address = address

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(addressData)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) renderPaste(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "addr")

	if len(address) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	paste, err := s.paste.Get(address)
	switch {
	case paste == nil:
		s.notFoundHandler(w, r)
		return
	case err != nil:
		s.serverError(w, r, err, http.StatusInternalServerError)
		return
	}

	renderedPaste, err := s.paste.Render(paste)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
		return
	}

	t := &modules.TemplateData{
		TextHighlighted: renderedPaste,
		Address:         fmt.Sprintf("%s/%s", r.Host, paste.Address),
		LineCount:       paste.LineCount,
		Lang:            paste.Lang,
	}

	err = s.template.Render(w, "paste.page.tmpl", t)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
	}
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error, code int) {
	log.Printf("ERROR: method: %s, url: %s, code: %d, err: %s", r.Method, r.URL.Path, code, err)
	http.Error(w, err.Error(), code)
}

func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	t := modules.TemplateData{
		Message:     "404 Page not found",
		IncludeHome: true,
	}
	w.WriteHeader(http.StatusNotFound)

	err := s.template.Render(w, "message.page.tmpl", t)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
	}
}

func prepareServer(config config.Config) *http.Server {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Address),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			cancel()
			os.Exit(1)
		}
	}()

	return server
}
