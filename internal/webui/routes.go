package webui

import "net/http"

func (webUI *WebUI) SetWebUIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /debug/", webUI.debugIndexHandler)
}
