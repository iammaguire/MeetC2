package main

import (
	"net/http"
	"encoding/json"
)

func (server WebInterface) beaconsHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(beacons)
}

func (server WebInterface) listenersHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(listeners)
}