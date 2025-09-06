package services

import (
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
)

// TimezoneLocationService maps timezones to approximate locations
type TimezoneLocationService struct{}

// NewTimezoneLocationService creates a new timezone location service
func NewTimezoneLocationService() *TimezoneLocationService {
	return &TimezoneLocationService{}
}

// timezoneLocationMap maps timezone IDs to representative locations
var timezoneLocationMap = map[string]*models.Location{
	// United States
	"America/New_York": {
		Latitude:  40.7128,
		Longitude: -74.0060,
		City:      "New York",
		Region:    "New York",
		Country:   "United States",
	},
	"America/Chicago": {
		Latitude:  41.8781,
		Longitude: -87.6298,
		City:      "Chicago",
		Region:    "Illinois",
		Country:   "United States",
	},
	"America/Denver": {
		Latitude:  39.7392,
		Longitude: -104.9903,
		City:      "Denver",
		Region:    "Colorado",
		Country:   "United States",
	},
	"America/Los_Angeles": {
		Latitude:  34.0522,
		Longitude: -118.2437,
		City:      "Los Angeles",
		Region:    "California",
		Country:   "United States",
	},
	"America/Phoenix": {
		Latitude:  33.4484,
		Longitude: -112.0740,
		City:      "Phoenix",
		Region:    "Arizona",
		Country:   "United States",
	},
	"America/Anchorage": {
		Latitude:  61.2181,
		Longitude: -149.9003,
		City:      "Anchorage",
		Region:    "Alaska",
		Country:   "United States",
	},
	"Pacific/Honolulu": {
		Latitude:  21.3099,
		Longitude: -157.8581,
		City:      "Honolulu",
		Region:    "Hawaii",
		Country:   "United States",
	},

	// Canada
	"America/Toronto": {
		Latitude:  43.6532,
		Longitude: -79.3832,
		City:      "Toronto",
		Region:    "Ontario",
		Country:   "Canada",
	},
	"America/Vancouver": {
		Latitude:  49.2827,
		Longitude: -123.1207,
		City:      "Vancouver",
		Region:    "British Columbia",
		Country:   "Canada",
	},
	"America/Halifax": {
		Latitude:  44.6488,
		Longitude: -63.5752,
		City:      "Halifax",
		Region:    "Nova Scotia",
		Country:   "Canada",
	},

	// Europe
	"Europe/London": {
		Latitude:  51.5074,
		Longitude: -0.1278,
		City:      "London",
		Region:    "England",
		Country:   "United Kingdom",
	},
	"Europe/Paris": {
		Latitude:  48.8566,
		Longitude: 2.3522,
		City:      "Paris",
		Region:    "Île-de-France",
		Country:   "France",
	},
	"Europe/Berlin": {
		Latitude:  52.5200,
		Longitude: 13.4050,
		City:      "Berlin",
		Region:    "Berlin",
		Country:   "Germany",
	},
	"Europe/Madrid": {
		Latitude:  40.4168,
		Longitude: -3.7038,
		City:      "Madrid",
		Region:    "Madrid",
		Country:   "Spain",
	},
	"Europe/Rome": {
		Latitude:  41.9028,
		Longitude: 12.4964,
		City:      "Rome",
		Region:    "Lazio",
		Country:   "Italy",
	},
	"Europe/Amsterdam": {
		Latitude:  52.3676,
		Longitude: 4.9041,
		City:      "Amsterdam",
		Region:    "North Holland",
		Country:   "Netherlands",
	},
	"Europe/Stockholm": {
		Latitude:  59.3293,
		Longitude: 18.0686,
		City:      "Stockholm",
		Region:    "Stockholm",
		Country:   "Sweden",
	},

	// Australia
	"Australia/Sydney": {
		Latitude:  -33.8688,
		Longitude: 151.2093,
		City:      "Sydney",
		Region:    "New South Wales",
		Country:   "Australia",
	},
	"Australia/Melbourne": {
		Latitude:  -37.8136,
		Longitude: 144.9631,
		City:      "Melbourne",
		Region:    "Victoria",
		Country:   "Australia",
	},
	"Australia/Brisbane": {
		Latitude:  -27.4698,
		Longitude: 153.0251,
		City:      "Brisbane",
		Region:    "Queensland",
		Country:   "Australia",
	},
	"Australia/Perth": {
		Latitude:  -31.9505,
		Longitude: 115.8605,
		City:      "Perth",
		Region:    "Western Australia",
		Country:   "Australia",
	},

	// Asia
	"Asia/Tokyo": {
		Latitude:  35.6762,
		Longitude: 139.6503,
		City:      "Tokyo",
		Region:    "Tokyo",
		Country:   "Japan",
	},
	"Asia/Shanghai": {
		Latitude:  31.2304,
		Longitude: 121.4737,
		City:      "Shanghai",
		Region:    "Shanghai",
		Country:   "China",
	},
	"Asia/Singapore": {
		Latitude:  1.3521,
		Longitude: 103.8198,
		City:      "Singapore",
		Region:    "Singapore",
		Country:   "Singapore",
	},
	"Asia/Dubai": {
		Latitude:  25.2048,
		Longitude: 55.2708,
		City:      "Dubai",
		Region:    "Dubai",
		Country:   "United Arab Emirates",
	},

	// New Zealand
	"Pacific/Auckland": {
		Latitude:  -36.8485,
		Longitude: 174.7633,
		City:      "Auckland",
		Region:    "Auckland",
		Country:   "New Zealand",
	},

	// South America
	"America/Sao_Paulo": {
		Latitude:  -23.5505,
		Longitude: -46.6333,
		City:      "São Paulo",
		Region:    "São Paulo",
		Country:   "Brazil",
	},
	"America/Buenos_Aires": {
		Latitude:  -34.6037,
		Longitude: -58.3816,
		City:      "Buenos Aires",
		Region:    "Buenos Aires",
		Country:   "Argentina",
	},

	// Mexico
	"America/Mexico_City": {
		Latitude:  19.4326,
		Longitude: -99.1332,
		City:      "Mexico City",
		Region:    "Mexico City",
		Country:   "Mexico",
	},

	// Default/UTC fallback
	"UTC": {
		Latitude:  51.5074, // Default to London for UTC
		Longitude: -0.1278,
		City:      "London",
		Region:    "England",
		Country:   "United Kingdom",
	},
}

