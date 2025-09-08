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

// HandleYotoWebhookStreaming processes webhook events and returns streaming playlist
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
	useStreaming := os.Getenv("USE_STREAMING")
	if useStreaming != "true" {
		// Fall back to the traditional webhook handler
		h.HandleYotoWebhookV4(c)
		return
	}

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

	// Store session information for streaming tracks
	sessionKey := fmt.Sprintf("ip_%s", clientIP)
	session := &StreamingSession{
		SessionID:    sessionKey,
		Location:     userLocation,
		BirdName:     bird.CommonName,
		BirdAudioURL: bird.AudioURL,
		CreatedAt:    time.Now(),
	}
	
	// Get voice ID for consistency across tracks
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}
	_, voiceID := h.audioManager.GetRandomIntroURL(baseURL)
	session.VoiceID = voiceID
	
	// Store session
	sessionStore[sessionKey] = session
	log.Printf("[WEBHOOK-STREAMING] Created session %s with bird %s", sessionKey, bird.CommonName)
	
	// Build streaming playlist response
	// According to Yoto docs, we need to return a playlist with streaming tracks
	response := gin.H{
		"title": fmt.Sprintf("Bird Song Explorer - %s", bird.CommonName),
		"metadata": gin.H{
			"description": fmt.Sprintf("Streaming playlist for %s", bird.CommonName),
		},
		"content": gin.H{
			"chapters": []gin.H{
				{
					"key":   "01",
					"title": "Introduction",
					"overlayLabel": "1",
					"tracks": []gin.H{
						{
							"key":      "01",
							"title":    "Introduction",
							"trackUrl": fmt.Sprintf("%s/api/v1/stream/intro", baseURL),
							"type":     "stream",
							"format":   "mp3",
							"duration": 30,
						},
					},
					"display": gin.H{
						"icon16x16": "yoto:#radio",
					},
				},
				{
					"key":   "02", 
					"title": "Today's Bird",
					"overlayLabel": "2",
					"tracks": []gin.H{
						{
							"key":      "01",
							"title":    "Bird Announcement",
							"trackUrl": fmt.Sprintf("%s/api/v1/stream/announcement", baseURL),
							"type":     "stream",
							"format":   "mp3",
							"duration": 10,
						},
					},
					"display": gin.H{
						"icon16x16": "yoto:#binoculars",
					},
				},
				{
					"key":   "03",
					"title": fmt.Sprintf("%s Song", bird.CommonName),
					"overlayLabel": "3",
					"tracks": []gin.H{
						{
							"key":      "01",
							"title":    fmt.Sprintf("%s Song", bird.CommonName),
							"trackUrl": fmt.Sprintf("%s/api/v1/stream/bird-song", baseURL),
							"type":     "stream",
							"format":   "mp3",
							"duration": 30,
						},
					},
					"display": gin.H{
						"icon16x16": "yoto:#bird",
					},
				},
				{
					"key":   "04",
					"title": "Bird Description",
					"overlayLabel": "4",
					"tracks": []gin.H{
						{
							"key":      "01",
							"title":    "Bird Description",
							"trackUrl": fmt.Sprintf("%s/api/v1/stream/description", baseURL),
							"type":     "stream",
							"format":   "mp3",
							"duration": 60,
						},
					},
					"display": gin.H{
						"icon16x16": "yoto:#info",
					},
				},
				{
					"key":   "05",
					"title": "See You Tomorrow",
					"overlayLabel": "5",
					"tracks": []gin.H{
						{
							"key":      "01",
							"title":    "See You Tomorrow Explorers",
							"trackUrl": fmt.Sprintf("%s/api/v1/stream/outro", baseURL),
							"type":     "stream",
							"format":   "mp3",
							"duration": 20,
						},
					},
					"display": gin.H{
						"icon16x16": "yoto:#hiking",
					},
				},
			},
		},
	}
	
	log.Printf("[WEBHOOK-STREAMING] Returning streaming playlist for %s (location: %v)", 
		bird.CommonName, userLocation != nil)
	
	c.JSON(http.StatusOK, response)
}