package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleYotoWebhookSimplified - Simplified webhook handler with clearer logic
func (h *Handler) HandleYotoWebhookSimplified(c *gin.Context) {
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

	// Get card ID
	cardID := webhook.CardID
	if cardID == "" {
		cardID = h.config.YotoCardID
	}

	// Step 1: Detect user's location
	clientIP := c.ClientIP()
	location, err := h.locationService.GetLocationFromIP(clientIP)
	
	log.Printf("[WEBHOOK] Initial location: %s, %s (lat: %f, lng: %f) from IP %s", 
		location.City, location.Country, location.Latitude, location.Longitude, clientIP)
	
	// Step 2: If location detection failed, try device timezone
	if err != nil || location.City == "London" {
		log.Printf("[WEBHOOK] IP location failed or defaulted, trying device timezone")
		
		if webhook.DeviceID != "" {
			deviceConfig, err := h.yotoClient.GetDeviceConfig(webhook.DeviceID)
			if err == nil && deviceConfig != nil && deviceConfig.Device.Config.GeoTimezone != "" {
				deviceTimezone := deviceConfig.Device.Config.GeoTimezone
				tzLocation := h.timezoneLocationService.GetLocationFromTimezone(deviceTimezone)
				
				// Only use timezone location if it's meaningful
				if tzLocation.City != "London" || deviceTimezone == "Europe/London" {
					location = tzLocation
					log.Printf("[WEBHOOK] Using timezone-based location: %s, %s (from timezone: %s)",
						location.City, location.Country, deviceTimezone)
				}
			}
		}
	}

	// Step 3: Get timezone for the location
	var tz *time.Location
	if h.timezoneLookup != nil {
		tz = h.timezoneLookup.GetTimezone(location.Latitude, location.Longitude)
	} else {
		tz, _ = time.LoadLocation("UTC")
	}
	
	localTime := time.Now().In(tz)
	localDate := localTime.Format("2006-01-02")

	// Step 4: Check cache for this location today
	locationKey := h.updateCache.GetLocationKey(location.Latitude, location.Longitude)
	
	if h.updateCache.HasBeenUpdated(cardID, localDate, locationKey) {
		cachedBird := h.updateCache.GetBirdName(cardID, localDate, locationKey)
		log.Printf("[WEBHOOK] Card already updated today for %s with %s", location.City, cachedBird)
		
		c.JSON(http.StatusOK, gin.H{
			"status":   "already_updated",
			"message":  fmt.Sprintf("Already showing %s for today", cachedBird),
			"bird":     cachedBird,
			"location": location.City,
			"date":     localDate,
		})
		return
	}

	// Step 5: Select a bird (uses cascading fallback internally)
	log.Printf("[WEBHOOK] Selecting bird for %s", location.City)
	bird, err := h.birdSelector.SelectBirdOfDay(location)
	
	if err != nil {
		log.Printf("[WEBHOOK] Failed to select bird: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to select bird"})
		return
	}

	log.Printf("[WEBHOOK] Selected bird: %s", bird.CommonName)

	// Step 6: Get intro and voice
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}
	
	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)

	// Step 7: Update the card
	contentManager := h.yotoClient.NewContentManager()
	
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

	// Step 8: Cache the update
	h.updateCache.MarkUpdated(cardID, localDate, locationKey, bird.CommonName)
	
	log.Printf("[WEBHOOK] Successfully updated card with %s for %s", bird.CommonName, location.City)
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  fmt.Sprintf("Card updated with %s", bird.CommonName),
		"bird":     bird.CommonName,
		"location": location.City,
		"date":     localDate,
		"timezone": tz.String(),
	})
}