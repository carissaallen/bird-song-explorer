package api

import (
	"log"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/internal/services"
	"github.com/callen/bird-song-explorer/pkg/yoto"
)

type Handler struct {
	config                  *config.Config
	locationService         *services.LocationService
	timezoneLocationService *services.TimezoneLocationService
	timezoneLookup          *services.TimezoneLookupService
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
		yotoClient:              yotoClient,
		updateCache:             services.NewUpdateCache(),
		availableBirds:          services.NewAvailableBirdsService(),
	}
}
