package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/gin-gonic/gin"
)

// DailyUpdateHandler handles the scheduled daily update of the Yoto card
func (h *Handler) DailyUpdateHandler(c *gin.Context) {
	// Prevent recursive calls
	if c.GetHeader("X-Internal-Call") == "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recursive call detected"})
		return
	}

	schedulerToken := c.GetHeader("X-Scheduler-Token")
	expectedToken := h.config.SchedulerToken

	if expectedToken != "" && schedulerToken != expectedToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid scheduler token"})
		return
	}

	log.Printf("DailyUpdateHandler: Starting daily update from %s", c.ClientIP())

	// Test external connectivity
	testResp, err := http.Get("https://httpbin.org/get")
	if err != nil {
		log.Printf("DailyUpdateHandler: Failed to reach httpbin.org: %v", err)
	} else {
		testResp.Body.Close()
		log.Printf("DailyUpdateHandler: Successfully reached httpbin.org")
	}

	// Get a random bird from anywhere in the world
	// We'll use a random location to get diverse birds
	randomLocations := []struct {
		lat, lng float64
		name string
	}{
		{40.7128, -74.0060, "New York"},     // North America
		{51.5074, -0.1278, "London"},        // Europe  
		{-33.8688, 151.2093, "Sydney"},      // Australia
		{35.6762, 139.6503, "Tokyo"},        // Asia
		{-1.2921, 36.8219, "Nairobi"},       // Africa
		{-23.5505, -46.6333, "São Paulo"},   // South America
	}
	
	// Pick a random location for bird selection (changes daily)
	now := time.Now()
	dayIndex := (now.Year()*365 + now.YearDay()) % len(randomLocations)
	selectedLoc := randomLocations[dayIndex]
	
	location := &models.Location{
		Latitude:  selectedLoc.lat,
		Longitude: selectedLoc.lng,
		Region:    selectedLoc.name,
	}
	
	// Get bird of the day from that region
	bird, err := h.birdSelector.SelectBirdOfDay(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get bird: %v", err)})
		return
	}
	log.Printf("DailyUpdateHandler: Selected bird: %s", bird.CommonName)
	
	// Store this as the daily global bird for fallback use
	localDate := time.Now().Format("2006-01-02")
	h.updateCache.SetDailyGlobalBird(localDate, bird.CommonName)
	log.Printf("DailyUpdateHandler: Stored %s as global bird for %s", bird.CommonName, localDate)

	// Get a generic intro (no bird name mentioned)
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}
	log.Printf("DailyUpdateHandler: Using base URL: %s", baseURL)

	// Get both the intro URL and the voice ID to ensure consistency across all tracks
	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)
	log.Printf("DailyUpdateHandler: Got intro URL and voice ID")

	contentManager := h.yotoClient.NewContentManager()
	log.Printf("DailyUpdateHandler: Created content manager")

	// Create and link the bird content to your MYO card
	cardID := h.config.YotoCardID // Add this to config - your MYO card ID (e.g., "ipHAS")
	if cardID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "YOTO_CARD_ID not configured"})
		return
	}

	log.Printf("DailyUpdateHandler: About to update card %s with bird %s from %s", cardID, bird.CommonName, location.Region)
	// Always use generic facts for the daily update since users are worldwide
	// Individual users may get location-specific facts if the bird is in their region
	log.Printf("DailyUpdateHandler: Using generic update (global audience)")
	err = contentManager.UpdateExistingCardContentWithDescriptionAndVoice(
		cardID,
		bird.CommonName,
		introURL,
		bird.AudioURL,
		bird.Description,
		voiceID,
	)
	log.Printf("DailyUpdateHandler: Update completed (or failed)")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to update Yoto card: %v", err),
			"bird":  bird.CommonName,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Successfully updated card with %s", bird.CommonName),
		"bird":      bird.CommonName,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
