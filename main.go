package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/http"
	"github.com/mortezadadgar/spaste/internal/paste"
	"github.com/mortezadadgar/spaste/internal/sqlite"
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

	db, err := sqlite.New(config)
	if err != nil {
		log.Fatal(err)
	}

	validator := validator.New()

	paste := paste.New(db, config)

	server := http.New(config, template, paste, validator)

	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}

	// wait for user signal
	<-registerSignalNotify()

	err = closeMain(server, db)
	if err != nil {
		log.Println("failed to close program, exiting now...")
		os.Exit(1)
	}
}

func registerSignalNotify() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	return sig
}

type store interface {
	Close() error
}

type server interface {
	Start() error
	Close() error
}

func closeMain(server server, store store) error {
	err := server.Close()
	if err != nil {
		return err
	}

	err = store.Close()
	if err != nil {
		return err
	}

	return nil
}
