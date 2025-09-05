package restapi

import (
	"context"

	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

func (api *RestAPI) BuildRouteReferences(ctx context.Context, agencyID string, stops []models.Stop) ([]models.Route, error) {
	routeIDSet := make(map[string]bool)
	originalRouteIDs := make([]string, 0)

	for _, stop := range stops {
		for _, routeID := range stop.StaticRouteIDs {
			_, originalRouteID, err := utils.ExtractAgencyIDAndCodeID(routeID)
			if err != nil {
				continue
			}

			if !routeIDSet[originalRouteID] {
				routeIDSet[originalRouteID] = true
				originalRouteIDs = append(originalRouteIDs, originalRouteID)
			}
		}
	}

	if len(originalRouteIDs) == 0 {
		return []models.Route{}, nil
	}

	routes, err := api.GtfsManager.GtfsDB.Queries.GetRoutesByIDs(ctx, originalRouteIDs)
	if err != nil {
		return nil, err
	}

	modelRoutes := make([]models.Route, 0, len(routes))
	for _, route := range routes {
		routeModel := models.Route{
			ID:                utils.FormCombinedID(agencyID, route.ID),
			AgencyID:          agencyID,
			ShortName:         route.ShortName.String,
			LongName:          route.LongName.String,
			Description:       route.Desc.String,
			Type:              models.RouteType(route.Type),
			URL:               route.Url.String,
			Color:             route.Color.String,
			TextColor:         route.TextColor.String,
			NullSafeShortName: route.ShortName.String,
		}
		modelRoutes = append(modelRoutes, routeModel)
	}

	return modelRoutes, nil
}

func (api *RestAPI) BuildRouteReferencesAsInterface(ctx context.Context, agencyID string, stops []models.Stop) ([]interface{}, error) {
	routes, err := api.BuildRouteReferences(ctx, agencyID, stops)
	if err != nil {
		return nil, err
	}

	routeRefs := make([]interface{}, len(routes))
	for i, route := range routes {
		routeRefs[i] = route
	}

	return routeRefs, nil
}
