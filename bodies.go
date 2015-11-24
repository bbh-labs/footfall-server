package main

import (
	"encoding/json"
	"net/http"
)

type bodiesHandler struct{}

var bodiesMap = make(map[string]string)

func (_ bodiesHandler) GET(w http.ResponseWriter, r *http.Request) {
	if data, err := json.Marshal(bodiesMap); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write([]byte(data))
	}
}

func (_ bodiesHandler) POST(w http.ResponseWriter, r *http.Request) {
	location := r.FormValue("location")
	bodiesMap[location] = r.FormValue("bodies")
	w.WriteHeader(http.StatusOK)
}
