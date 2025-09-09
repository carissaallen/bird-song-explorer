package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
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

	// Get a random bird from diverse locations with focus on US, UK, Mexico, Canada
	randomLocations := []struct {
		lat, lng float64
		name     string
	}{
		// United States (expanded coverage)
		{40.7128, -74.0060, "New York"},
		{34.0522, -118.2437, "Los Angeles"},
		{41.8781, -87.6298, "Chicago"},
		{29.7604, -95.3698, "Houston"},
		{33.4484, -112.0740, "Phoenix"},
		{39.7392, -104.9903, "Denver"},
		{47.6062, -122.3321, "Seattle"},
		{25.7617, -80.1918, "Miami"},
		{42.3601, -71.0589, "Boston"},
		{37.7749, -122.4194, "San Francisco"},

		// Canada (expanded coverage)
		{43.6532, -79.3832, "Toronto"},
		{45.5017, -73.5673, "Montreal"},
		{49.2827, -123.1207, "Vancouver"},
		{51.0447, -114.0719, "Calgary"},
		{53.5461, -113.4938, "Edmonton"},
		{45.4215, -75.6972, "Ottawa"},

		// United Kingdom (expanded coverage)
		{51.5074, -0.1278, "London"},
		{53.4808, -2.2426, "Manchester"},
		{55.9533, -3.1883, "Edinburgh"},
		{52.4862, -1.8904, "Birmingham"},
		{51.4545, -2.5879, "Bristol"},
		{53.8008, -1.5491, "Leeds"},

		// Mexico (expanded coverage)
		{19.4326, -99.1332, "Mexico City"},
		{20.6597, -103.3496, "Guadalajara"},
		{25.6866, -100.3161, "Monterrey"},
		{21.1619, -86.8515, "Cancun"},
		{32.5149, -117.0382, "Tijuana"},
		{31.6904, -106.4245, "Ciudad Juárez"},

		// Other diverse locations
		{-33.8688, 151.2093, "Sydney"},
		{35.6762, 139.6503, "Tokyo"},
		{-1.2921, 36.8219, "Nairobi"},
		{-23.5505, -46.6333, "São Paulo"},
		{48.8566, 2.3522, "Paris"},
		{52.5200, 13.4050, "Berlin"},
		{55.7558, 37.6173, "Moscow"},
		{19.0760, 72.8777, "Mumbai"},
		{1.3521, 103.8198, "Singapore"},
		{-34.6037, -58.3816, "Buenos Aires"},
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
	h.updateCache.SetDailyGlobalBirdWithAudio(localDate, bird.CommonName, bird.AudioURL)
	log.Printf("DailyUpdateHandler: Stored %s as global bird for %s with audio URL", bird.CommonName, localDate)

	// Get a generic intro (no bird name mentioned)
	// Use the configured service URL or fall back to host
	baseURL := os.Getenv("SERVICE_URL")
	if baseURL == "" {
		// Fall back to using request host
		baseURL = fmt.Sprintf("https://%s", c.Request.Host)
		if h.config.Environment == "development" {
			baseURL = fmt.Sprintf("http://%s", c.Request.Host)
		}
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

	log.Printf("DailyUpdateHandler: About to update card %s with bird %s (using streaming for dynamic content)",
		cardID, bird.CommonName)

	// Check if we should use streaming
	useStreaming := os.Getenv("USE_STREAMING")
	if useStreaming == "true" {
		// Use streaming URLs for dynamic content
		log.Printf("DailyUpdateHandler: Using streaming URLs for dynamic location-aware content")
		err = contentManager.UpdateCardWithStreamingTracks(cardID, bird.CommonName, baseURL, "")
	} else {
		// Use the old approach with pre-uploaded content
		log.Printf("DailyUpdateHandler: Using pre-uploaded content (legacy mode)")
		err = contentManager.UpdateExistingCardContentWithDescriptionVoiceAndLocation(
			cardID,
			bird.CommonName,
			introURL,
			bird.AudioURL,
			bird.Description,
			voiceID,
			0, // No latitude - triggers generic facts
			0, // No longitude - triggers generic facts
		)
	}
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
		"message":   fmt.Sprintf("Successfully set daily bird as %s (generic facts)", bird.CommonName),
		"bird":      bird.CommonName,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
