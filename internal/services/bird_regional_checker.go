package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
)

// BirdRegionalChecker checks if a bird has been spotted near a user's location
type BirdRegionalChecker struct {
	ebirdAPIKey string
}

func NewBirdRegionalChecker(ebirdAPIKey string) *BirdRegionalChecker {
	return &BirdRegionalChecker{
		ebirdAPIKey: ebirdAPIKey,
	}
}

// IsRegionalBird checks if a bird has been spotted within radius km of location in the last days
func (c *BirdRegionalChecker) IsRegionalBird(birdName string, location *models.Location, radiusKm int, days int) (bool, error) {
	if location == nil {
		return false, fmt.Errorf("location is nil")
	}

	// Get species code for the bird
	speciesCode, err := c.getSpeciesCode(birdName)
	if err != nil {
		log.Printf("[REGIONAL] Failed to get species code for %s: %v", birdName, err)
		return false, err
	}

	// Check recent observations near the location
	hasObservations, err := c.checkRecentObservations(speciesCode, location.Latitude, location.Longitude, radiusKm, days)
	if err != nil {
		log.Printf("[REGIONAL] Failed to check observations for %s: %v", birdName, err)
		return false, err
	}

	return hasObservations, nil
}

// getSpeciesCode gets the eBird species code for a common name
func (c *BirdRegionalChecker) getSpeciesCode(commonName string) (string, error) {
	// eBird taxonomy search endpoint
	apiURL := fmt.Sprintf("https://api.ebird.org/v2/ref/taxonomy/ebird?fmt=json&locale=en&q=%s",
		url.QueryEscape(commonName))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-eBirdApiToken", c.ebirdAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var results []struct {
		SpeciesCode string `json:"speciesCode"`
		ComName     string `json:"comName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no species found for name: %s", commonName)
	}

	// Return the first match
	return results[0].SpeciesCode, nil
}

// checkRecentObservations checks if there are recent observations of the species near the location
func (c *BirdRegionalChecker) checkRecentObservations(speciesCode string, lat, lng float64, radiusKm, days int) (bool, error) {
	// eBird recent observations endpoint
	apiURL := fmt.Sprintf("https://api.ebird.org/v2/data/obs/geo/recent/%s?lat=%.4f&lng=%.4f&dist=%d&back=%d",
		speciesCode, lat, lng, radiusKm, days)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-eBirdApiToken", c.ebirdAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// We just need to know if there are any observations
	var observations []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&observations); err != nil {
		// If we can't decode, assume no observations
		return false, nil
	}

	hasObservations := len(observations) > 0
	
	if hasObservations {
		log.Printf("[REGIONAL] Found %d observations of %s within %dkm in last %d days", 
			len(observations), speciesCode, radiusKm, days)
	} else {
		log.Printf("[REGIONAL] No observations of %s within %dkm in last %d days", 
			speciesCode, radiusKm, days)
	}

	return hasObservations, nil
}

// GetRegionalityMessage returns a message about whether the bird is regional to the user
func (c *BirdRegionalChecker) GetRegionalityMessage(birdName string, location *models.Location, isRegional bool) string {
	if location == nil {
		return ""
	}

	if isRegional {
		// Bird has been spotted in the user's area
		return fmt.Sprintf("The %s has been recently spotted near %s! Listen carefully - you might hear one nearby.", 
			birdName, location.City)
	} else if location.City != "" {
		// We know the user's location but the bird hasn't been spotted there
		return fmt.Sprintf("While the %s hasn't been spotted recently in %s, it's still a fascinating bird to learn about!", 
			birdName, location.City)
	}

	// No location information
	return ""
}