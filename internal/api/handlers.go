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
	updateCache             *services.UpdateCache
	availableBirds          *services.AvailableBirdsService
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
		updateCache:             services.NewUpdateCache(),
		availableBirds:          services.NewAvailableBirdsService(),
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
