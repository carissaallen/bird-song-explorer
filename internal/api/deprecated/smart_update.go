package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/gin-gonic/gin"
)

// SmartUpdateHandler updates the card with location-aware bird selection
// This can be called by:
// 1. Daily scheduler
// 2. Manual trigger
// 3. Webhook (if it ever works)
func (h *Handler) SmartUpdateHandler(c *gin.Context) {
	// Check if this is from scheduler
	schedulerToken := c.GetHeader("X-Scheduler-Token")
	isScheduler := schedulerToken == h.config.SchedulerToken && h.config.SchedulerToken != ""

	// Get card ID
	cardID := c.Param("cardId")
	if cardID == "" {
		cardID = h.config.YotoCardID
	}
	if cardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No card ID provided"})
		return
	}

	// Try to detect location from request
	clientIP := c.ClientIP()
	if clientIP == "::1" || clientIP == "127.0.0.1" {
		// Try to get real IP from headers
		if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
			clientIP = strings.Split(forwarded, ",")[0]
		}
	}

	location, err := h.locationService.GetLocationFromIP(clientIP)
	
	// If location detection failed or is default, try to use a stored location
	usingFallback := false
	if err != nil || location.City == "London" {
		log.Printf("[SMART_UPDATE] Location detection failed or defaulted for IP %s", clientIP)
		
		// For scheduler, use rotating global locations
		if isScheduler {
			location = h.getRotatingGlobalLocation()
			log.Printf("[SMART_UPDATE] Scheduler using rotating location: %s", location.City)
		} else {
			// For manual triggers, we're stuck with the default
			usingFallback = true
			log.Printf("[SMART_UPDATE] Using fallback location for manual trigger")
		}
	} else {
		log.Printf("[SMART_UPDATE] Detected location: %s, %s from IP %s", 
			location.City, location.Country, clientIP)
	}

	// Get timezone
	var tz *time.Location
	if h.timezoneLookup != nil {
		tz = h.timezoneLookup.GetTimezone(location.Latitude, location.Longitude)
	} else {
		tz, _ = time.LoadLocation("UTC")
	}
	
	localTime := time.Now().In(tz)
	localDate := localTime.Format("2006-01-02")

	// Check cache to prevent repeated updates
	locationKey := h.updateCache.GetLocationKey(location.Latitude, location.Longitude)
	
	// For scheduler, we might want to force update once per day
	// For manual triggers, check cache to prevent spam
	if !isScheduler && h.updateCache.HasBeenUpdated(cardID, localDate, locationKey) {
		cachedBird := h.updateCache.GetBirdName(cardID, localDate, locationKey)
		
		// If it's been less than an hour, don't update
		// This prevents spamming but allows multiple updates per day
		log.Printf("[SMART_UPDATE] Recently updated with %s for %s", cachedBird, location.City)
		
		c.JSON(http.StatusOK, gin.H{
			"status":   "recently_updated",
			"message":  fmt.Sprintf("Recently updated with %s", cachedBird),
			"bird":     cachedBird,
			"location": location.City,
		})
		return
	}

	// Select bird based on location
	log.Printf("[SMART_UPDATE] Selecting bird for %s", location.City)
	bird, err := h.birdSelector.SelectBirdOfDay(location)
	
	if err != nil {
		log.Printf("[SMART_UPDATE] Failed to select bird: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to select bird"})
		return
	}

	log.Printf("[SMART_UPDATE] Selected bird: %s", bird.CommonName)

	// Get intro and voice
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}
	
	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)

	// Update the card
	contentManager := h.yotoClient.NewContentManager()
	
	err = contentManager.UpdateExistingCardContentWithDescriptionVoiceAndLocation(
		cardID,
		bird.CommonName,
		introURL,
		bird.AudioURL,
		bird.Description,
		voiceID,
		location.Latitude,
		location.Longitude,
	)

	if err != nil {
		log.Printf("[SMART_UPDATE] Failed to update card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to update card: %v", err),
		})
		return
	}

	// Cache the update
	h.updateCache.MarkUpdated(cardID, localDate, locationKey, bird.CommonName)
	
	log.Printf("[SMART_UPDATE] Successfully updated card with %s for %s", 
		bird.CommonName, location.City)
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Card updated with %s", bird.CommonName),
		"bird":      bird.CommonName,
		"location":  location.City,
		"source":    map[string]bool{"scheduler": isScheduler, "fallback": usingFallback},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// getRotatingGlobalLocation returns a different location each day for variety
func (h *Handler) getRotatingGlobalLocation() *models.Location {
	locations := []struct {
		lat, lng float64
		name     string
		country  string
	}{
		{40.7128, -74.0060, "New York", "USA"},
		{51.5074, -0.1278, "London", "UK"},
		{-33.8688, 151.2093, "Sydney", "Australia"},
		{35.6762, 139.6503, "Tokyo", "Japan"},
		{-1.2921, 36.8219, "Nairobi", "Kenya"},
		{-23.5505, -46.6333, "São Paulo", "Brazil"},
		{64.1466, -21.9426, "Reykjavik", "Iceland"},
		{55.7558, 37.6173, "Moscow", "Russia"},
		{19.0760, 72.8777, "Mumbai", "India"},
		{-34.6037, -58.3816, "Buenos Aires", "Argentina"},
		{48.8566, 2.3522, "Paris", "France"},
		{52.5200, 13.4050, "Berlin", "Germany"},
		{37.7749, -122.4194, "San Francisco", "USA"},
		{1.3521, 103.8198, "Singapore", "Singapore"},
		{-26.2041, 28.0473, "Johannesburg", "South Africa"},
	}
	
	now := time.Now()
	dayIndex := (now.Year()*365 + now.YearDay()) % len(locations)
	selected := locations[dayIndex]
	
	return &models.Location{
		Latitude:  selected.lat,
		Longitude: selected.lng,
		City:      selected.name,
		Country:   selected.country,
		Region:    selected.name,
	}
}