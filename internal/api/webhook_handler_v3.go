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
	
	log.Printf("[WEBHOOK] Initial location detection: %s, %s (lat: %f, lng: %f) from IP %s", 
		location.City, location.Country, location.Latitude, location.Longitude, clientIP)
	
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
	var tz *time.Location
	if h.timezoneLookup != nil {
		tz = h.timezoneLookup.GetTimezone(location.Latitude, location.Longitude)
	} else {
		// Fallback to the old method if timezone lookup service is not available
		tz = GetTimezoneFromLocation(location.Latitude, location.Longitude)
		if tz == nil {
			log.Printf("[WEBHOOK] Warning: timezone detection returned nil, using UTC")
			tz, _ = time.LoadLocation("UTC")
		}
	}
	localTime := time.Now().In(tz)
	localDate := localTime.Format("2006-01-02")

	// Generate location key for caching (groups nearby users)
	locationKey := h.updateCache.GetLocationKey(location.Latitude, location.Longitude)

	// Only check cache for real locations, not fallback
	// This allows the card to update when proper location is detected later
	if !usingFallbackLocation {
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
	} else {
		log.Printf("[WEBHOOK] Skipping cache check for fallback location")
	}

	// Select a bird for this location OR use global bird if fallback
	var bird *models.Bird
	
	if usingFallbackLocation {
		// Try to get the daily global bird with cached audio
		globalBirdName, globalAudioURL, hasGlobal := h.updateCache.GetDailyGlobalBirdWithAudio(localDate)
		if hasGlobal && globalAudioURL != "" {
			log.Printf("[WEBHOOK] Using cached global bird %s with audio for fallback location", globalBirdName)
			// Create bird with cached data
			bird = &models.Bird{
				CommonName:     globalBirdName,
				AudioURL:       globalAudioURL,
				Region:         "Global",
				Description:    fmt.Sprintf("The %s is a fascinating bird found in many regions around the world.", globalBirdName),
			}
		} else if hasGlobal {
			log.Printf("[WEBHOOK] Using global bird %s for fallback location (no cached audio)", globalBirdName)
			// Get the full bird details for the global bird
			bird, err = h.birdSelector.GetBirdByName(globalBirdName)
			if err != nil {
				log.Printf("[WEBHOOK] Failed to get global bird details for %s: %v, selecting new global bird", globalBirdName, err)
				// Select a new global bird from random locations
				globalLocation := &models.Location{
					Latitude:  40.7128,
					Longitude: -74.0060,
					City:      "Global",
					Region:    "Global",
					Country:   "Global",
				}
				bird, err = h.birdSelector.SelectBirdOfDay(globalLocation)
				if err != nil {
					log.Printf("[WEBHOOK] Failed to select new global bird: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to select any bird",
					})
					return
				}
				// Cache the new global bird with audio
				h.updateCache.SetDailyGlobalBirdWithAudio(localDate, bird.CommonName, bird.AudioURL)
				log.Printf("[WEBHOOK] Selected and cached new global bird: %s", bird.CommonName)
			}
		} else {
			log.Printf("[WEBHOOK] No global bird cached, selecting new global bird")
			// Select a new global bird from random locations
			globalLocation := &models.Location{
				Latitude:  40.7128,
				Longitude: -74.0060,
				City:      "Global",
				Region:    "Global",
				Country:   "Global",
			}
			bird, err = h.birdSelector.SelectBirdOfDay(globalLocation)
			if err != nil {
				log.Printf("[WEBHOOK] Failed to select global bird: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to select any bird",
				})
				return
			}
			// Cache the new global bird with audio
			h.updateCache.SetDailyGlobalBirdWithAudio(localDate, bird.CommonName, bird.AudioURL)
			log.Printf("[WEBHOOK] Selected and cached new global bird: %s", bird.CommonName)
		}
	} else {
		// Select a regional bird for this actual location
		log.Printf("[WEBHOOK] Selecting regional bird for %s (lat: %f, lng: %f)", 
			location.City, location.Latitude, location.Longitude)
		bird, err = h.birdSelector.SelectBirdOfDay(location)
		if err != nil {
			log.Printf("[WEBHOOK] Failed to select regional bird for %s: %v", location.City, err)
		} else {
			log.Printf("[WEBHOOK] Successfully selected regional bird: %s for %s", bird.CommonName, location.City)
		}
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
	// If using fallback, pass 0,0 to trigger generic facts instead of London facts
	factLat := location.Latitude
	factLng := location.Longitude
	if usingFallbackLocation {
		factLat = 0
		factLng = 0
		log.Printf("[WEBHOOK] Using generic facts (no location coordinates)")
	} else {
		log.Printf("[WEBHOOK] Using location-aware facts for %s (lat:%f, lng:%f)", 
			location.City, factLat, factLng)
	}
	
	err = contentManager.UpdateExistingCardContentWithDescriptionVoiceAndLocation(
		cardID,
		bird.CommonName,
		introURL,
		bird.AudioURL,
		bird.Description,
		voiceID,
		factLat,
		factLng,
	)

	if err != nil {
		log.Printf("[WEBHOOK] Failed to update card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    fmt.Sprintf("Failed to update card: %v", err),
			"location": location.City,
		})
		return
	}

	// Mark this location as updated for today (only for real locations)
	if !usingFallbackLocation {
		h.updateCache.MarkUpdated(cardID, localDate, locationKey, bird.CommonName)
		log.Printf("[WEBHOOK] Cached %s for location %s", bird.CommonName, locationKey)
	} else {
		log.Printf("[WEBHOOK] Not caching fallback bird (will allow proper update later)")
	}

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