package server

import (
	"io/ioutil"
	"net/http"
)

type Log struct {
	Proto  string
	Uri    string
	Method string
	Header http.Header
	Body   string
}

type HTTPServer struct {
	mux    *http.ServeMux
	chLogs chan<- Log
}

func NewHTTPServer(chLogs chan<- Log) *HTTPServer {
	s := &HTTPServer{
		mux:    http.NewServeMux(),
		chLogs: chLogs,
	}

	s.mux.HandleFunc("/", s.catchAll)

	return s
}

func (s *HTTPServer) catchAll(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	logMessage := Log{
		Proto:  r.Proto,
		Uri:    r.RequestURI,
		Method: r.Method,
		Header: r.Header,
		Body:   string(body),
	}

	select {
	case s.chLogs <- logMessage:
	default:
	}
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HTTPServer", "HTTP Debug Server")

	s.mux.ServeHTTP(w, r)
}
