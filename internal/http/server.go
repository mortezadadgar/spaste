package http

// TODO:
//      - addSnippet: unlikely: check on repetitive addresses

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

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
	"github.com/mortezadadgar/spaste/internal/snippets"
	"github.com/mortezadadgar/spaste/internal/template"
)

type Server struct {
	cfg      *config.Config
	template *template.Template
	snippet  *snippets.Snippet
}

// New returns a new instance of server.
func New(cfg *config.Config, template *template.Template, snippet *snippets.Snippet) *Server {
	return &Server{
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

type envelope struct {
	Snippet any `json:"snippet"`
}

func (s *Server) addSnippet(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
		return
	}

	var input struct {
		Text      string `json:"text"`
		Lang      string `json:"lang"`
		LineCount int    `json:"lineCount"`
	}

	err = json.Unmarshal(b, &envelope{&input})
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
		return
	}

	// input.Text is being checked from js.
	if len(input.Lang) < 1 || input.LineCount < 1 {
		log.Errorln("input.Lang or input.LineCount validation error")
		s.serverError(w, true)
		return
	}

	var output struct {
		Addr string `json:"addr"`
	}

	output.Addr, err = s.snippet.MakeAddress(int64(s.cfg.AddressLength))
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
		return
	}

	output.Addr = fmt.Sprintf("%s.%s", output.Addr, input.Lang)

	err = s.snippet.Add(input.Text, input.Lang, input.LineCount, output.Addr)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, true)
		return
	}

	resp, err := json.Marshal(envelope{output})
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
		return
	}
}

// we tend to ignore errors from this function to reduce verbosity.
func (s *Server) serverError(w http.ResponseWriter, isJS bool) {
	template.Data.Message = "Internal server error :("
	template.Data.IncludeHome = false

	// I dont like this portion of code myself
	// but had no other choice than write it
	// directly to stream and return from here
	if !isJS {
		_ = s.template.Render(w, "message.page.tmpl", template.Data)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	_ = s.template.Render(buf, "message.page.tmpl", template.Data)

	var errorOutput struct {
		HTML string `json:"errHTML"`
	}
	errorOutput.HTML = buf.String()

	resp, _ := json.Marshal(errorOutput)

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}

func (s *Server) renderSnippet(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "addr")

	if len(addr) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	snippet, err := s.snippet.Get(addr)
	switch {
	case snippet == nil:
		s.notFoundHandler(w, r)
		return
	case err != nil:
		log.Errorln(err)
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

	// convert from string to template.HTML
	template.Data.TextHighlighted = template.ToHTML(buf.String())
	template.Data.Address = fmt.Sprintf("%s/%s", r.Host, snippet.Address)
	template.Data.LineCount = snippet.LineCount
	template.Data.Lang = snippet.Lang

	err = s.template.Render(w, "snippet.page.tmpl", template.Data)
	if err != nil {
		log.Errorln(err)
		s.serverError(w, false)
	}
}

func (s *Server) notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	template.Data.Message = "404 Page not found"
	template.Data.IncludeHome = true
	w.WriteHeader(http.StatusNotFound)

	err := s.template.Render(w, "message.page.tmpl", template.Data)
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

	r.Post("/snippets", s.addSnippet)

	r.Get("/{addr:[Aa-zZ.]+}", s.renderSnippet)

	r.NotFound(s.notFoundHandler)

	return r
}

// Start setup the server and eventually start it.
func (s *Server) Start() {
	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", s.cfg.Address),
		Handler: s.routes(),
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		close(sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Errorln("unable to shutdown the server, exiting!")
			cancel()
			os.Exit(1)
		}
	}()

	log.Printf("Started server on address %s\n", s.cfg.Address)

	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
