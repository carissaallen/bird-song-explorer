package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/callen/bird-song-explorer/internal/models"
)

type LocationService struct{}

func NewLocationService() *LocationService {
	return &LocationService{}
}

func (s *LocationService) GetLocationFromIP(ip string) (*models.Location, error) {
	if ip == "" || ip == "::1" || ip == "127.0.0.1" {
		return nil, fmt.Errorf("invalid IP address for geolocation: %s", ip)
	}

	// Using ip-api.com instead of ipapi.co (better rate limits for free tier)
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[LOCATION] Failed to get IP location for %s: %v", ip, err)
		return nil, fmt.Errorf("failed to get IP location: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status    string  `json:"status"`
		City      string  `json:"city"`
		Region    string  `json:"regionName"`
		Country   string  `json:"country"`
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
		Message   string  `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[LOCATION] Failed to decode IP location response: %v", err)
		return nil, fmt.Errorf("failed to decode location response: %w", err)
	}

	if result.Status != "success" {
		log.Printf("[LOCATION] IP geolocation failed for %s: %s", ip, result.Message)
		return nil, fmt.Errorf("IP geolocation failed: %s", result.Message)
	}

	log.Printf("[LOCATION] Successfully resolved IP %s to %s, %s", ip, result.City, result.Country)
	
	return &models.Location{
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
		City:      result.City,
		Region:    result.Region,
		Country:   result.Country,
		IPAddress: ip,
	}, nil
}
