package restapi

import (
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
	"net/http"
)

func (api *RestAPI) vehiclesForAgencyHandler(w http.ResponseWriter, r *http.Request) {
	id := utils.ExtractIDFromParams(r)

	agency := api.GtfsManager.FindAgency(id)
	if agency == nil {
		// return an empty list response.
		api.sendResponse(w, r, models.NewListResponse([]interface{}{}, models.ReferencesModel{}))
		return
	}

	vehiclesForAgency := api.GtfsManager.VehiclesForAgencyID(id)
	vehiclesList := make([]models.VehicleStatus, 0, len(vehiclesForAgency))
	for _, vehicle := range vehiclesForAgency {
		vehiclesList = append(vehiclesList, models.VehicleStatus{
			VehicleID: vehicle.ID.ID,
			TripID:    vehicle.Trip.ID.ID,
		})
	}

	references := models.ReferencesModel{
		Agencies:   []models.AgencyReference{},
		Routes:     []interface{}{},
		Situations: []interface{}{},
		StopTimes:  []interface{}{},
		Stops:      []interface{}{},
		Trips:      []interface{}{},
	}

	response := models.NewListResponse(vehiclesList, references)
	api.sendResponse(w, r, response)
}
