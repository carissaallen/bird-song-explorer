package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/callen/bird-song-explorer/internal/models"
)

type LocationService struct{}

func NewLocationService() *LocationService {
	return &LocationService{}
}

func (s *LocationService) GetLocationFromIP(ip string) (*models.Location, error) {
	if ip == "" || ip == "::1" || ip == "127.0.0.1" {
		return s.getDefaultLocation(), nil
	}

	url := fmt.Sprintf("https://ipapi.co/%s/json/", ip)
	resp, err := http.Get(url)
	if err != nil {
		return s.getDefaultLocation(), nil
	}
	defer resp.Body.Close()

	var result struct {
		City      string  `json:"city"`
		Region    string  `json:"region"`
		Country   string  `json:"country_name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Error     bool    `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getDefaultLocation(), nil
	}

	if result.Error {
		return s.getDefaultLocation(), nil
	}

	return &models.Location{
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
		City:      result.City,
		Region:    result.Region,
		Country:   result.Country,
		IPAddress: ip,
	}, nil
}

func (s *LocationService) getDefaultLocation() *models.Location {
	return &models.Location{
		Latitude:  51.5074,
		Longitude: -0.1278,
		City:      "London",
		Region:    "England",
		Country:   "United Kingdom",
	}
}
