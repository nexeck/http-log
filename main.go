package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

type Config struct {
	Debug bool `short:"d" long:"debug" description:"enable exception display and pprof endpoints (warn: dangerous)"`
	Quiet bool `short:"q" long:"quiet" description:"disable verbose output"`
	HTTP  struct {
		Bind string `short:"b" long:"bind" description:"address and port to bind to" default:":8080"`
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("URL.Path %s | Body: %s", r.URL.Path, r.Body)
}

var (
	config Config
)

func main() {
	// Parse options
	_, err := flags.Parse(&config)
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			os.Exit(1) // help printed
		} else {
			os.Exit(2) // error message written already
		}
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if config.Quiet {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	h := http.NewServeMux()

	h.HandleFunc("/", handler)

	log.Debug().Msgf("Listen Address: %s", config.HTTP.Bind)
	log.Fatal().Err(http.ListenAndServe(config.HTTP.Bind, h)).Msg("Startup failed")
}
