package main

import (
	"context"
	"log/slog"

	"github.com/grodier/rss-app/server"
)

type Application struct {
	config config
	logger *slog.Logger
}

func NewApplication(logger *slog.Logger) *Application {
	return &Application{
		config: defaultConfig(),
		logger: logger,
	}
}

type config struct {
	env    string
	server serverConfig
}

type serverConfig struct {
	port int
}

func defaultConfig() config {
	return config{
		env: "development",
		server: serverConfig{
			port: 8080,
		},
	}
}

func (app *Application) Run(ctx context.Context, args []string) error {
	srv := server.NewServer(app.logger)
	srv.Port = app.config.server.port
	srv.Env = app.config.env

	if err := srv.Serve(); err != nil {
		return err
	}

	return nil
}
