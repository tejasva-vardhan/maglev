package utils

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// ExtractIDFromParams retrieves a parameter value from the request context and removes file extensions like ".json".
func ExtractIDFromParams(r *http.Request, paramName string) string {
	params := httprouter.ParamsFromContext(r.Context())
	rawID := params.ByName(paramName)
	return strings.Split(rawID, ".json")[0]
}
