package models

import (
	"time"
)

type Bird struct {
	ID               int       `json:"id"`
	CommonName       string    `json:"common_name"`
	ScientificName   string    `json:"scientific_name"`
	Family           string    `json:"family"`
	Order            string    `json:"order"`
	Region           string    `json:"region"`
	AudioURL         string    `json:"audio_url"`
	AudioAttribution string    `json:"audio_attribution"`
	IconURL          string    `json:"icon_url"`
	Facts            []string  `json:"facts"`
	Description      string    `json:"description"`
	WikipediaURL     string    `json:"wikipedia_url"`
	Latitude         float64   `json:"latitude,omitempty"`
	Longitude        float64   `json:"longitude,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type BirdOfDay struct {
	ID        int       `json:"id"`
	Date      time.Time `json:"date"`
	BirdID    int       `json:"bird_id"`
	Bird      *Bird     `json:"bird,omitempty"`
	Location  Location  `json:"location"`
	CreatedAt time.Time `json:"created_at"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city"`
	Region    string  `json:"region"`
	Country   string  `json:"country"`
	IPAddress string  `json:"ip_address,omitempty"`
}
