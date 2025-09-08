package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	if err == nil && location.City != "London" {
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
			if tzLocation.City != "London" || deviceTimezone == "Europe/London" {
				location = tzLocation
				fmt.Printf("Using timezone-based location: %s, %s (from timezone: %s)\n",
					location.City, location.Country, deviceTimezone)
			}
		}

		// If still using default location, log a warning
		if location.City == "London" && deviceTimezone != "Europe/London" {
			log.Printf("WARNING: Using default location (London, UK) - IP: %s, Timezone: %s\n",
				clientIP, deviceTimezone)
		}
	}

	// Determine local time for this location
	var tz *time.Location
	if h.timezoneLookup != nil {
		tz = h.timezoneLookup.GetTimezone(location.Latitude, location.Longitude)
	} else {
		// Fallback to the old method if timezone lookup service is not available
		tz = GetTimezoneFromLocation(location.Latitude, location.Longitude)
	}
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

	// Generate intro and get the voice ID that was used
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	var voiceID string
	introURL, err := h.introGenerator.GetDynamicIntroURL(bird.CommonName, baseURL)
	if err != nil {
		// Fall back to static intro and get the voice ID
		introURL, voiceID = h.audioManager.GetRandomIntroURL(baseURL)
	} else {
		// When using dynamic intro, we also need to get the voice ID
		// The dynamic intro should use the same daily voice
		_, voiceID = h.audioManager.GetRandomIntroURL(baseURL)
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

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"message":        fmt.Sprintf("Card updated with %s for %s", bird.CommonName, location.City),
		"location":       location.City,
		"localDate":      localDate,
		"timezone":       tz.String(),
		"deviceTimezone": deviceTimezone,
	})
}

// GetTimezoneFromLocation gets timezone from coordinates (simplified version)
func GetTimezoneFromLocation(lat, lon float64) *time.Location {
	// Simplified timezone detection based on longitude
	// In production, use a proper timezone library like timezonefinder

	// North America - expanded ranges for better coverage
	if lon >= -125 && lon <= -115 && lat >= 30 && lat <= 60 {
		loc, _ := time.LoadLocation("America/Los_Angeles")
		return loc
	} else if lon >= -115 && lon <= -105 && lat >= 30 && lat <= 60 {
		loc, _ := time.LoadLocation("America/Denver")
		return loc
	} else if lon >= -105 && lon <= -90 && lat >= 25 && lat <= 60 {
		loc, _ := time.LoadLocation("America/Chicago")
		return loc
	} else if lon >= -90 && lon <= -70 && lat >= 25 && lat <= 60 {
		loc, _ := time.LoadLocation("America/New_York")
		return loc
	}
	
	// Europe
	if lon >= -10 && lon <= 2 && lat >= 48 && lat <= 60 {
		loc, _ := time.LoadLocation("Europe/London")
		return loc
	} else if lon >= 2 && lon <= 25 && lat >= 40 && lat <= 60 {
		loc, _ := time.LoadLocation("Europe/Berlin")
		return loc
	}
	
	// Australia
	if lon >= 140 && lon <= 155 && lat >= -40 && lat <= -25 {
		loc, _ := time.LoadLocation("Australia/Sydney")
		return loc
	}
	
	// Asia
	if lon >= 135 && lon <= 145 && lat >= 30 && lat <= 45 {
		loc, _ := time.LoadLocation("Asia/Tokyo")
		return loc
	}

	// Default to UTC for any unmatched location
	loc, _ := time.LoadLocation("UTC")
	return loc
}
