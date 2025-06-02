package restapi

import (
	"log"
	"net/http"

	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

func (api *RestAPI) reportProblemWithTripHandler(w http.ResponseWriter, r *http.Request) {

	tripID := utils.ExtractIDFromParams(r)

	// TODO: Add required validation
	if tripID == "" {
		api.sendNull(w, r)
		return
	}

	query := r.URL.Query()

	serviceDate := query.Get("serviceDate")
	vehicleID := query.Get("vehicleId")
	stopID := query.Get("stopId")
	code := query.Get("code")
	userComment := query.Get("userComment")
	userOnVehicle := query.Get("userOnVehicle")
	userVehicleNumber := query.Get("userVehicleNumber")
	userLat := query.Get("userLat")
	userLon := query.Get("userLon")
	userLocationAccuracy := query.Get("userLocationAccuracy")

	// TODO: Add storage logic for the problem report, I leave it as a log statement for now
	log.Printf("Problem report received for trip %s: code=%s, serviceDate=%s, vehicleId=%s, stopId=%s, "+
		"userComment=%s, userOnVehicle=%s, userVehicleNumber=%s, userLat=%s, userLon=%s, userLocationAccuracy=%s",
		tripID, code, serviceDate, vehicleID, stopID, userComment, userOnVehicle,
		userVehicleNumber, userLat, userLon, userLocationAccuracy)

	api.sendResponse(w, r, models.NewOKResponse(struct{}{}))
}
