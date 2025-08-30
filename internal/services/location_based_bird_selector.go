package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
)

// TimezoneBirdSelector manages bird selection based on user's approximate location
type TimezoneBirdSelector struct {
	birdSelector *BirdSelector
	cache        map[string]*models.Bird
	cacheMutex   sync.RWMutex
}

// GetBirdForLocation gets the bird of the day for a specific location and its local date
func (tbs *TimezoneBirdSelector) GetBirdForLocation(location *models.Location) (*models.Bird, error) {
	localDate := tbs.getLocalDate(location)

	cacheKey := fmt.Sprintf("%.2f_%.2f_%s",
		location.Latitude,
		location.Longitude,
		localDate.Format("2006-01-02"))

	tbs.cacheMutex.RLock()
	if cachedBird, exists := tbs.cache[cacheKey]; exists {
		tbs.cacheMutex.RUnlock()
		return cachedBird, nil
	}
	tbs.cacheMutex.RUnlock()

	// Select a new bird for this location and date
	bird, err := tbs.birdSelector.SelectBirdOfDay(location)
	if err != nil {
		return nil, err
	}

	tbs.cacheMutex.Lock()
	tbs.cache[cacheKey] = bird
	tbs.cacheMutex.Unlock()

	// Clean old cache entries periodically
	go tbs.cleanOldCache()

	return bird, nil
}

// getLocalDate determines the local date for a given location
func (tbs *TimezoneBirdSelector) getLocalDate(location *models.Location) time.Time {
	// Estimate timezone offset based on longitude
	hoursOffset := int(location.Longitude / 15.0)

	utcTime := time.Now().UTC()
	localTime := utcTime.Add(time.Duration(hoursOffset) * time.Hour)

	return time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, time.UTC)
}

// cleanOldCache removes cache entries older than 2 days
func (tbs *TimezoneBirdSelector) cleanOldCache() {
	tbs.cacheMutex.Lock()
	defer tbs.cacheMutex.Unlock()

	cutoffDate := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	for key := range tbs.cache {
		keyParts := []byte(key)
		lastUnderscore := -1
		for i := len(keyParts) - 1; i >= 0; i-- {
			if keyParts[i] == '_' {
				lastUnderscore = i
				break
			}
		}

		if lastUnderscore > 0 {
			dateStr := string(keyParts[lastUnderscore+1:])
			if dateStr < cutoffDate {
				delete(tbs.cache, key)
			}
		}
	}
}
