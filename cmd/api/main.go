package main

import (
	"flag"
	"fmt"
	"log/slog"
	"maglev.onebusaway.org/internal/gtfs"
	"net/http"
	"os"
	"strings"
	"time"
)

// Define a config struct to hold all the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environment for the
// application (development, staging, production, etc.). We will read in these
// configuration settings from command-line flags when the application starts.
type config struct {
	port    int
	env     string
	apiKeys []string
	gtfsURL string
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the config struct and a
// logger, but it will grow to include a lot more as our build progresses.
type application struct {
	config      config
	logger      *slog.Logger
	gtfsManager *gtfs.Manager
}

func main() {
	var cfg config
	var apiKeysFlag string

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&apiKeysFlag, "api-keys", "test", "Comma Separated API Keys (test, etc)")
	flag.StringVar(&cfg.gtfsURL, "gtfs-url", "https://www.soundtransit.org/GTFS-rail/40_gtfs.zip", "URL for a static GTFS zip file")
	flag.Parse()

	if apiKeysFlag != "" {
		cfg.apiKeys = strings.Split(apiKeysFlag, ",")
		for i := range cfg.apiKeys {
			cfg.apiKeys[i] = strings.TrimSpace(cfg.apiKeys[i])
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gtfsManager, err := gtfs.InitGTFSManager(cfg.gtfsURL)
	if err != nil {
		logger.Error("failed to initialize GTFS manager", "error", err)
	}

	gtfsManager.PrintStatistics()

	app := &application{
		config:      cfg,
		logger:      logger,
		gtfsManager: gtfsManager,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)
	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
