package ui

import (
	"net/http"
	"strconv"
)

func parseIntParam(r *http.Request, key string, def int) int {
	valueRaw := r.URL.Query().Get(key)
	valueInt, err := strconv.Atoi(valueRaw)
	if err != nil {
		return def
	}

	return valueInt
}
