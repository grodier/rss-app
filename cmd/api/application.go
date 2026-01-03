package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/grodier/rss-app/internal/pgsql"
	"github.com/grodier/rss-app/internal/server"
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
	db     struct {
		dsn string
	}
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
	app.config = app.ParseConfigs(args)

	db := pgsql.NewDB(app.config.db.dsn)
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	app.logger.Info("database connection pool established")

	srv := server.NewServer(app.logger)
	srv.Port = app.config.server.port
	srv.Env = app.config.env
	srv.Version = version

	if err := srv.Serve(); err != nil {
		return err
	}

	return nil
}

func (app *Application) ParseConfigs(args []string) config {
	config := defaultConfig()

	fs := flag.NewFlagSet("rss-go", flag.ContinueOnError)

	fs.StringVar(&config.env, "env", config.env, "Environment (development|production)")
	fs.IntVar(&config.server.port, "port", config.server.port, "Server port")

	fs.StringVar(&config.db.dsn, "db-dsn", os.Getenv("RSSAPP_DB_DSN"), "Database DSN")

	fs.Parse(args)

	if config.env != "development" && config.env != "production" {
		app.logger.Warn("invalid environment value, falling back to default", "provided", config.env, "default", "development")
		config.env = "development"
	}

	return config
}
