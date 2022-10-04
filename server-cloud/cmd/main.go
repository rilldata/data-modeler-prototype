package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/server-cloud/database"
	_ "github.com/rilldata/rill/server-cloud/database/postgres"
	"github.com/rilldata/rill/server-cloud/server"
)

type Config struct {
	Env               string `default:"development"`
	DatabaseDriver    string `default:"postgres" split_words:"true"`
	DatabaseURL       string `split_words:"true"`
	Port              int    `default:"8080" split_words:"true"`
	SessionsSecret    string `split_words:"true"`
	Auth0ClientID     string `split_words:"true"`
	Auth0ClientSecret string `split_words:"true"`
	Auth0CallbackURL  string `split_words:"true"`
	Auth0Domain       string `split_words:"true"`
}

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Failed to load .env: %s", err.Error())
		os.Exit(1)
	}

	// Init config
	var conf Config
	err = envconfig.Process("rill_cloud", &conf)
	if err != nil {
		fmt.Printf("Failed to load config: %s", err.Error())
		os.Exit(1)
	}

	// Init logger
	var logger *zap.Logger
	if conf.Env == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Printf("Error creating logger: %s", err.Error())
		os.Exit(1)
	}

	// Init db
	db, err := database.Open(conf.DatabaseDriver, conf.DatabaseURL)
	if err != nil {
		logger.Fatal("error connecting to database", zap.Error(err))
	}

	// Auto-run migrations (TODO: don't do this in production)
	err = db.Migrate(context.Background())
	if err != nil {
		logger.Fatal("error migrating database", zap.Error(err))
	}

	srvConf := server.Config{
		Port:              conf.Port,
		Auth0ClientID:     conf.Auth0ClientID,
		Auth0ClientSecret: conf.Auth0ClientSecret,
		Auth0CallbackURL:  conf.Auth0CallbackURL,
		Auth0Domain:       conf.Auth0Domain,
		SessionsSecret:    conf.SessionsSecret,
	}

	// Init server
	server, err := server.New(logger, db, srvConf)
	if err != nil {
		logger.Fatal("error creating server", zap.Error(err))
	}

	// Run server
	logger.Info("serving http", zap.Int("port", conf.Port))

	ctx := graceful.WithCancelOnTerminate(context.Background())
	err = server.Serve(ctx, conf.Port)
	if err != nil {
		logger.Error("server crashed", zap.Error(err))
	}

	logger.Info("server shutdown gracefully")
}
