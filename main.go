package main

import (
	"log"

	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/http"
	"github.com/mortezadadgar/spaste/internal/snippet"
	"github.com/mortezadadgar/spaste/internal/store"
	"github.com/mortezadadgar/spaste/internal/template"
	"github.com/mortezadadgar/spaste/internal/validator"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	template, err := template.New(config.StaticBase+"/templates/", true)
	if err != nil {
		log.Fatal(err)
	}

	db, err := store.NewSQLiteStore(config)
	if err != nil {
		log.Fatal(err)
	}

	validator := validator.New()

	snippet := snippet.New(db, validator)

	server := http.New(config, template, snippet)

	server.Start()
}
