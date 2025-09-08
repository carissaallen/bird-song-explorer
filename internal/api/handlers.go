package api

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
	"github.com/callen/bird-song-explorer/pkg/yoto"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	config                  *config.Config
	locationService         *services.LocationService
	timezoneLocationService *services.TimezoneLocationService
	timezoneLookup          *services.TimezoneLookupService
	birdSelector            *services.BirdSelector
	yotoClient              *yoto.Client
	audioManager            *services.AudioManager
	narrationManager        *services.NarrationManager
	introGenerator          *services.DynamicIntroGenerator
	updateCache            *services.UpdateCache
}

func NewHandler(cfg *config.Config) *Handler {
	yotoClient := yoto.NewClient(
		cfg.YotoClientID,
		"", // No client secret needed for public client
		cfg.YotoAPIBaseURL,
	)

	// Set the access and refresh tokens if available
	if cfg.YotoAccessToken != "" && cfg.YotoRefreshToken != "" {
		// The expiresIn is not stored, so we'll use a default of 24 hours
		// The client will check token expiry and refresh as needed
		yotoClient.SetTokens(cfg.YotoAccessToken, cfg.YotoRefreshToken, 86400)
	}

	// Initialize timezone lookup service
	timezoneLookup, err := services.NewTimezoneLookupService()
	if err != nil {
		log.Printf("Failed to initialize timezone lookup service: %v, will use fallback", err)
	}

	return &Handler{
		config:                  cfg,
		locationService:         services.NewLocationService(),
		timezoneLocationService: services.NewTimezoneLocationService(),
		timezoneLookup:          timezoneLookup,
		birdSelector:            services.NewBirdSelector(cfg.EBirdAPIKey, cfg.XenoCantoAPIKey),
		yotoClient:              yotoClient,
		audioManager:            services.NewAudioManager(),
		narrationManager:        services.NewNarrationManager(cfg.ElevenLabsAPIKey),
		introGenerator:          services.NewDynamicIntroGenerator(cfg.ElevenLabsAPIKey),
		updateCache:            services.NewUpdateCache(),
	}
}

func (h *Handler) GetBirdOfDay(c *gin.Context) {
	clientIP := c.ClientIP()
	if clientIP == "::1" {
		clientIP = c.GetHeader("X-Forwarded-For")
		if clientIP != "" {
			clientIP = strings.Split(clientIP, ",")[0]
		}
	}

	location, err := h.locationService.GetLocationFromIP(clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to determine location",
		})
		return
	}

	bird, err := h.birdSelector.SelectBirdOfDay(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to select bird",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bird":     bird,
		"location": location,
	})
}

