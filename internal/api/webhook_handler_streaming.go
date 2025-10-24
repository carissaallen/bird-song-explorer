package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/gin-gonic/gin"
)

// HandleYotoWebhookStreaming processes webhook events and updates card with streaming URLs
func (h *Handler) HandleYotoWebhookStreaming(c *gin.Context) {
	var webhook struct {
		EventType string `json:"eventType"`
		CardID    string `json:"cardId"`
		DeviceID  string `json:"deviceId"`
		UserID    string `json:"userId"`
	}

	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
		return
	}

	// Only process card.played events
	if webhook.EventType != "card.played" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "Not a card.played event"})
		return
	}

	// Check if streaming mode is enabled
	// Always use streaming mode with human voices from GCS
	log.Printf("[WEBHOOK-STREAMING] Processing card.played event with streaming mode")

	// Use configured card ID or webhook card ID
	cardID := webhook.CardID
	if cardID == "" {
		cardID = h.config.YotoCardID
	}

	// Attempt to detect user's location
	var userLocation *models.Location

	// Try IP-based location first
	clientIP := c.ClientIP()
	location, err := h.locationService.GetLocationFromIP(clientIP)
	if err == nil && location != nil {
		userLocation = location
		log.Printf("[WEBHOOK-STREAMING] Detected location from IP %s: %s, %s",
			clientIP, location.City, location.Country)
	} else {
		log.Printf("[WEBHOOK-STREAMING] Failed to detect location from IP %s: %v", clientIP, err)
	}

	// If IP detection failed, try device timezone
	if userLocation == nil && webhook.DeviceID != "" {
		deviceConfig, err := h.yotoClient.GetDeviceConfig(webhook.DeviceID)
		if err == nil && deviceConfig != nil && deviceConfig.Device.Config.GeoTimezone != "" {
			deviceTimezone := deviceConfig.Device.Config.GeoTimezone
			tzLocation := h.timezoneLocationService.GetLocationFromTimezone(deviceTimezone)
			if tzLocation != nil {
				userLocation = tzLocation
				log.Printf("[WEBHOOK-STREAMING] Detected location from timezone %s: %s, %s",
					deviceTimezone, tzLocation.City, tzLocation.Country)
			}
		}
	}

	// Get a new bird for this session (not daily bird, fresh each time)
	var bird *models.Bird
	var err2 error

	if userLocation != nil {
		// Select a regional bird for this location
		log.Printf("[WEBHOOK-STREAMING] Selecting bird for %s (lat: %f, lng: %f)",
			userLocation.City, userLocation.Latitude, userLocation.Longitude)
		bird, err2 = h.birdSelector.SelectBirdOfDay(userLocation)
		if err2 != nil {
			log.Printf("[WEBHOOK-STREAMING] Failed to select regional bird for %s: %v", userLocation.City, err2)
		} else {
			log.Printf("[WEBHOOK-STREAMING] Selected regional bird: %s for %s", bird.CommonName, userLocation.City)
		}
	}

	// If no regional bird or no location, select a global bird
	if bird == nil {
		log.Printf("[WEBHOOK-STREAMING] Selecting global bird")
		globalLocation := &models.Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			City:      "Global",
			Region:    "Global",
			Country:   "Global",
		}
		bird, err2 = h.birdSelector.SelectBirdOfDay(globalLocation)
		if err2 != nil {
			log.Printf("[WEBHOOK-STREAMING] Failed to select global bird: %v", err2)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to select any bird",
			})
			return
		}
		log.Printf("[WEBHOOK-STREAMING] Selected global bird: %s", bird.CommonName)
	}

	sessionID := fmt.Sprintf("%s_%d", cardID, time.Now().Unix())

	// Store session information for streaming tracks
	session := &StreamingSession{
		SessionID:      sessionID,
		Location:       userLocation,
		BirdName:       bird.CommonName,
		ScientificName: bird.ScientificName,
		BirdAudioURL:   bird.AudioURL,
		CreatedAt:      time.Now(),
	}

	// Get voice ID for consistency across tracks
	baseURL := h.config.BaseURL
	if baseURL == "" {
		// Fallback to request host if BASE_URL not configured
		baseURL = fmt.Sprintf("https://%s", c.Request.Host)
		if h.config.Environment == "development" {
			baseURL = fmt.Sprintf("http://%s", c.Request.Host)
		}
	}
	// Use a default voice since we're using human narration now
	session.VoiceID = "default"

	// Store session with the session ID as key
	sessionStore[sessionID] = session
	// Also store by IP for backward compatibility
	ipSessionKey := fmt.Sprintf("ip_%s", clientIP)
	sessionStore[ipSessionKey] = session

	log.Printf("[WEBHOOK-STREAMING] Created session %s with bird %s (scientific: %s, audio URL: %s)",
		sessionID, bird.CommonName, bird.ScientificName, bird.AudioURL)

	// Update the card content with streaming URLs for the new bird
	// Pass the session ID so the streaming URLs can reference it
	contentManager := h.yotoClient.NewContentManager()
	err = contentManager.UpdateCardWithStreamingTracks(cardID, bird.CommonName, baseURL, sessionID)
	if err != nil {
		log.Printf("[WEBHOOK-STREAMING] Failed to update card with streaming tracks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to update card: %v", err),
		})
		return
	}

	response := gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Card updated with streaming tracks for %s", bird.CommonName),
		"bird":    bird.CommonName,
	}

	if userLocation != nil {
		response["location"] = userLocation.City
		response["regional"] = true
	} else {
		response["location"] = "Global"
		response["regional"] = false
	}

	log.Printf("[WEBHOOK-STREAMING] Successfully updated card with streaming tracks for %s (location: %v)",
		bird.CommonName, userLocation != nil)

	c.JSON(http.StatusOK, response)
}
