package restapi

type StopDelayInfo struct {
	ArrivalDelay   int64
	DepartureDelay int64
}

func (api *RestAPI) GetScheduleDeviation(tripID string) (int, bool) {
	tripUpdates := api.GtfsManager.GetTripUpdatesForTrip(tripID)
	if len(tripUpdates) == 0 {
		return 0, false
	}

	tu := tripUpdates[0]

	if tu.Delay != nil {
		return int(tu.Delay.Seconds()), true
	}

	for _, stu := range tu.StopTimeUpdates {
		if stu.Arrival != nil && stu.Arrival.Delay != nil {
			return int(stu.Arrival.Delay.Seconds()), true
		}
		if stu.Departure != nil && stu.Departure.Delay != nil {
			return int(stu.Departure.Delay.Seconds()), true
		}
	}

	return 0, false
}

func (api *RestAPI) GetStopDelaysFromTripUpdates(tripID string) map[string]StopDelayInfo {
	delays := make(map[string]StopDelayInfo)

	tripUpdates := api.GtfsManager.GetTripUpdatesForTrip(tripID)
	if len(tripUpdates) == 0 {
		return delays
	}

	for _, stu := range tripUpdates[0].StopTimeUpdates {
		if stu.StopID == nil {
			continue
		}

		info := StopDelayInfo{}
		if stu.Arrival != nil && stu.Arrival.Delay != nil {
			info.ArrivalDelay = int64(stu.Arrival.Delay.Seconds())
		}
		if stu.Departure != nil && stu.Departure.Delay != nil {
			info.DepartureDelay = int64(stu.Departure.Delay.Seconds())
		}

		delays[*stu.StopID] = info
	}

	return delays
}