// DEPRECATED: Old webhook handler - use HandleYotoWebhookV3 instead
// func (h *Handler) HandleYotoWebhook(c *gin.Context) {
// 	// Check if this is an OAuth callback (has 'code' query parameter)
// 	if code := c.Query("code"); code != "" {
// 		h.HandleOAuthCallback(c)
// 		return
// 	}
//
// 	var webhook struct {
// 		EventType string `json:"eventType"`
// 		CardID    string `json:"cardId"`
// 		DeviceID  string `json:"deviceId"`
// 		UserID    string `json:"userId"`
// 	}
//
// 	if err := c.ShouldBindJSON(&webhook); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
// 		return
// 	}
//
// 	if webhook.EventType != "card.played" {
// 		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
// 		return
// 	}
//
// 	cardID := webhook.CardID
// 	if cardID == "" {
// 		cardID = "ipHAS"
// 	}
//
// 	// First, try to get location from IP address (most accurate)
// 	clientIP := c.ClientIP()
// 	location, err := h.locationService.GetLocationFromIP(clientIP)
// 	var deviceTimezone string
//
// 	if err == nil && location.City != "London" {
// 		log.Printf("Using IP-based location: %s, %s (IP: %s)\n",
// 			location.City, location.Country, clientIP)
// 	} else {
// 		// If IP location failed or returned default, try device timezone as fallback
// 		if webhook.DeviceID != "" {
// 			deviceConfig, err := h.yotoClient.GetDeviceConfig(webhook.DeviceID)
// 			if err == nil && deviceConfig != nil {
// 				deviceTimezone = deviceConfig.Device.Config.GeoTimezone
// 			}
// 		}
//
// 		// If we have a device timezone, use it
// 		if deviceTimezone != "" {
// 			tzLocation := h.timezoneLocationService.GetLocationFromTimezone(deviceTimezone)
// 			if tzLocation.City != "London" || deviceTimezone == "Europe/London" {
// 				location = tzLocation
// 				log.Printf("Using timezone-based location: %s, %s (from timezone: %s)\n",
// 					location.City, location.Country, deviceTimezone)
// 			}
// 		}
//
// 		// If still using default location, log a warning
// 		if location.City == "London" && deviceTimezone != "Europe/London" {
// 			log.Printf("WARNING: Using default location (London, UK) - IP: %s, Timezone: %s\n",
// 				clientIP, deviceTimezone)
// 		}
// 	}
//
// 	bird, err := h.birdSelector.SelectBirdOfDay(location)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to select bird"})
// 		return
// 	}
//
// 	// Select voice for this session (consistent throughout)
// 	voice := h.narrationManager.SelectDailyVoice()
//
// 	baseURL := "https://yoto-bird-song-explorer-362662614716.us-central1.run.app"
// 	introFile := h.getIntroFileForVoice(voice.Name)
// 	introURL := fmt.Sprintf("%s/audio/intros/%s", baseURL, introFile)
//
// 	contentManager := yoto.NewContentManager(h.yotoClient)
//
// 	err = contentManager.UpdateExistingCardContentWithDescriptionAndVoice(
// 		cardID,
// 		bird.CommonName,
// 		introURL,
// 		bird.AudioURL,
// 		"", // No description for test
// 		voice.VoiceID,
// 	)
//
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to update card",
// 			"details": err.Error(),
// 		})
// 		return
// 	}
//
// 	c.JSON(http.StatusOK, gin.H{
// 		"status":         "updated",
// 		"bird":           bird.CommonName,
// 		"narrator":       voice.Name,
// 		"location":       location.City,
// 		"cardId":         cardID,
// 		"deviceTimezone": deviceTimezone,
// 	})
// }

func (h *Handler) GetRandomIntro(c *gin.Context) {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)

	c.JSON(http.StatusOK, gin.H{
		"intro_url": introURL,
		"voice_id":  voiceID,
		"message":   "Random intro selected",
	})
}

func (h *Handler) createTracksFromBird(bird *models.Bird) []yoto.Track {
	baseURL := "https://yoto-bird-song-explorer-362662614716.us-central1.run.app"

	voice := h.narrationManager.GetSelectedVoiceName()

	introFile := h.getIntroFileForVoice(voice)

	tracks := []yoto.Track{
		{
			Title:    "Welcome to Bird Song Explorer",
			URL:      fmt.Sprintf("%s/audio/intros/%s", baseURL, introFile),
			Duration: 10,
			Order:    1,
		},
		{
			Title:    bird.CommonName + " Song",
			URL:      bird.AudioURL,
			Duration: 30,
			Order:    2,
		},
	}

	for i := range bird.Facts {
		tracks = append(tracks, yoto.Track{
			Title:    fmt.Sprintf("Fact %d - %s", i+1, voice),
			URL:      "", // Would be TTS URL with same voice
			Duration: 15,
			Order:    i + 3,
		})
	}

	return tracks
}

