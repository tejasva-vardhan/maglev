package webui

import "net/http"

func SetWebUIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /debug/", debugIndexHandler)
}
