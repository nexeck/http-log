package server

import (
	"github.com/nexeck/multicast"
	"github.com/rs/zerolog"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Log struct {
	Time          string
	Proto         string
	Host          string
	Uri           string
	Method        string
	Header        http.Header
	ContentLength int64
	Body          string
	Form          url.Values
	PostForm      url.Values
}

func (l *Log) MarshalZerologObject(e *zerolog.Event) {
	e.
		Str("Proto", l.Proto).
		Str("Host", l.Host).
		Str("Uri", l.Uri).
		Str("Method", l.Method).
		Int64("ContentLength", l.ContentLength).
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
		Time:          time.Now().Format(time.RFC3339Nano),
		Proto:         r.Proto,
		Host:          r.Host,
		Uri:           r.RequestURI,
		Method:        r.Method,
		Header:        r.Header,
		ContentLength: r.ContentLength,
		Body:          string(body),
		Form:          r.Form,
		PostForm:      r.PostForm,
	}

	s.multicastGroup.Send(logMessage)
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HTTPServer", "HTTP Debug Server")

	s.mux.ServeHTTP(w, r)
}
