package server

import (
	"fmt"
	"log/slog"
	"net/http"
)

type Server struct {
	Port int
	Env  string

	server *http.Server
	logger *slog.Logger
}

func NewServer(logger *slog.Logger) *Server {
	s := &Server{
		logger: logger,
		server: &http.Server{
			ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		},
	}

	return s
}

func (s *Server) Serve() error {
	fmt.Println("Server Serving")
	return nil
}
