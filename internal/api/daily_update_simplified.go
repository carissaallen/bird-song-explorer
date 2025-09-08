package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/gin-gonic/gin"
)

// DailyUpdateHandlerSimplified - Simplified daily update that just picks a global bird
func (h *Handler) DailyUpdateHandlerSimplified(c *gin.Context) {
	// Check scheduler token
	schedulerToken := c.GetHeader("X-Scheduler-Token")
	expectedToken := h.config.SchedulerToken

	if expectedToken != "" && schedulerToken != expectedToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid scheduler token"})
		return
	}

	log.Printf("[DAILY_UPDATE] Starting daily update")

	// Pick a diverse global location that changes daily
	locations := []struct {
		lat, lng float64
		name     string
	}{
		{40.7128, -74.0060, "New York"},        // North America
		{51.5074, -0.1278, "London"},           // Europe  
		{-33.8688, 151.2093, "Sydney"},         // Australia
		{35.6762, 139.6503, "Tokyo"},           // Asia
		{-1.2921, 36.8219, "Nairobi"},          // Africa
		{-23.5505, -46.6333, "São Paulo"},      // South America
		{64.1466, -21.9426, "Reykjavik"},       // Iceland
		{55.7558, 37.6173, "Moscow"},           // Russia
		{19.0760, 72.8777, "Mumbai"},           // India
		{-34.6037, -58.3816, "Buenos Aires"},   // Argentina
	}
	
	// Pick location based on day of year for variety
	now := time.Now()
	dayIndex := (now.Year()*365 + now.YearDay()) % len(locations)
	selectedLoc := locations[dayIndex]
	
	location := &models.Location{
		Latitude:  selectedLoc.lat,
		Longitude: selectedLoc.lng,
		City:      selectedLoc.name,
		Region:    selectedLoc.name,
	}
	
	log.Printf("[DAILY_UPDATE] Using location: %s (lat: %f, lng: %f)", 
		location.City, location.Latitude, location.Longitude)
	
	// Get bird from that location (will use cascading fallback if needed)
	bird, err := h.birdSelector.SelectBirdOfDay(location)
	if err != nil {
		log.Printf("[DAILY_UPDATE] Failed to select bird: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get bird: %v", err)})
		return
	}
	
	log.Printf("[DAILY_UPDATE] Selected bird: %s", bird.CommonName)

	// Get intro and voice
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}

	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)

	// Update the card
	contentManager := h.yotoClient.NewContentManager()
	cardID := h.config.YotoCardID
	
	if cardID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "YOTO_CARD_ID not configured"})
		return
	}

	log.Printf("[DAILY_UPDATE] Updating card %s with %s", cardID, bird.CommonName)
	
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
		log.Printf("[DAILY_UPDATE] Failed to update card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to update Yoto card: %v", err),
			"bird":  bird.CommonName,
		})
		return
	}

	log.Printf("[DAILY_UPDATE] Successfully updated card with %s from %s", 
		bird.CommonName, location.City)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Successfully updated card with %s", bird.CommonName),
		"bird":      bird.CommonName,
		"location":  location.City,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}