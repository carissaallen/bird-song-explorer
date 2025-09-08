package api

import "time"

// GetTimezoneFromLocation gets timezone from coordinates (simplified version)
func GetTimezoneFromLocation(lat, lon float64) *time.Location {
	// Simplified timezone detection based on longitude
	// In production, use a proper timezone library like timezonefinder

	// North America - expanded ranges for better coverage
	if lon >= -125 && lon <= -115 && lat >= 30 && lat <= 60 {
		loc, _ := time.LoadLocation("America/Los_Angeles")
		return loc
	} else if lon >= -115 && lon <= -105 && lat >= 30 && lat <= 60 {
		loc, _ := time.LoadLocation("America/Denver")
		return loc
	} else if lon >= -105 && lon <= -90 && lat >= 25 && lat <= 60 {
		loc, _ := time.LoadLocation("America/Chicago")
		return loc
	} else if lon >= -90 && lon <= -70 && lat >= 25 && lat <= 60 {
		loc, _ := time.LoadLocation("America/New_York")
		return loc
	}

	// Europe
	if lon >= -10 && lon <= 2 && lat >= 48 && lat <= 60 {
		loc, _ := time.LoadLocation("Europe/London")
		return loc
	} else if lon >= 2 && lon <= 25 && lat >= 40 && lat <= 60 {
		loc, _ := time.LoadLocation("Europe/Berlin")
		return loc
	}

	// Australia
	if lon >= 140 && lon <= 155 && lat >= -40 && lat <= -25 {
		loc, _ := time.LoadLocation("Australia/Sydney")
		return loc
	}

	// Asia
	if lon >= 135 && lon <= 145 && lat >= 30 && lat <= 45 {
		loc, _ := time.LoadLocation("Asia/Tokyo")
		return loc
	}

	// Default to UTC for any unmatched location
	loc, _ := time.LoadLocation("UTC")
	return loc
}