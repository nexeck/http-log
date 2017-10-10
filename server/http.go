package server

import (
	"github.com/nexeck/multicast"
	"github.com/rs/zerolog"
	"io/ioutil"
	"net/http"
	"time"
)

type Log struct {
	Time   string
	Proto  string
	Uri    string
	Method string
	Header http.Header
	Body   string
}

func (l *Log) MarshalZerologObject(e *zerolog.Event) {
	e.
		Str("Proto", l.Proto).
		Str("Uri", l.Uri).
		Str("Method", l.Method).
		Str("Body", l.Body).
		Msg("Request")
}

type HTTPServer struct {
	mux            *http.ServeMux
	multicastGroup *multicast.Group
}

func NewHTTPServer(multicastGroup *multicast.Group) *HTTPServer {
	s := &HTTPServer{
		mux:            http.NewServeMux(),
		multicastGroup: multicastGroup,
	}

	s.mux.HandleFunc("/", s.catchAll)

	return s
}

func (s *HTTPServer) catchAll(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	logMessage := Log{
		Time:   time.Now().Format(time.RFC3339Nano),
		Proto:  r.Proto,
		Uri:    r.RequestURI,
		Method: r.Method,
		Header: r.Header,
		Body:   string(body),
	}

	s.multicastGroup.Send(logMessage)
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HTTPServer", "HTTP Debug Server")

	s.mux.ServeHTTP(w, r)
}
