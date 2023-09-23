package main

import (
	"log"

	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/paste"
	"github.com/mortezadadgar/spaste/internal/server"
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

	db, err := store.New(config)
	if err != nil {
		log.Fatal(err)
	}

	validator := validator.New()

	paste := paste.New(db, config)

	server := server.New(config, template, paste, validator)

	log.Fatal(server.ListenAndServe())
}