// GetLocationFromTimezone returns an approximate location based on timezone
func (s *TimezoneLocationService) GetLocationFromTimezone(timezone string) *models.Location {
	// Clean up timezone string
	timezone = strings.TrimSpace(timezone)

	// Try exact match first
	if location, ok := timezoneLocationMap[timezone]; ok {
		return location
	}

	// Try to find a similar timezone (e.g., "US/Eastern" -> "America/New_York")
	if strings.Contains(timezone, "Eastern") || timezone == "EST" || timezone == "EDT" {
		return timezoneLocationMap["America/New_York"]
	}
	if strings.Contains(timezone, "Central") || timezone == "CST" || timezone == "CDT" {
		return timezoneLocationMap["America/Chicago"]
	}
	if strings.Contains(timezone, "Mountain") || timezone == "MST" || timezone == "MDT" {
		return timezoneLocationMap["America/Denver"]
	}
	if strings.Contains(timezone, "Pacific") || timezone == "PST" || timezone == "PDT" {
		return timezoneLocationMap["America/Los_Angeles"]
	}

	// Check for GMT offsets and map to approximate regions
	if strings.HasPrefix(timezone, "GMT") || strings.HasPrefix(timezone, "UTC") {
		// Parse offset if present
		if strings.Contains(timezone, "-5") || strings.Contains(timezone, "-05") {
			return timezoneLocationMap["America/New_York"]
		}
		if strings.Contains(timezone, "-6") || strings.Contains(timezone, "-06") {
			return timezoneLocationMap["America/Chicago"]
		}
		if strings.Contains(timezone, "-7") || strings.Contains(timezone, "-07") {
			return timezoneLocationMap["America/Denver"]
		}
		if strings.Contains(timezone, "-8") || strings.Contains(timezone, "-08") {
			return timezoneLocationMap["America/Los_Angeles"]
		}
		if strings.Contains(timezone, "+0") || strings.Contains(timezone, "-0") {
			return timezoneLocationMap["Europe/London"]
		}
		if strings.Contains(timezone, "+10") {
			return timezoneLocationMap["Australia/Sydney"]
		}
		if strings.Contains(timezone, "+1") || strings.Contains(timezone, "+01") {
			return timezoneLocationMap["Europe/Paris"]
		}
		if strings.Contains(timezone, "+9") || strings.Contains(timezone, "+09") {
			return timezoneLocationMap["Asia/Tokyo"]
		}
	}

	// Default fallback - use London, England
	return &models.Location{
		Latitude:  51.5074,
		Longitude: -0.1278,
		City:      "London",
		Region:    "England",
		Country:   "United Kingdom",
	}
}
