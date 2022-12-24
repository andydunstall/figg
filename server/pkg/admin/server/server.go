package server

import (
	"net"
	"net/http"

	// Import so pprof registers HTTP handles to the server.
	_ "net/http/pprof"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Serve(lis net.Listener) error {
	return http.Serve(lis, nil)
}
