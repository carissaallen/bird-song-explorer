package services

import (
	"math"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/pkg/ebird"
)

// BirdRegionalMatcher checks if a bird is found in a specific region
type BirdRegionalMatcher struct {
	ebirdClient *ebird.Client
}

// NewBirdRegionalMatcher creates a new regional matcher
func NewBirdRegionalMatcher(ebirdAPIKey string) *BirdRegionalMatcher {
	return &BirdRegionalMatcher{
		ebirdClient: ebird.NewClient(ebirdAPIKey),
	}
}

// IsBirdInRegion checks if a bird has been seen near the given location
func (brm *BirdRegionalMatcher) IsBirdInRegion(bird *models.Bird, latitude, longitude float64) (bool, *RegionalInfo) {
	// If no eBird client, can't check
	if brm.ebirdClient == nil {
		return false, nil
	}

	// Get recent observations within 50km
	observations, err := brm.ebirdClient.GetRecentObservations(latitude, longitude, 30)
	if err != nil {
		// If API fails, return false but don't error
		return false, nil
	}

	// Check if this specific bird has been seen
	sightingCount := 0
	nearestDistance := 999.0

	for _, obs := range observations {
		// Check both common and scientific names
		if strings.EqualFold(obs.CommonName, bird.CommonName) ||
			strings.EqualFold(obs.ScientificName, bird.ScientificName) {
			sightingCount++

			// Calculate distance
			dist := calculateDistance(latitude, longitude, obs.Latitude, obs.Longitude)
			if dist < nearestDistance {
				nearestDistance = dist
			}

			// Track most recent sighting (would need date parsing)
			// For now, just count them
		}
	}

	if sightingCount > 0 {
		return true, &RegionalInfo{
			SightingCount:   sightingCount,
			NearestDistance: nearestDistance,
			IsCommonHere:    sightingCount > 5, // More than 5 sightings = common
		}
	}

	return false, nil
}

// GetBirdRange returns the general range/habitat of a bird
func (brm *BirdRegionalMatcher) GetBirdRange(bird *models.Bird) BirdRange {
	// This would ideally use a bird range database
	// For now, use simple heuristics based on bird names

	commonName := strings.ToLower(bird.CommonName)

	// North American birds
	if strings.Contains(commonName, "american") ||
		strings.Contains(commonName, "northern") ||
		strings.Contains(commonName, "eastern") ||
		strings.Contains(commonName, "western") {
		return BirdRange{
			Regions: []string{"North America"},
			Habitat: getHabitatFromName(commonName),
		}
	}

	// European birds
	if strings.Contains(commonName, "european") ||
		strings.Contains(commonName, "eurasian") {
		return BirdRange{
			Regions: []string{"Europe", "Asia"},
			Habitat: getHabitatFromName(commonName),
		}
	}

	// Tropical birds
	if strings.Contains(commonName, "tropical") ||
		strings.Contains(commonName, "parrot") ||
		strings.Contains(commonName, "toucan") {
		return BirdRange{
			Regions: []string{"Tropical"},
			Habitat: "Tropical forests",
		}
	}

	// Water birds (found globally)
	if strings.Contains(commonName, "duck") ||
		strings.Contains(commonName, "goose") ||
		strings.Contains(commonName, "swan") ||
		strings.Contains(commonName, "heron") {
		return BirdRange{
			Regions: []string{"Global"},
			Habitat: "Wetlands, lakes, and rivers",
		}
	}

	// Default to global/unknown
	return BirdRange{
		Regions: []string{"Various regions worldwide"},
		Habitat: "Various habitats",
	}
}

// RegionalInfo contains information about regional sightings
type RegionalInfo struct {
	SightingCount   int
	NearestDistance float64 // in miles
	IsCommonHere    bool
}

// BirdRange describes where a bird is typically found
type BirdRange struct {
	Regions []string
	Habitat string
}

// Helper function to calculate distance between two coordinates
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 3959.0 // miles

	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Helper function to guess habitat from bird name
func getHabitatFromName(name string) string {
	if strings.Contains(name, "forest") || strings.Contains(name, "wood") {
		return "Forests and woodlands"
	}
	if strings.Contains(name, "meadow") || strings.Contains(name, "field") {
		return "Grasslands and fields"
	}
	if strings.Contains(name, "mountain") || strings.Contains(name, "alpine") {
		return "Mountain regions"
	}
	if strings.Contains(name, "desert") {
		return "Desert regions"
	}
	if strings.Contains(name, "shore") || strings.Contains(name, "coastal") {
		return "Coastal areas"
	}
	return "Various habitats"
}
