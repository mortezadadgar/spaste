package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/paste"
	"github.com/mortezadadgar/spaste/internal/template"
	"github.com/mortezadadgar/spaste/internal/validator"
)

type server struct {
	config    config.Config
	template  template.Template
	paste     paste.Paste
	validator *validator.Validator
	*http.Server
}

// New returns a new instance of server.
func New(config config.Config, template template.Template, paste paste.Paste, validator *validator.Validator) *server {
	s := server{
		config:    config,
		template:  template,
		paste:     paste,
		validator: validator,
		Server:    prepareServer(config),
	}

	r := chi.NewMux()

	// r.Use(s.logger)
	r.Use(s.recoverer)

	fs := http.FileServer(http.Dir("./static"))
	r.With(s.disallowRootFS).
		Handle("/static/*", http.StripPrefix("/static", fs))

	r.Get("/", s.renderIndex)
	r.Post("/paste", s.createPaste)
	r.Get("/{addr:.+}", s.renderPaste)
	r.NotFound(s.notFoundHandler)

	s.Handler = r

	return &s
}

func (s *server) Start() error {
	l, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}

	log.Printf("Started listening on %s\n", s.config.Address)

	go s.Serve(l)
	return nil
}

var ErrEmptyCreatePasteBody = errors.New("empty reponse body not allowed for createPaste")

func (s *server) disallowRootFS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) logger(next http.Handler) http.Handler {
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

func (s *server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("%s\nSTACK: %s", err, debug.Stack())
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *server) renderIndex(w http.ResponseWriter, r *http.Request) {
	err := s.template.Render(w, "index.page.tmpl", nil)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
	}
}

func (s *server) serverError(w http.ResponseWriter, r *http.Request, err error, code int) {
	log.Printf("ERROR: method: %s, url: %s, code: %d, err: %s", r.Method, r.URL.Path, code, err)
	http.Error(w, err.Error(), code)
}

func (s *server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	t := paste.TemplateData{
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
