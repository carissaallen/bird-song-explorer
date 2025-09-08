package services

import (
	"log"
	"time"

	timezoneLookup "github.com/evanoberholster/timezoneLookup/v2"
)

// TimezoneLookupService provides timezone lookup from coordinates
type TimezoneLookupService struct {
	cache *timezoneLookup.Timezonecache
}

// NewTimezoneLookupService creates a new timezone lookup service
func NewTimezoneLookupService() (*TimezoneLookupService, error) {
	cache := &timezoneLookup.Timezonecache{}
	
	// For now, we'll use a simple approach with pre-loaded timezone data
	// In production, you'd want to load this from a proper timezone database
	
	return &TimezoneLookupService{
		cache: cache,
	}, nil
}

// GetTimezone returns the timezone for given coordinates
func (s *TimezoneLookupService) GetTimezone(lat, lon float64) *time.Location {
	// For now, use a simplified approach based on coordinates
	// This is more reliable than the previous hardcoded ranges
	
	tzName := s.getTimezoneNameFromCoordinates(lat, lon)
	
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		log.Printf("[TIMEZONE] Failed to load timezone %s: %v, using UTC", tzName, err)
		loc, _ = time.LoadLocation("UTC")
		return loc
	}

	log.Printf("[TIMEZONE] Found timezone %s for coordinates %.4f, %.4f", tzName, lat, lon)
	return loc
}

// getTimezoneNameFromCoordinates provides a more comprehensive timezone mapping
func (s *TimezoneLookupService) getTimezoneNameFromCoordinates(lat, lon float64) string {
	// United States timezones (more comprehensive)
	if lat >= 24 && lat <= 50 && lon >= -125 && lon <= -66 {
		// Pacific Time
		if lon >= -125 && lon <= -114 {
			if lat >= 42 && lat <= 49 && lon >= -124 && lon <= -117 {
				// Oregon & Washington
				return "America/Los_Angeles"
			}
			if lat >= 32 && lat <= 42 {
				// California & Nevada
				return "America/Los_Angeles"
			}
			// General Pacific
			return "America/Los_Angeles"
		}
		// Mountain Time
		if lon >= -114 && lon <= -102 {
			if lat >= 31 && lat <= 49 {
				// Arizona doesn't observe DST
				if lat >= 31 && lat <= 37 && lon >= -114 && lon <= -109 {
					return "America/Phoenix"
				}
				return "America/Denver"
			}
		}
		// Central Time
		if lon >= -102 && lon <= -87 {
			return "America/Chicago"
		}
		// Eastern Time
		if lon >= -87 && lon <= -66 {
			return "America/New_York"
		}
	}

	// Canada
	if lat >= 41 && lat <= 84 {
		if lon >= -141 && lon <= -123 {
			return "America/Vancouver"
		}
		if lon >= -123 && lon <= -110 {
			return "America/Edmonton"
		}
		if lon >= -110 && lon <= -90 {
			return "America/Winnipeg"
		}
		if lon >= -90 && lon <= -74 {
			return "America/Toronto"
		}
		if lon >= -74 && lon <= -52 {
			return "America/Halifax"
		}
	}

	// Europe
	if lat >= 35 && lat <= 71 && lon >= -25 && lon <= 40 {
		if lon >= -10 && lon <= 2 {
			return "Europe/London"
		}
		if lon >= 2 && lon <= 15 {
			return "Europe/Paris"
		}
		if lon >= 15 && lon <= 25 {
			return "Europe/Berlin"
		}
		if lon >= 25 && lon <= 40 {
			return "Europe/Athens"
		}
	}

	// Asia
	if lat >= -10 && lat <= 55 && lon >= 60 && lon <= 150 {
		if lon >= 60 && lon <= 85 {
			return "Asia/Dubai"
		}
		if lon >= 85 && lon <= 97 {
			return "Asia/Kolkata"
		}
		if lon >= 97 && lon <= 110 {
			return "Asia/Bangkok"
		}
		if lon >= 110 && lon <= 130 {
			return "Asia/Shanghai"
		}
		if lon >= 130 && lon <= 145 {
			return "Asia/Tokyo"
		}
	}

	// Australia
	if lat >= -45 && lat <= -10 && lon >= 110 && lon <= 155 {
		if lon >= 110 && lon <= 130 {
			return "Australia/Perth"
		}
		if lon >= 130 && lon <= 145 {
			return "Australia/Adelaide"
		}
		if lon >= 145 && lon <= 155 {
			return "Australia/Sydney"
		}
	}

	// South America
	if lat >= -55 && lat <= 15 && lon >= -82 && lon <= -35 {
		if lon >= -82 && lon <= -70 {
			return "America/Lima"
		}
		if lon >= -70 && lon <= -50 {
			return "America/Santiago"
		}
		if lon >= -50 && lon <= -35 {
			return "America/Sao_Paulo"
		}
	}

	// Africa
	if lat >= -35 && lat <= 37 && lon >= -20 && lon <= 50 {
		if lon >= -20 && lon <= 0 {
			return "Africa/Casablanca"
		}
		if lon >= 0 && lon <= 20 {
			return "Africa/Lagos"
		}
		if lon >= 20 && lon <= 40 {
			return "Africa/Cairo"
		}
		if lon >= 40 && lon <= 50 {
			return "Africa/Nairobi"
		}
	}

	// Default to UTC
	return "UTC"
}

// Close closes the timezone cache
func (s *TimezoneLookupService) Close() {
	if s.cache != nil {
		s.cache.Close()
	}
}