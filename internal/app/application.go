package app

import (
	"log/slog"
	"maglev.onebusaway.org/internal/gtfs"
)

// Application holds the dependencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the Config struct and a
// logger, but it will grow to include a lot more as our build progresses.
type Application struct {
	Config      Config
	GtfsConfig  gtfs.Config
	Logger      *slog.Logger
	GtfsManager *gtfs.Manager
}

// Config holds all the configuration settings for our Application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environment for the
// Application (development, staging, production, etc.). We will read in these
// configuration settings from command-line flags when the Application starts.
type Config struct {
	Port    int
	Env     string
	ApiKeys []string
}
