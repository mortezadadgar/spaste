package http

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/mortezadadgar/spaste/config"
	"github.com/mortezadadgar/spaste/snippets"
	"github.com/mortezadadgar/spaste/template"
)

type templateData struct {
	Text string // TODO: do we need a bigger type
	Addr string
}

type server struct {
	cfg          *config.Config
	template     *template.Template
	snippet      *snippets.Snippet
	templateData *templateData
}

func NewServer(cfg *config.Config, template *template.Template, snippet *snippets.Snippet) server {
	return server{
		cfg:      cfg,
		template: template,
		snippet:  snippet,
	}
}

// - midlewares -

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
// TODO: try to add status code. ref: https://ndersson.me/post/capturing_status_code_in_net_http/
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
				w.Header().Set("Connection", "close")
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("%s\nSTACK: %s", err, debug.Stack())
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// - routes -

func (s *server) indexHandler(w http.ResponseWriter, _ *http.Request) {
	err := s.template.Render(w, "index.page.tmpl", nil)
	if err != nil {
		log.Fatal(err)
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
	// return "hbhhdifdhg", nil // HACK
}

type envelope struct {
	Snippet interface{} `json:"snippet"`
}

// BAD BAD BAD
var t templateData

// TODO:
// - unlikley: check on repetitive addresses
func (s *server) addSnippet(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var input struct {
		Text string `json:"text"`
	}

	err = json.Unmarshal(b, &envelope{&input})
	if err != nil {
		log.Fatal(err)
	}

	var output struct {
		Addr string `json:"addr"`
	}

	// TODO: append filetype when formats when syntax highlighting in available
	output.Addr, err = genRandAddr(10)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := json.Marshal(envelope{output})
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)

	id, err := s.snippet.Add(input.Text, output.Addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%d: %s\n", id, output.Addr)

	t = templateData{
		Addr: output.Addr,
		Text: input.Text,
	}
}

// TODO:
//   - handle errors
//   - multiple requests
func (s *server) getSnippet(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "id")

	id, err := strconv.Atoi(param)
	if err != nil {
		log.Fatal(err)
	}

	snippet, err := s.snippet.Get(id)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(snippet)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(data))
}

func (s *server) renderUserSnippet(w http.ResponseWriter, r *http.Request) {
	// param := chi.URLParam(r, "url")

	err := s.template.Render(w, "index.page.tmpl", t)
	if err != nil {
		log.Fatal(err)
	}
}

type templateDate struct {
	Text string
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
		r.Get("/{id}", s.getSnippet)
	})

	r.Get("/{url:[Aa-zZ]+}", s.renderUserSnippet)

	return r
}

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
			log.Fatal("unable to shutdown the server!")
		}
	}()

	log.Printf("Started server on address %s", s.cfg.Address)

	log.Fatal(srv.ListenAndServe())
}

// - error handling -

func logError(r *http.Request, message string) {
	log.Printf("ERROR: %s %s: %s", r.Method, r.URL.Path, message)
}

func (s *server) Error(w http.ResponseWriter, r *http.Request, code int, message error) {
	// hide server errors from users
	if code == http.StatusInternalServerError {
		logError(r, message.Error())
		message = errors.New("internal server error")
	}

	http.Error(w, message.Error(), code)
}
