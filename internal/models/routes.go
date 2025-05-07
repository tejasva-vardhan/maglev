package models

type RouteType int

type Route struct {
	ID                string    `json:"id"`
	AgencyID          string    `json:"agencyId"`
	ShortName         string    `json:"shortName"`
	LongName          string    `json:"longName"`
	Description       string    `json:"description"`
	Type              RouteType `json:"type"`
	URL               string    `json:"url"`
	Color             string    `json:"color"`
	TextColor         string    `json:"textColor"`
	NullSafeShortName string    `json:"nullSafeShortName"`
}

func NewRoute(id, agencyID, shortName, longName, description string, routeType RouteType, url, color, textColor, nullSafeShortName string) Route {
	return Route{
		ID:                id,
		AgencyID:          agencyID,
		ShortName:         shortName,
		LongName:          longName,
		Description:       description,
		Type:              routeType,
		URL:               url,
		Color:             color,
		TextColor:         textColor,
		NullSafeShortName: nullSafeShortName,
	}
}

type RouteResponse struct {
	Code        int       `json:"code"`
	CurrentTime int64     `json:"currentTime"`
	Data        RouteData `json:"data"`
	Text        string    `json:"text"`
	Version     int       `json:"version"`
}

type RouteData struct {
	LimitExceeded bool            `json:"limitExceeded"`
	List          []Route         `json:"list"`
	References    ReferencesModel `json:"references"`
}
