package webui

import (
	"embed"
	"github.com/davecgh/go-spew/spew"
	"html"
	"html/template"
	"net/http"
)

//go:embed debug_index.html
var templateFS embed.FS

type debugData struct {
	Title string
	Pre   string
}

func writeDebugData(w http.ResponseWriter, title string, data interface{}) {
	content := html.EscapeString(spew.Sdump(data))
	w.Header().Set("Content-Type", "text/html")
	tmpl, err := template.ParseFS(templateFS, "debug_index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataStruct := debugData{
		Title: title,
		Pre:   content,
	}

	err = tmpl.Execute(w, dataStruct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func debugIndexHandler(w http.ResponseWriter, r *http.Request) {
	out := map[string]string{
		"hello": "world",
	}
	writeDebugData(w, "Hello, World!", out)
}
