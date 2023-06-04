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
	sonfig, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	template, err := template.New(sonfig.StaticBase+"/templates/", true)
	if err != nil {
		log.Fatal(err)
	}

	store := store.New()

	snippet := snippets.New(store)

	server := http.New(sonfig, template, snippet)

	server.Start()
}
