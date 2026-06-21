package handler

import (
	"encoding/json"
	"net/http"
)

func parseJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
