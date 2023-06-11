package http

// TODO:
//      - addSnippet: unlikely: check on repetitive addresses

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/go-chi/chi"
	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/log"
	"github.com/mortezadadgar/spaste/internal/snippet"
	"github.com/mortezadadgar/spaste/internal/template"
)

type Server struct {
	cfg      config.Config
	template template.Template
	snippet  snippet.Snippet
}

// New returns a new instance of server.
func New(cfg config.Config, template template.Template, snippet snippet.Snippet) Server {
	return Server{
		cfg:      cfg,
		template: template,
		snippet:  snippet,
	}
}

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

func (s *Server) renderIndex(w http.ResponseWriter, _ *http.Request) {
	err := s.template.Render(w, "index.page.tmpl", nil)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, false)
	}
}

func (s *Server) createSnippet(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text      string `json:"text"`
		Lang      string `json:"lang"`
		LineCount int    `json:"lineCount"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
		return
	}

	var output struct {
		Address string `json:"address"`
	}

	output.Address, err = s.snippet.MakeAddress(s.cfg.AddressLength)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
		return
	}
	output.Address = fmt.Sprintf("%s.%s", output.Address, input.Lang)

	err = s.snippet.Create(input.Text, input.Lang, input.LineCount, output.Address)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
		return
	}

	resp, err := json.Marshal(output)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
	}
}

// we tend to ignore errors from this function to reduce verbosity.
func (s *Server) serverError(w http.ResponseWriter, isFromJS bool) error {
	t := template.Data{
		Message:     "404 Page not found",
		IncludeHome: true,
	}

	// I dont like this portion of code myself
	// but had no other choice than write it
	// directly to stream and return from here
	if !isFromJS {
		err := s.template.Render(w, "message.page.tmpl", t)
		if err != nil {
			return fmt.Errorf("failed to render non-js server error: %s", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	buf := new(bytes.Buffer)
	err := s.template.Render(buf, "message.page.tmpl", t)
	if err != nil {
		return fmt.Errorf("failed to render js server error: %s", err)
	}

	var errorOutput struct {
		HTML string `json:"errorHTML"`
	}
	errorOutput.HTML = buf.String()

	resp, err := json.Marshal(errorOutput)
	if err != nil {
		return fmt.Errorf("failed to encode server error: %s", err)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		return fmt.Errorf("failed to write server error: %s", err)
	}

	return nil
}

func (s *Server) renderSnippet(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "addr")

	if len(address) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	snippet, err := s.snippet.Get(address)
	switch {
	case snippet == nil:
		// log.Printf("snippet %s not found", address)
		s.notFoundHandler(w, r)
		return
	case err != nil:
		log.Errorln(err)
		s.serverError(w, false)
		return
	}

	lexer := lexers.Get(snippet.Lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	formatter := html.New(html.WithClasses(true))

	iterator, err := lexer.Tokenise(nil, string(snippet.Text))
	if err != nil {
		log.Errorln(err)
		s.serverError(w, false)
		return
	}

	style := styles.Get("doom-one")
	if style == nil {
		style = styles.Fallback
	}

	buf := new(bytes.Buffer)
	err = formatter.Format(buf, style, iterator)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, false)
		return
	}

	t := &template.Data{
		TextHighlighted: template.ToHTML(buf.String()),
		Address:         fmt.Sprintf("%s/%s", r.Host, snippet.Address),
		LineCount:       snippet.LineCount,
		Lang:            snippet.Lang,
	}

	err = s.template.Render(w, "snippet.page.tmpl", t)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, false)
	}
}

func (s *Server) notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	t := template.Data{
		Message:     "404 Page not found",
		IncludeHome: true,
	}
	w.WriteHeader(http.StatusNotFound)

	err := s.template.Render(w, "message.page.tmpl", t)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, false)
	}
}

func (s *Server) routes() *chi.Mux {
	r := chi.NewMux()
	r.Use(s.logger)
	r.Use(s.recoverer)

	fs := http.FileServer(http.Dir("./static"))
	r.With(s.disallowRootFS).
		Handle("/static/*", http.StripPrefix("/static", fs))

	r.Get("/", s.renderIndex)

	r.Post("/snippet", s.createSnippet)

	r.Get("/{addr:[Aa-zZ.]+}", s.renderSnippet)

	r.NotFound(s.notFoundHandler)

	return r
}

// Start setup the server and eventually start it.
func (s *Server) Start() {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%s", s.cfg.Address),
		Handler:      s.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Errorln("unable to shutdown the server, exiting!")
			cancel()
			os.Exit(1)
		}
	}()

	// This is a lie!
	log.Printf("Started server on address %s\n", s.cfg.Address)

	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
