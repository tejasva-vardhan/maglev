package gtfs

type Config struct {
	GtfsURL                 string
	TripUpdatesURL          string
	VehiclePositionsURL     string
	RealTimeAuthHeaderKey   string
	RealTimeAuthHeaderValue string
}

func (config Config) realTimeDataEnabled() bool {
	return config.TripUpdatesURL != "" && config.VehiclePositionsURL != ""
}
