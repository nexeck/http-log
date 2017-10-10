package main

import (
	"github.com/jessevdk/go-flags"
	gracefulhttp "github.com/nexeck/graceful/http"
	"github.com/nexeck/http-log/server"
	"github.com/nexeck/multicast"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
)

type Config struct {
	Debug bool `long:"debug" description:"enable exception display and pprof endpoints (warn: dangerous)"`
	Quiet bool `long:"quiet" description:"disable verbose output"`
	HTTP  struct {
		Bind string `long:"http-bind" description:"address and port to bind to" default:":8080"`
	}
	WS struct {
		Enabled bool   `long:"ws-enable" description:"enable websocket"`
		Bind    string `long:"ws-bind" description:"address and port to bind to" default:":8081"`
	}
	Pprof struct {
		Enable bool   `long:"pprof-enable" description:"enable pprof"`
		Bind   string `long:"pprof-bind" description:"address and port to bind to" default:":8082"`
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
	multicastGroup := multicast.NewGroup()
	go multicastGroup.Broadcast()
	defer multicastGroup.Quit()

	startHTTPServer(&wg, multicastGroup)

	if config.WS.Enabled {
		startWSServer(&wg, multicastGroup.Join())
	}

	startLogServer(&wg, multicastGroup.Join())

	if config.Pprof.Enable {
		startPprofServer(&wg)
	}

	wg.Wait()
}

func startPprofServer(wg *sync.WaitGroup) {
	httpServer := &http.Server{
		Addr: config.Pprof.Bind,
	}

	r := http.NewServeMux()

	// Register pprof handlers
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	httpServer.Handler = r

	gracefulHTTP := gracefulhttp.New(httpServer, 5)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		gracefulHTTP.Run()
	}(wg)
}

func startHTTPServer(wg *sync.WaitGroup, multicastGroup *multicast.Group) {
	httpServer := &http.Server{
		Addr:    config.HTTP.Bind,
		Handler: server.NewHTTPServer(multicastGroup),
	}

	gracefulHTTP := gracefulhttp.New(httpServer, 5)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		gracefulHTTP.Run()
	}(wg)
}

func startWSServer(wg *sync.WaitGroup, multicastMember *multicast.Member) {
	wsServer := &http.Server{
		Addr:    config.WS.Bind,
		Handler: server.NewWSServer(multicastMember),
	}

	gracefulHTTP := gracefulhttp.New(wsServer, 5)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		gracefulHTTP.Run()
	}(wg)
}

func startLogServer(wg *sync.WaitGroup, multicastMember *multicast.Member) {
	logServer := server.NewLogServer(multicastMember)

	wg.Add(1)
	go func(wg *sync.WaitGroup, logServer *server.LogServer) {
		defer wg.Done()

		logServer.Run()
	}(wg, logServer)
}
