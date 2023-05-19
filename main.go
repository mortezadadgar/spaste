package main

import (
	"log"

	"github.com/mortezadadgar/spaste/config"
	"github.com/mortezadadgar/spaste/http"
	"github.com/mortezadadgar/spaste/snippets"
	"github.com/mortezadadgar/spaste/store"
	"github.com/mortezadadgar/spaste/template"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New(cfg.StaticBase+"/templates/", false)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewStore()

	s, err := snippets.NewSnippets(store)
	if err != nil {
		log.Fatal(err)
	}

	server := http.NewServer(cfg, t, s)

	server.Start()
}
