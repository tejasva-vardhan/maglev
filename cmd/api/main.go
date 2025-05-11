package main

import (
	"flag"
	"fmt"
	"log/slog"
	"maglev.onebusaway.org/internal/app"
	"maglev.onebusaway.org/internal/gtfs"
	"net/http"
	"os"
	"strings"
	"time"
)

type restAPI struct {
	app *app.Application
}

func main() {
	var cfg app.Config
	var gtfsCfg gtfs.Config
	var apiKeysFlag string

	flag.IntVar(&cfg.Port, "port", 4000, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&apiKeysFlag, "api-keys", "test", "Comma Separated API Keys (test, etc)")
	flag.StringVar(&gtfsCfg.GtfsURL, "gtfs-url", "https://www.soundtransit.org/GTFS-rail/40_gtfs.zip", "URL for a static GTFS zip file")
	flag.StringVar(&gtfsCfg.TripUpdatesURL, "trip-updates-url", "https://api.pugetsound.onebusaway.org/api/gtfs_realtime/trip-updates-for-agency/40.pb?key=org.onebusaway.iphone", "URL for a GTFS-RT trip updates feed")
	flag.StringVar(&gtfsCfg.VehiclePositionsURL, "vehicle-positions-url", "https://api.pugetsound.onebusaway.org/api/gtfs_realtime/vehicle-positions-for-agency/40.pb?key=org.onebusaway.iphone", "URL for a GTFS-RT vehicle positions feed")
	flag.StringVar(&gtfsCfg.RealTimeAuthHeaderKey, "realtime-auth-header-name", "", "Optional header name for GTFS-RT auth")
	flag.StringVar(&gtfsCfg.RealTimeAuthHeaderValue, "realtime-auth-header-value", "", "Optional header value for GTFS-RT auth")
	flag.Parse()

	if apiKeysFlag != "" {
		cfg.ApiKeys = strings.Split(apiKeysFlag, ",")
		for i := range cfg.ApiKeys {
			cfg.ApiKeys[i] = strings.TrimSpace(cfg.ApiKeys[i])
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gtfsManager, err := gtfs.InitGTFSManager(gtfsCfg)
	if err != nil {
		logger.Error("failed to initialize GTFS manager", "error", err)
	}

	gtfsManager.PrintStatistics()

	coreApp := &app.Application{
		Config:      cfg,
		GtfsConfig:  gtfsCfg,
		Logger:      logger,
		GtfsManager: gtfsManager,
	}

	api := restAPI{app: coreApp}

	mux := http.NewServeMux()
	api.setRoutes(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.Env)
	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
