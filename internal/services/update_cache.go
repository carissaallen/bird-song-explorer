package services

import (
	"fmt"
	"sync"
	"time"
)

// UpdateCache tracks which cards have been updated for which locations today
type UpdateCache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
}

// CacheEntry stores information about an update
type CacheEntry struct {
	BirdName     string
	BirdAudioURL string // Store the audio URL to avoid re-fetching
	UpdatedAt    time.Time
	LocationKey  string
}

// NewUpdateCache creates a new cache
func NewUpdateCache() *UpdateCache {
	cache := &UpdateCache{
		entries: make(map[string]CacheEntry),
	}

	// Start cleanup goroutine to remove old entries at midnight
	go cache.cleanupLoop()

	return cache
}

// GetCacheKey generates a cache key for a card and date
func (uc *UpdateCache) GetCacheKey(cardID string, date string, locationKey string) string {
	// Include location in cache key so different locations get different birds
	return fmt.Sprintf("%s_%s_%s", cardID, date, locationKey)
}

// GetLocationKey generates a location identifier
func (uc *UpdateCache) GetLocationKey(latitude, longitude float64) string {
	// Round to 1 decimal place (about 11km precision)
	// This groups nearby users together
	return fmt.Sprintf("%.1f_%.1f", latitude, longitude)
}

// HasBeenUpdated checks if a card has been updated today for this location
func (uc *UpdateCache) HasBeenUpdated(cardID string, date string, locationKey string) bool {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	key := uc.GetCacheKey(cardID, date, locationKey)
	entry, exists := uc.entries[key]

	if !exists {
		return false
	}

	// Check if the entry is from today
	return entry.UpdatedAt.Format("2006-01-02") == date
}

// GetBirdName returns the bird name for a cached update
func (uc *UpdateCache) GetBirdName(cardID string, date string, locationKey string) string {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	key := uc.GetCacheKey(cardID, date, locationKey)
	entry, exists := uc.entries[key]

	if !exists {
		return ""
	}

	return entry.BirdName
}

// MarkUpdated records that a card has been updated
func (uc *UpdateCache) MarkUpdated(cardID string, date string, locationKey string, birdName string) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	key := uc.GetCacheKey(cardID, date, locationKey)
	uc.entries[key] = CacheEntry{
		BirdName:    birdName,
		UpdatedAt:   time.Now(),
		LocationKey: locationKey,
	}
}

// cleanupLoop removes old cache entries
func (uc *UpdateCache) cleanupLoop() {
	for {
		// Wait until midnight
		now := time.Now()
		midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		time.Sleep(time.Until(midnight))

		// Clear old entries
		uc.mu.Lock()
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		for key, entry := range uc.entries {
			if entry.UpdatedAt.Format("2006-01-02") <= yesterday {
				delete(uc.entries, key)
			}
		}
		uc.mu.Unlock()
	}
}

// GetStats returns cache statistics for debugging
func (uc *UpdateCache) GetStats() map[string]interface{} {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	uniqueLocations := make(map[string]bool)
	for _, entry := range uc.entries {
		uniqueLocations[entry.LocationKey] = true
	}

	return map[string]interface{}{
		"total_entries":    len(uc.entries),
		"unique_locations": len(uniqueLocations),
	}
}

// SetDailyGlobalBird stores the daily global bird selected by the scheduler
func (uc *UpdateCache) SetDailyGlobalBird(date string, birdName string) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	// Use a special key for the global daily bird
	key := fmt.Sprintf("GLOBAL_DAILY_%s", date)
	uc.entries[key] = CacheEntry{
		BirdName:    birdName,
		UpdatedAt:   time.Now(),
		LocationKey: "GLOBAL",
	}
}

// GetDailyGlobalBird retrieves the daily global bird for fallback
func (uc *UpdateCache) GetDailyGlobalBird(date string) (string, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	key := fmt.Sprintf("GLOBAL_DAILY_%s", date)
	entry, exists := uc.entries[key]

	if !exists {
		return "", false
	}

	return entry.BirdName, true
}
