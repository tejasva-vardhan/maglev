package utils

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"testing"
)

func TestExtractIDFromParams(t *testing.T) {
	req, _ := http.NewRequest("GET", "/agency/40.json", nil)

	params := httprouter.Params{
		{Key: "id.json", Value: "40.json"},
	}
	ctx := context.WithValue(context.Background(), httprouter.ParamsKey, params)
	req = req.WithContext(ctx)

	id := ExtractIDFromParams(req, "id.json")
	expected := "40"

	if id != expected {
		t.Errorf("expected %s, got %s", expected, id)
	}
}
