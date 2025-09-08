package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleYotoWebhookUnified handles webhook events from Yoto and uses smart update logic
func (h *Handler) HandleYotoWebhookUnified(c *gin.Context) {
	var webhook struct {
		EventType string `json:"eventType"`
		CardID    string `json:"cardId"`
		DeviceID  string `json:"deviceId"`
		UserID    string `json:"userId"`
		Timestamp string `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Printf("[WEBHOOK] Failed to parse webhook data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
		return
	}

	log.Printf("[WEBHOOK] Received event: %s for card: %s from device: %s", 
		webhook.EventType, webhook.CardID, webhook.DeviceID)

	// Only process card.played events
	if webhook.EventType != "card.played" {
		log.Printf("[WEBHOOK] Ignoring non-card.played event: %s", webhook.EventType)
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "Not a card.played event"})
		return
	}

	// Get card ID
	cardID := webhook.CardID
	if cardID == "" {
		cardID = h.config.YotoCardID
	}
	if cardID == "" {
		log.Printf("[WEBHOOK] No card ID available")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No card ID provided"})
		return
	}

	log.Printf("[WEBHOOK] Processing card.played for card %s", cardID)

	// Step 1: Detect user's location from IP
	clientIP := c.ClientIP()
	location, err := h.locationService.GetLocationFromIP(clientIP)
	
	log.Printf("[WEBHOOK] Location detection from IP %s: %s, %s (err: %v)", 
		clientIP, location.City, location.Country, err)
	
	// Step 2: If location failed or is London default, try device timezone
	if err != nil || location.City == "London" {
		log.Printf("[WEBHOOK] IP location failed/defaulted, trying device timezone for device %s", 
			webhook.DeviceID)
		
		if webhook.DeviceID != "" {
			deviceConfig, err := h.yotoClient.GetDeviceConfig(webhook.DeviceID)
			if err == nil && deviceConfig != nil && deviceConfig.Device.Config.GeoTimezone != "" {
				deviceTimezone := deviceConfig.Device.Config.GeoTimezone
				tzLocation := h.timezoneLocationService.GetLocationFromTimezone(deviceTimezone)
				
				log.Printf("[WEBHOOK] Device timezone %s maps to %s, %s", 
					deviceTimezone, tzLocation.City, tzLocation.Country)
				
				// Only use if it's not also defaulting to London
				if tzLocation.City != "London" || deviceTimezone == "Europe/London" {
					location = tzLocation
					log.Printf("[WEBHOOK] Using timezone-based location: %s", location.City)
				}
			} else {
				log.Printf("[WEBHOOK] Failed to get device config or timezone: %v", err)
			}
		}
	}

	log.Printf("[WEBHOOK] Final location: %s, %s (lat: %f, lng: %f)", 
		location.City, location.Country, location.Latitude, location.Longitude)

	// Step 3: Get timezone for the location
	var tz *time.Location
	if h.timezoneLookup != nil {
		tz = h.timezoneLookup.GetTimezone(location.Latitude, location.Longitude)
	} else {
		tz, _ = time.LoadLocation("UTC")
	}
	
	localTime := time.Now().In(tz)
	localDate := localTime.Format("2006-01-02")

	log.Printf("[WEBHOOK] Local date/time: %s in timezone %s", localDate, tz.String())

	// Step 4: Check cache to prevent rapid updates
	locationKey := h.updateCache.GetLocationKey(location.Latitude, location.Longitude)
	
	if h.updateCache.HasBeenUpdated(cardID, localDate, locationKey) {
		cachedBird := h.updateCache.GetBirdName(cardID, localDate, locationKey)
		log.Printf("[WEBHOOK] Card already updated today with %s for %s", cachedBird, location.City)
		
		c.JSON(http.StatusOK, gin.H{
			"status":   "already_updated",
			"message":  fmt.Sprintf("Already showing %s for today", cachedBird),
			"bird":     cachedBird,
			"location": location.City,
			"date":     localDate,
		})
		return
	}

	// Step 5: Select bird using cascading search
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
	log.Printf("[WEBHOOK] Got intro URL and voice ID")

	// Step 7: Update the card
	contentManager := h.yotoClient.NewContentManager()
	
	log.Printf("[WEBHOOK] Updating card %s with %s", cardID, bird.CommonName)
	
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
		"source":   "webhook",
	})
}