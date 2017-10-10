package server

import (
	"github.com/nexeck/multicast"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
)

type LogServer struct {
	multicastMember *multicast.Member
	quit            chan bool
}

func NewLogServer(multicastMember *multicast.Member) *LogServer {
	s := &LogServer{
		multicastMember: multicastMember,
		quit:            make(chan bool),
	}

	return s
}

func (s *LogServer) Run() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	go func(logServer *LogServer) {
		<-stopChan
		logServer.Quit()
	}(s)

	for {
		select {
		case logMessage := <-s.multicastMember.Read():
			log.Info().Interface("Run", logMessage).Msg("Received")

		case <-s.quit:
			log.Info().Msg("Shutdown Log Server")
			return

		}
	}
}

func (s *LogServer) Quit() {
	s.quit <- true
}
