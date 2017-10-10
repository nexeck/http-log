package server

import (
	"github.com/gorilla/websocket"
	"github.com/nexeck/multicast"
	"github.com/rs/zerolog/log"
	"net/http"
)

type WSServer struct {
	mux             *http.ServeMux
	upgrader        websocket.Upgrader
	multicastMember *multicast.Member
}

func NewWSServer(multicastMember *multicast.Member) *WSServer {
	s := &WSServer{
		mux: http.NewServeMux(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		multicastMember: multicastMember,
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
		case logMessage := <-s.multicastMember.Read():
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
