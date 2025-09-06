package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/gin-gonic/gin"
)

// HandleYotoWebhookV3 processes webhook events with location-based bird selection
func (h *Handler) HandleYotoWebhookV3(c *gin.Context) {
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

	// Use configured card ID or webhook card ID
	cardID := webhook.CardID
	if cardID == "" {
		cardID = h.config.YotoCardID
	}

	// Get user's location
	clientIP := c.ClientIP()
	location, err := h.locationService.GetLocationFromIP(clientIP)
	usingFallbackLocation := false
	
	// Try device timezone as fallback
	var deviceTimezone string
	if (err != nil || location.City == "London") && webhook.DeviceID != "" {
		deviceConfig, err := h.yotoClient.GetDeviceConfig(webhook.DeviceID)
		if err == nil && deviceConfig != nil {
			deviceTimezone = deviceConfig.Device.Config.GeoTimezone
			if deviceTimezone != "" {
				tzLocation := h.timezoneLocationService.GetLocationFromTimezone(deviceTimezone)
				if tzLocation.City != "London" {
					location = tzLocation
				} else if deviceTimezone != "Europe/London" {
					// We're using London as a fallback, not actual London location
					usingFallbackLocation = true
				}
			}
		} else if location.City == "London" {
			// No device timezone and defaulted to London
			usingFallbackLocation = true
		}
	} else if err != nil {
		// Location detection failed entirely
		usingFallbackLocation = true
	}

	// Log location detection
	if usingFallbackLocation {
		log.Printf("[WEBHOOK] Using fallback location (London) - will use global bird (IP: %s, TZ: %s)", 
			clientIP, deviceTimezone)
	} else {
		log.Printf("[WEBHOOK] Card played from %s, %s (IP: %s, TZ: %s)", 
			location.City, location.Country, clientIP, deviceTimezone)
	}

	// Get today's date for this location
	tz := GetTimezoneFromLocation(location.Latitude, location.Longitude)
	localTime := time.Now().In(tz)
	localDate := localTime.Format("2006-01-02")

	// Generate location key for caching (groups nearby users)
	locationKey := h.updateCache.GetLocationKey(location.Latitude, location.Longitude)

	// Check if we've already updated this card for this location today
	if h.updateCache.HasBeenUpdated(cardID, localDate, locationKey) {
		cachedBird := h.updateCache.GetBirdName(cardID, localDate, locationKey)
		log.Printf("[WEBHOOK] Card already updated today for %s with %s", location.City, cachedBird)
		
		c.JSON(http.StatusOK, gin.H{
			"status":   "already_updated",
			"message":  fmt.Sprintf("Already showing %s for %s today", cachedBird, location.City),
			"bird":     cachedBird,
			"location": location.City,
			"date":     localDate,
		})
		return
	}

	// Select a bird for this location OR use global bird if fallback
	var bird *models.Bird
	
	if usingFallbackLocation {
		// Try to get the daily global bird
		globalBirdName, hasGlobal := h.updateCache.GetDailyGlobalBird(localDate)
		if hasGlobal {
			log.Printf("[WEBHOOK] Using global bird %s for fallback location", globalBirdName)
			// We need to get the full bird details, not just the name
			// For now, we'll still need to select a bird but we'll use the global bird's name
			// This ensures consistency with the daily update
			bird, err = h.birdSelector.GetBirdByName(globalBirdName)
			if err != nil {
				log.Printf("[WEBHOOK] Failed to get global bird details for %s: %v", globalBirdName, err)
				// Fall back to selecting a new bird
				bird, err = h.birdSelector.SelectBirdOfDay(location)
			}
		} else {
			log.Printf("[WEBHOOK] No global bird cached, selecting new bird")
			bird, err = h.birdSelector.SelectBirdOfDay(location)
		}
	} else {
		// Select a regional bird for this actual location
		log.Printf("[WEBHOOK] Selecting regional bird for %s", location.City)
		bird, err = h.birdSelector.SelectBirdOfDay(location)
	}
	
	if err != nil || bird == nil {
		log.Printf("[WEBHOOK] Failed to select bird: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to select bird"})
		return
	}

	// Get intro and voice for consistency
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}
	
	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)

	// Update the card with location-specific bird and facts
	contentManager := h.yotoClient.NewContentManager()
	
	// Pass location coordinates for location-aware facts
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
		log.Printf("[WEBHOOK] Failed to update card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    fmt.Sprintf("Failed to update card: %v", err),
			"location": location.City,
		})
		return
	}

	// Mark this location as updated for today
	h.updateCache.MarkUpdated(cardID, localDate, locationKey, bird.CommonName)

	log.Printf("[WEBHOOK] Successfully updated card with %s for %s", bird.CommonName, location.City)
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  fmt.Sprintf("Card updated with %s for %s", bird.CommonName, location.City),
		"bird":     bird.CommonName,
		"location": location.City,
		"date":     localDate,
		"timezone": tz.String(),
	})
}