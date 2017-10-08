package server

import (
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
)

type WSServer struct {
	mux      *http.ServeMux
	upgrader websocket.Upgrader
	chLogs   <-chan Log
}

func NewWSServer(chLogs <-chan Log) *WSServer {
	s := &WSServer{
		mux: http.NewServeMux(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		chLogs: chLogs,
	}

	s.mux.HandleFunc("/http-debug", s.httpDebug)

	return s
}

func (s *WSServer) httpDebug(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Upgrade failed")
		return
	}
	defer c.Close()

	for {
		select {
		// send message to the client
		case logMessage := <-s.chLogs:
			err = c.WriteJSON(logMessage)
			if err != nil {
				log.Error().Err(err).Msg("Write failed")
				break
			}
		}
	}
}

func (s *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HTTPServer", "HTTP Debug Websocket Server")

	s.mux.ServeHTTP(w, r)
}
