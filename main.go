package main

import (
	"context"
	"github.com/jessevdk/go-flags"
	"github.com/nexeck/http-log/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Debug bool `long:"debug" description:"enable exception display and pprof endpoints (warn: dangerous)"`
	Quiet bool `long:"quiet" description:"disable verbose output"`
	HTTP  struct {
		Bind string `long:"http-bind" description:"address and port to bind to" default:":8080"`
	}
	WS struct {
		Bind string `long:"ws-bind" description:"address and port to bind to" default:":8081"`
	}
}

var (
	config Config
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Parse options
	_, err := flags.Parse(&config)
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			os.Exit(1) // help printed
		} else {
			log.Fatal().Msgf("Error: %s", e.Message)
			os.Exit(2) // error message written already
		}
	}

	if config.Quiet {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	if config.HTTP.Bind == config.WS.Bind {
		log.Fatal().Msgf("HTTP and WS Bind cannot be the same")
	}

	var wg sync.WaitGroup
	chLogs := make(chan server.Log)

	startHTTPServer(&wg, chLogs)
	startWSServer(&wg, chLogs)
	startLogServer(&wg, chLogs)

	wg.Wait()
	close(chLogs)
}

func startHTTPServer(wg *sync.WaitGroup, chLogs chan server.Log) {
	httpServer := &http.Server{
		Addr:    config.HTTP.Bind,
		Handler: server.NewHTTPServer(chLogs),
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Info().Msgf("HTTP Listen Address: %s", config.HTTP.Bind)
		log.Error().Err(httpServer.ListenAndServe()).Msg("HTTPServer startup failed")

		wg.Done()
	}(wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		graceful(httpServer, "HTTPServer", 5*time.Second)

		wg.Done()
	}(wg)
}

func startWSServer(wg *sync.WaitGroup, chLogs chan server.Log) {
	wsServer := &http.Server{
		Addr:    config.WS.Bind,
		Handler: server.NewWSServer(chLogs),
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Info().Msgf("WS Listen Address: %s", config.WS.Bind)
		log.Error().Err(wsServer.ListenAndServe()).Msg("WSServer startup failed")

		wg.Done()
	}(wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		graceful(wsServer, "WSServer", 5*time.Second)

		wg.Done()
	}(wg)
}

func startLogServer(wg *sync.WaitGroup, chLogs chan server.Log) {
	logServer := server.NewLogServer(chLogs)

	logServer.Log()
}

func graceful(hs *http.Server, serverName string, timeout time.Duration) {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Info().Msgf("Shutdown %s with timeout: %s", serverName, timeout)

	if err := hs.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msgf("Failed stop %s", serverName)
	}
}
