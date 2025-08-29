package main

import (
	"log"
	"net/http"
	"os"

	"github.com/callen/bird-song-explorer/internal/api"
	"github.com/callen/bird-song-explorer/internal/config"
)

func main() {
	cfg := config.Load()

	router := api.SetupRouter(cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Bird Song Explorer server on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
