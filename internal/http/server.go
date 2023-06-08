package http

// TODO:
//      - logger: try to add status code. ref: https://ndersson.me/post/capturing_status_code_in_net_http/
//      - addSnippet: unlikely: check on repetitive addresses
//      - addSnippet: append filetype when formats when syntax highlighting in available
//		- addSnippet: show a popup on empty text
//      - strings builder?

import (
	"bytes"
	"context"
	"encoding/json"
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

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/go-chi/chi"
	"github.com/mortezadadgar/spaste/internal/config"
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

func disallowRootFS(next http.Handler) http.Handler {
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
				err := s.template.Render(w, "error.page.tmpl", nil)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("%s\nSTACK: %s", err, debug.Stack())
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) ServerError(w http.ResponseWriter, _ *http.Request) {
	err := s.template.Render(w, "error.page.tmpl", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) indexHandler(w http.ResponseWriter, _ *http.Request) {
	err := s.template.Render(w, "index.page.tmpl", nil)
	if err != nil {
		log.Fatal(err)
	}
}

type envelope struct {
	Snippet any `json:"snippet"`
}

func (s *Server) addSnippet(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var input struct {
		Text      string `json:"text"`
		Lang      string `json:"lang"`
		LineCount int    `json:"lineCount"`
	}

	err = json.Unmarshal(b, &envelope{&input})
	if err != nil {
		log.Println(err)
		return
	}

	// TODO: validation

	var output struct {
		Addr string `json:"addr"`
	}

	output.Addr, err = s.snippet.MakeAddress(int64(s.cfg.AddressLength))
	if err != nil {
		log.Println(err)
		return
	}

	output.Addr = fmt.Sprintf("%s.%s", output.Addr, input.Lang)

	resp, err := json.Marshal(envelope{output})
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		log.Println(err)
		return
	}

	err = s.snippet.Add(input.Text, input.Lang, input.LineCount, output.Addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s, Text: %s, Lang: %s\n", output.Addr, input.Text, input.Lang)
}

func (s *Server) renderUserSnippet(w http.ResponseWriter, r *http.Request) {
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
		log.Fatal(err)
	}

	lexer := lexers.Get(snippet.Lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	formatter := html.New(html.WithClasses(true))

	iterator, err := lexer.Tokenise(nil, string(snippet.Text))
	if err != nil {
		log.Fatal(err)
	}

	style := styles.Get("doom-one")
	if style == nil {
		style = styles.Fallback
	}

	buf := new(bytes.Buffer)
	err = formatter.Format(buf, style, iterator)
	if err != nil {
		log.Fatal(err)
	}

	// convert from string to template.HTML
	template.Data.TextHighlighted = template.ToHTML(buf.String())
	template.Data.Address = fmt.Sprintf("%s/%s", r.Host, snippet.Address)
	template.Data.LineCount = snippet.LineCount

	err = s.template.Render(w, "snippet.page.tmpl", template.Data)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	err := s.template.Render(w, "404.page.tmpl", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) routes() *chi.Mux {
	r := chi.NewMux()
	r.Use(s.logger)
	r.Use(s.recoverer)

	fs := http.FileServer(http.Dir("./static"))
	r.With(disallowRootFS).
		Handle("/static/*", http.StripPrefix("/static", fs))

	r.Get("/", s.indexHandler)

	r.Route("/snippets", func(r chi.Router) {
		r.HandleFunc("/", s.addSnippet)
	})

	r.Get("/{addr:[Aa-zZ.]+}", s.renderUserSnippet)

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
			log.Println("unable to shutdown the server, exiting!")
			cancel()
			os.Exit(1)
		}
	}()

	log.Printf("Started server on address %s\n", s.cfg.Address)

	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}