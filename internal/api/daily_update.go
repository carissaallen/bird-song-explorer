package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	
	// Get today's bird for the default location
	location, _ := h.locationService.GetLocationFromIP("")

	// Get bird of the day
	bird, err := h.birdSelector.SelectBirdOfDay(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get bird: %v", err)})
		return
	}
	log.Printf("DailyUpdateHandler: Selected bird: %s", bird.CommonName)

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

	log.Printf("DailyUpdateHandler: About to update card %s with bird %s", cardID, bird.CommonName)
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