// DEPRECATED: Manual update handler - use test-webhook endpoint instead
// func (h *Handler) UpdateCardManually(c *gin.Context) {
// 	cardID := c.Param("cardId")
// 	if cardID == "" {
// 		cardID = "ipHAS" // Default to your card
// 	}
//
// 	// Get location from query params or use default
// 	location := &models.Location{
// 		Latitude:  51.5074,
// 		Longitude: -0.1278,
// 		City:      "London",
// 		Region:    "England",
// 		Country:   "United Kingdom",
// 	}
//
// 	bird, err := h.birdSelector.SelectBirdOfDay(location)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to select bird",
// 			"details": err.Error(),
// 		})
// 		return
// 	}
//
// 	voice := h.narrationManager.SelectDailyVoice()
//
// 	baseURL := "https://yoto-bird-song-explorer-362662614716.us-central1.run.app"
// 	introFile := h.getIntroFileForVoice(voice.Name)
// 	introURL := fmt.Sprintf("%s/audio/intros/%s", baseURL, introFile)
//
// 	contentManager := yoto.NewContentManager(h.yotoClient)
//
// 	_, voiceID := h.audioManager.GetRandomIntroURL(baseURL)
//
// 	err = contentManager.UpdateExistingCardContentWithDescriptionAndVoice(
// 		cardID,
// 		bird.CommonName,
// 		introURL,
// 		bird.AudioURL,
// 		"", // No description
// 		voiceID,
// 	)
//
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to update card",
// 			"details": err.Error(),
// 		})
// 		return
// 	}
//
// 	c.JSON(http.StatusOK, gin.H{
// 		"status":   "Card updated successfully!",
// 		"cardId":   cardID,
// 		"bird":     bird.CommonName,
// 		"narrator": voice.Name,
// 		"location": location.City,
// 		"intro":    introURL,
// 		"song":     bird.AudioURL,
// 	})
// }

func (h *Handler) getIntroFileForVoice(voiceName string) string {
	// Map voice to available intro files
	rand.Seed(time.Now().UnixNano())

	// Build intro list for the specific voice
	intros := []string{}
	for i := 0; i < 8; i++ {
		intros = append(intros, fmt.Sprintf("intro_%02d_%s.mp3", i, voiceName))
	}

	// If no intros exist for this voice, fall back to Antoni
	if voiceName != "Amelia" && voiceName != "Antoni" && voiceName != "Charlotte" &&
		voiceName != "Peter" && voiceName != "Drake" && voiceName != "Sally" {
		// Use Antoni as fallback
		intros = []string{}
		for i := 0; i < 8; i++ {
			intros = append(intros, fmt.Sprintf("intro_%02d_Antoni.mp3", i))
		}
	}

	return intros[rand.Intn(len(intros))]
}

// ServeIntroWithNatureSounds serves intro files mixed with nature sounds
func (h *Handler) ServeIntroWithNatureSounds(c *gin.Context) {
	filename := c.Param("filename")

	// Log the request
	log.Printf("[INTRO_HANDLER] Serving intro with nature sounds: %s", filename)

	// Read the original intro file
	introPath := fmt.Sprintf("./assets/final_intros/%s", filename)
	introData, err := os.ReadFile(introPath)
	if err != nil {
		log.Printf("[INTRO_HANDLER] Failed to read intro file %s: %v", filename, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Intro file not found"})
		return
	}

	// Create intro mixer
	introMixer := services.NewIntroMixer()

	// Get user timezone from header or use default
	userTimezone := c.GetHeader("X-User-Timezone")

	// Mix intro with nature sounds
	log.Printf("[INTRO_HANDLER] Mixing intro with nature sounds for timezone: %s", userTimezone)
	mixedData, err := introMixer.MixIntroWithNatureSoundsForUser(introData, "", userTimezone)
	if err != nil {
		log.Printf("[INTRO_HANDLER] Failed to mix intro with nature sounds: %v", err)
		// Fall back to serving the original intro
		c.Data(http.StatusOK, "audio/mpeg", introData)
		return
	}

	log.Printf("[INTRO_HANDLER] Successfully mixed intro (original: %d bytes, mixed: %d bytes)",
		len(introData), len(mixedData))

	// Set appropriate headers for audio streaming
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Serve the mixed audio
	c.Data(http.StatusOK, "audio/mpeg", mixedData)
}
