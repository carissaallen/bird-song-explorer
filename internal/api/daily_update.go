package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
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

	// Always select bird from available prerecorded birds (streaming mode only)
	bird := h.availableBirds.GetCyclingBird()
	log.Printf("DailyUpdateHandler: Selected bird: %s", bird.CommonName)

	// Store this as the daily global bird for fallback use
	localDate := time.Now().Format("2006-01-02")
	h.updateCache.SetDailyGlobalBirdWithAudio(localDate, bird.CommonName, bird.AudioURL)
	log.Printf("DailyUpdateHandler: Stored %s as global bird for %s with audio URL", bird.CommonName, localDate)

	// Get a generic intro (no bird name mentioned)
	// Use the configured service URL or fall back to host
	baseURL := os.Getenv("SERVICE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s", c.Request.Host)
		if h.config.Environment == "development" {
			baseURL = fmt.Sprintf("http://%s", c.Request.Host)
		}
	}

	contentManager := h.yotoClient.NewContentManager()

	cardID := h.config.YotoCardID
	if cardID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "YOTO_CARD_ID not configured"})
		return
	}

	// Create session BEFORE updating card to ensure icon and bird name match
	sessionID := h.CreateSessionForBird(cardID, bird.CommonName)
	log.Printf("[DAILY_UPDATE] Created session %s for bird: %s", sessionID, bird.CommonName)

	err = contentManager.UpdateCardWithStreamingTracks(cardID, bird.CommonName, baseURL, sessionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to update Yoto card: %v", err),
			"bird":  bird.CommonName,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Successfully set daily bird as %s (generic facts)", bird.CommonName),
		"bird":      bird.CommonName,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
