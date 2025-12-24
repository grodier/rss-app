package main

import (
	"context"
	"fmt"
	"log/slog"
)

type Application struct {
	config config
	logger *slog.Logger
}

func NewApplication(logger *slog.Logger) *Application {
	return &Application{
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
	fmt.Println("Application Running")
	return nil
}
