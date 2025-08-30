package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DailyUpdateHandler handles the scheduled daily update of the Yoto card
func (h *Handler) DailyUpdateHandler(c *gin.Context) {
	schedulerToken := c.GetHeader("X-Scheduler-Token")
	expectedToken := h.config.SchedulerToken

	if expectedToken != "" && schedulerToken != expectedToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid scheduler token"})
		return
	}

	// Get today's bird for the default location
	location, _ := h.locationService.GetLocationFromIP("")

	// Get bird of the day
	bird, err := h.birdSelector.SelectBirdOfDay(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get bird: %v", err)})
		return
	}

	// Get a generic intro (no bird name mentioned)
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}

	// Get both the intro URL and the voice ID to ensure consistency across all tracks
	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)

	contentManager := h.yotoClient.NewContentManager()

	// Create and link the bird content to your MYO card
	cardID := h.config.YotoCardID // Add this to config - your MYO card ID (e.g., "ipHAS")
	if cardID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "YOTO_CARD_ID not configured"})
		return
	}

	err = contentManager.UpdateExistingCardContentWithDescriptionAndVoice(
		cardID,
		bird.CommonName,
		introURL,
		bird.AudioURL,
		bird.Description,
		voiceID,
	)

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
