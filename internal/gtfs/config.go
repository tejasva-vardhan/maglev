package gtfs

import (
	"maglev.onebusaway.org/internal/appconf"
)

type Config struct {
	GtfsURL                 string
	TripUpdatesURL          string
	VehiclePositionsURL     string
	RealTimeAuthHeaderKey   string
	RealTimeAuthHeaderValue string
	GTFSDataPath            string
	Env                     appconf.Environment
}

func (config Config) realTimeDataEnabled() bool {
	return config.TripUpdatesURL != "" && config.VehiclePositionsURL != ""
}
