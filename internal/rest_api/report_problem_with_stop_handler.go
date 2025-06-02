package restapi

import (
	"log"
	"net/http"

	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

func (api *RestAPI) reportProblemWithStopHandler(w http.ResponseWriter, r *http.Request) {
	stopID := utils.ExtractIDFromParams(r)

	// TODO: Add required validation
	if stopID == "" {
		api.sendNull(w, r)
		return
	}

	query := r.URL.Query()

	code := query.Get("code")
	userComment := query.Get("userComment")
	userLat := query.Get("userLat")
	userLon := query.Get("userLon")
	userLocationAccuracy := query.Get("userLocationAccuracy")

	// TODO: Add storage logic for the problem report, I leave it as a log statement for now
	log.Printf("Problem report received for stop %s: code=%s, userComment=%s, "+
		"userLat=%s, userLon=%s, userLocationAccuracy=%s",
		stopID, code, userComment, userLat, userLon, userLocationAccuracy)

	api.sendResponse(w, r, models.NewOKResponse(struct{}{}))
}
