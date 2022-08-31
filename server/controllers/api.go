package controllers

import (
	"encoding/json"
	"net/http"
)

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).Header().Set("Access-Control-Allow-Origin", API_ORIGIN)
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, access-control-allow-origin, access-control-allow-headers")
}

func GetChannels(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	json.NewEncoder(w).Encode(Channels)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	json.NewEncoder(w).Encode(Users)
}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	json.NewEncoder(w).Encode(Files)
}
