package server

import (
	"github.com/rs/zerolog/log"
)

type LogServer struct {
	chLogs <-chan Log
}

func NewLogServer(chLogs <-chan Log) *LogServer {
	s := &LogServer{
		chLogs: chLogs,
	}

	return s
}

func (s *LogServer) Log() {
	for {
		select {
		// send message to the client
		case logMessage := <-s.chLogs:
			log.Debug().
				Str("Proto", logMessage.Proto).
				Str("Uri", logMessage.Uri).
				Str("Method", logMessage.Method).
				Str("Body", logMessage.Body).
				Msg("Request")
		}
	}
}
