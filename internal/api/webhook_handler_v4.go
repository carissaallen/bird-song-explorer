package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
	"github.com/gin-gonic/gin"
)

// HandleYotoWebhookV4 processes webhook events with improved location handling
// This version removes London defaults and adds regional bird checking
func (h *Handler) HandleYotoWebhookV4(c *gin.Context) {
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

	// Attempt to detect user's location
	var userLocation *models.Location
	var locationSource string
	
	// Try IP-based location first
	clientIP := c.ClientIP()
	location, err := h.locationService.GetLocationFromIP(clientIP)
	if err == nil && location != nil {
		userLocation = location
		locationSource = "ip"
		log.Printf("[WEBHOOK] Detected location from IP %s: %s, %s", 
			clientIP, location.City, location.Country)
	} else {
		log.Printf("[WEBHOOK] Failed to detect location from IP %s: %v", clientIP, err)
	}
	
	// If IP detection failed, try device timezone
	if userLocation == nil && webhook.DeviceID != "" {
		deviceConfig, err := h.yotoClient.GetDeviceConfig(webhook.DeviceID)
		if err == nil && deviceConfig != nil && deviceConfig.Device.Config.GeoTimezone != "" {
			deviceTimezone := deviceConfig.Device.Config.GeoTimezone
			tzLocation := h.timezoneLocationService.GetLocationFromTimezone(deviceTimezone)
			if tzLocation != nil {
				userLocation = tzLocation
				locationSource = "timezone"
				log.Printf("[WEBHOOK] Detected location from timezone %s: %s, %s", 
					deviceTimezone, tzLocation.City, tzLocation.Country)
			}
		}
	}

	// Get today's date (use UTC if no location detected)
	var tz *time.Location
	if userLocation != nil {
		if h.timezoneLookup != nil {
			tz = h.timezoneLookup.GetTimezone(userLocation.Latitude, userLocation.Longitude)
		} else {
			tz = GetTimezoneFromLocation(userLocation.Latitude, userLocation.Longitude)
		}
	}
	if tz == nil {
		tz, _ = time.LoadLocation("UTC")
	}
	
	localTime := time.Now().In(tz)
	localDate := localTime.Format("2006-01-02")

	// Get the daily bird that was set by the scheduler
	dailyBirdName, dailyBirdAudio, hasDailyBird := h.updateCache.GetDailyGlobalBirdWithAudio(localDate)
	if !hasDailyBird {
		log.Printf("[WEBHOOK] No daily bird set for %s", localDate)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No daily bird available. Please wait for the daily update.",
		})
		return
	}

	// Check if we've already updated for this user's location today (if we have location)
	var locationKey string
	if userLocation != nil {
		locationKey = h.updateCache.GetLocationKey(userLocation.Latitude, userLocation.Longitude)
		
		// Check if already updated for this location
		if h.updateCache.HasBeenUpdated(cardID, localDate, locationKey) {
			cachedBird := h.updateCache.GetBirdName(cardID, localDate, locationKey)
			log.Printf("[WEBHOOK] Already updated for %s with %s", userLocation.City, cachedBird)
			
			c.JSON(http.StatusOK, gin.H{
				"status":   "already_updated",
				"message":  fmt.Sprintf("Already showing %s for %s today", cachedBird, userLocation.City),
				"bird":     cachedBird,
				"location": userLocation.City,
				"date":     localDate,
			})
			return
		}
	}

	// Check if the daily bird is regional to the user (if we have their location)
	var isRegional bool
	var regionalMessage string
	
	if userLocation != nil {
		// Create regional checker
		regionalChecker := services.NewBirdRegionalChecker(h.config.EBirdAPIKey)
		
		// Check if bird has been spotted within 160km in last 30 days
		isRegional, err = regionalChecker.IsRegionalBird(dailyBirdName, userLocation, 160, 30)
		if err != nil {
			log.Printf("[WEBHOOK] Failed to check regionality for %s: %v", dailyBirdName, err)
		}
		
		// Get appropriate message
		regionalMessage = regionalChecker.GetRegionalityMessage(dailyBirdName, userLocation, isRegional)
		
		log.Printf("[WEBHOOK] Bird %s is regional to %s: %v", 
			dailyBirdName, userLocation.City, isRegional)
	} else {
		log.Printf("[WEBHOOK] No user location detected, using generic facts")
	}

	// Get intro and voice for consistency
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}
	
	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)

	// Build bird description with regional context
	birdDescription := fmt.Sprintf("Today's bird is the %s.", dailyBirdName)
	if regionalMessage != "" {
		birdDescription = fmt.Sprintf("%s %s", birdDescription, regionalMessage)
	}

	// Update the card
	contentManager := h.yotoClient.NewContentManager()
	
	// Use location coordinates for facts if available, otherwise use 0,0 for generic facts
	factLat := float64(0)
	factLng := float64(0)
	if userLocation != nil {
		// Always use location coordinates when we have them
		// The fact generator will handle regional vs non-regional appropriately
		factLat = userLocation.Latitude
		factLng = userLocation.Longitude
		if isRegional {
			log.Printf("[WEBHOOK] Using location-aware facts for %s (bird is regional)", userLocation.City)
		} else {
			log.Printf("[WEBHOOK] Using location-aware facts for %s (bird not regional but location known)", userLocation.City)
		}
	} else {
		log.Printf("[WEBHOOK] Using generic facts (no location detected)")
	}
	
	err = contentManager.UpdateExistingCardContentWithDescriptionVoiceAndLocation(
		cardID,
		dailyBirdName,
		introURL,
		dailyBirdAudio,
		birdDescription,
		voiceID,
		factLat,
		factLng,
	)

	if err != nil {
		log.Printf("[WEBHOOK] Failed to update card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to update card: %v", err),
		})
		return
	}

	// Cache the update if we have location
	if userLocation != nil && locationKey != "" {
		h.updateCache.MarkUpdated(cardID, localDate, locationKey, dailyBirdName)
		log.Printf("[WEBHOOK] Cached update for location %s", locationKey)
	}

	// Build response
	response := gin.H{
		"status":      "success",
		"bird":        dailyBirdName,
		"date":        localDate,
		"isRegional":  isRegional,
	}
	
	if userLocation != nil {
		response["location"] = userLocation.City
		response["locationSource"] = locationSource
		response["message"] = fmt.Sprintf("Card updated with %s for %s", dailyBirdName, userLocation.City)
	} else {
		response["location"] = "Unknown"
		response["locationSource"] = "none"
		response["message"] = fmt.Sprintf("Card updated with %s (generic facts)", dailyBirdName)
	}
	
	if regionalMessage != "" {
		response["regionalMessage"] = regionalMessage
	}

	log.Printf("[WEBHOOK] Successfully updated card with %s (regional: %v, location: %v)", 
		dailyBirdName, isRegional, userLocation != nil)
	
	c.JSON(http.StatusOK, response)
}