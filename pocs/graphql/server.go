package main

import (
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type Server struct {
	server http.Server
}

func NewServer(schema *graphql.Schema, port int, path string) (*Server, error) {
	s := &Server{}

	mux := http.NewServeMux()
	mux.Handle(path, handler.New(&handler.Config{
		Schema:   schema,
		Pretty:   true,
		GraphiQL: true,
	}))

	s.server.Addr = fmt.Sprintf("0.0.0.0:%d", port)
	s.server.Handler = mux
	return s, nil
}

func (s *Server) Start() error {
	err := s.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("error listening for connections: %w", err)
	}
	return nil
}
