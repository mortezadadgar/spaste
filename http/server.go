package http

// TODO:
//      - logger: try to add status code. ref: https://ndersson.me/post/capturing_status_code_in_net_http/
//      - addSnippet: unlikely: check on repetitive addresses
//      - addSnippet: append filetype when formats when syntax highlighting in available

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/mortezadadgar/spaste/config"
	"github.com/mortezadadgar/spaste/snippets"
	"github.com/mortezadadgar/spaste/template"
)

type server struct {
	cfg      *config.Config
	template *template.Template
	snippet  *snippets.Snippet
}

// New returns a new instance of server
func New(cfg *config.Config, template *template.Template, snippet *snippets.Snippet) server {
	return server{
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

// "GET http://localhost:8080/ HTTP/1.1" from 127.0.0.1:58300 - 200 2057B in 96.831µs
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
				s.template.Render(w, "error.page.tmpl", nil)
				log.Printf("%s\nSTACK: %s", err, debug.Stack())
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *server) indexHandler(w http.ResponseWriter, _ *http.Request) {
	err := s.template.Render(w, "ndex.page.tmpl", nil)
	if err != nil {
		log.Fatal(err)
	}
}

type envelope struct {
	Snippet interface{} `json:"snippet"`
}

func (s *server) addSnippet(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var input struct {
		Text string `json:"text"`
	}

	err = json.Unmarshal(b, &envelope{&input})
	if err != nil {
		log.Println(err)
		return
	}

	var output struct {
		Addr string `json:"addr"`
	}

	output.Addr, err = genRandAddr(10)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := json.Marshal(envelope{output})
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)

	id, err := s.snippet.Add(input.Text, output.Addr)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("%d: %s, Text: %s\n", id, output.Addr, input.Text)
}

func (s *server) renderUserSnippet(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "addr")

	if len(addr) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	snippet := s.snippet.Get(addr)
	if snippet == nil {
		s.notFoundHandler(w, r)
		return
	}

	err := s.template.Render(w, "ndex.page.tmpl", snippet)
	if err != nil {
		log.Println(err)
	}
}

func (s *server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	err := s.template.Render(w, "404.page.tmpl", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *server) routes() *chi.Mux {
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

	r.Get("/{addr:[Aa-zZ]+}", s.renderUserSnippet)

	r.NotFound(s.notFoundHandler)

	return r
}

// Start setup the server and eventually start it
func (s *server) Start() {
	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", s.cfg.Address),
		Handler: s.routes(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		close(sig)
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Println("unable to shutdown the server!")
		}
	}()

	log.Printf("Started server on address %s\n", s.cfg.Address)

	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func genRandAddr(len int64) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buffer := make([]byte, len)
	for i := range buffer {
		r, err := rand.Int(rand.Reader, big.NewInt(len))
		if err != nil {
			return "", err
		}
		buffer[i] = letters[r.Int64()]
	}

	return string(buffer), nil
}
