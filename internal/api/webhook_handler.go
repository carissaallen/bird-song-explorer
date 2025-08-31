package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/callen/bird-song-explorer/internal/services"
	"github.com/gin-gonic/gin"
)

// YotoWebhookRequest represents the webhook payload from Yoto
type YotoWebhookRequest struct {
	EventType string                 `json:"eventType"`
	CardID    string                 `json:"cardId"`
	DeviceID  string                 `json:"deviceId"`
	UserID    string                 `json:"userId"`
	Timestamp string                 `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// HandleYotoWebhookV2 processes webhook events from Yoto when a card is played
func (h *Handler) HandleYotoWebhookV2(c *gin.Context) {
	var webhook YotoWebhookRequest
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	if webhook.EventType != "card.played" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "Not a card.played event"})
		return
	}

	if webhook.CardID != h.config.YotoCardID {
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "Not our card"})
		return
	}

	// First, try to get location from IP address (most accurate)
	clientIP := c.ClientIP()
	location, err := h.locationService.GetLocationFromIP(clientIP)
	var deviceTimezone string

	// Log successful IP-based location
	if err == nil && location.City != "Bend" {
		log.Printf("Using IP-based location: %s, %s (IP: %s)\n",
			location.City, location.Country, clientIP)
	} else {
		// If IP location failed or returned default, try device timezone as fallback
		if webhook.DeviceID != "" {
			deviceConfig, err := h.yotoClient.GetDeviceConfig(webhook.DeviceID)
			if err == nil && deviceConfig != nil {
				deviceTimezone = deviceConfig.Device.Config.GeoTimezone
			}
		}

		// If we have a device timezone, use it
		if deviceTimezone != "" {
			tzLocation := h.timezoneLocationService.GetLocationFromTimezone(deviceTimezone)
			// Only use timezone location if it's not the default
			if tzLocation.City != "Bend" || deviceTimezone == "America/Denver" {
				location = tzLocation
				fmt.Printf("Using timezone-based location: %s, %s (from timezone: %s)\n",
					location.City, location.Country, deviceTimezone)
			}
		}

		// If still using default location, log a warning
		if location.City == "Bend" && deviceTimezone != "America/Denver" {
			log.Printf("WARNING: Using default location (Bend, OR) - IP: %s, Timezone: %s\n",
				clientIP, deviceTimezone)
		}
	}

	// Determine local time for this location
	tz := GetTimezoneFromLocation(location.Latitude, location.Longitude)
	localTime := time.Now().In(tz)
	localDate := localTime.Format("2006-01-02")

	// Check if we need to update the card for this location's "today"
	// cacheKey := fmt.Sprintf("%s_%s", webhook.CardID, localDate) // For future use

	// Get the bird for this location and date
	bird, err := h.birdSelector.SelectBirdOfDay(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to select bird"})
		return
	}

	timeHelper := services.NewUserTimeHelper()
	userContext := timeHelper.GetUserTimeContext(deviceTimezone)
	log.Printf("User Time Context - Timezone: %s, Local Time: %v, Hour: %v, Nature Sound: %v",
		deviceTimezone, userContext["local_time"], userContext["hour"], userContext["nature_sound"])

	timezoneLogger := services.NewTimezoneLogger()
	timezoneLogger.LogTimezoneUsage(webhook.DeviceID, webhook.CardID, deviceTimezone, location.City)

	// Generate intro and get the voice ID that was used
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	var voiceID string

	var introURL string

	if h.audioManager.IsNatureMixEnabled() && deviceTimezone != "" {
		introURL, voiceID = h.audioManager.GetRandomIntroWithNatureForTimezone(baseURL, deviceTimezone)
		log.Printf("Using nature-mixed intro for timezone %s with %s sounds",
			deviceTimezone, userContext["nature_sound"])
	} else {
		// Fall back to regular intro selection
		var err error
		introURL, err = h.introGenerator.GetDynamicIntroURL(bird.CommonName, baseURL)
		if err != nil {
			// Fall back to static intro and get the voice ID
			introURL, voiceID = h.audioManager.GetRandomIntroURL(baseURL)
		} else {
			// The dynamic intro should use the same daily voice
			_, voiceID = h.audioManager.GetRandomIntroURL(baseURL)
		}
	}

	// Update the card with location-specific bird using the voice from the intro
	contentManager := h.yotoClient.NewContentManager()
	err = contentManager.UpdateExistingCardContentWithDescriptionAndVoice(
		webhook.CardID,
		bird.CommonName,
		introURL,
		bird.AudioURL,
		"", // No description for webhook
		voiceID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to update card: %v", err),
			"location":  location.City,
			"localDate": localDate,
		})
		return
	}

	// Include user time context in response
	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"message":        fmt.Sprintf("Card updated with %s for %s", bird.CommonName, location.City),
		"location":       location.City,
		"localDate":      localDate,
		"timezone":       tz.String(),
		"deviceTimezone": deviceTimezone,
		"userLocalTime":  userContext["local_time"],
		"userHour":       userContext["hour"],
		"natureSound":    userContext["nature_sound"],
		"timePeriod":     userContext["time_period"],
	})
}

// GetTimezoneFromLocation gets timezone from coordinates (simplified version)
func GetTimezoneFromLocation(lat, lon float64) *time.Location {
	// Simplified timezone detection based on longitude
	// In production, use a proper timezone library like timezonefinder

	if lon >= -125 && lon <= -115 && lat >= 30 && lat <= 50 {
		loc, _ := time.LoadLocation("America/Los_Angeles")
		return loc
	} else if lon >= -115 && lon <= -105 && lat >= 30 && lat <= 50 {
		loc, _ := time.LoadLocation("America/Denver")
		return loc
	} else if lon >= -105 && lon <= -90 && lat >= 30 && lat <= 50 {
		loc, _ := time.LoadLocation("America/Chicago")
		return loc
	} else if lon >= -90 && lon <= -70 && lat >= 30 && lat <= 50 {
		loc, _ := time.LoadLocation("America/New_York")
		return loc
	} else if lon >= -10 && lon <= 2 && lat >= 48 && lat <= 60 {
		loc, _ := time.LoadLocation("Europe/London")
		return loc
	} else if lon >= 2 && lon <= 25 && lat >= 40 && lat <= 60 {
		loc, _ := time.LoadLocation("Europe/Berlin")
		return loc
	} else if lon >= 140 && lon <= 155 && lat >= -40 && lat <= -25 {
		loc, _ := time.LoadLocation("Australia/Sydney")
		return loc
	}

	return time.UTC
}
