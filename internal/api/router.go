package api

import (
	"net/http"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	handler := NewHandler(cfg)

	router.GET("/health", healthCheck)

	router.GET("/audio/intros/:filename", handler.ServeIntroWithNatureSounds)
	router.Static("/audio/outros", "./assets/final_outros")
	router.Static("/audio/cache", "./audio_cache")   // This already serves everything under audio_cache including dynamic_intros
	router.Static("/assets/icons", "./assets/icons") // Serve icon files for dynamic icon URLs
	// Removed prerecorded route - now using GCS for all bird audio

	v1 := router.Group("/api/v1")
	{
		v1.GET("/bird-of-day", handler.GetBirdOfDay)
		v1.POST("/yoto/webhook", handler.HandleYotoWebhookStreaming) // Streaming: Returns playlist when USE_STREAMING=true
		// v1.POST("/webhook", handler.HandleYotoWebhookUnified)  // DEPRECATED: Use /api/v1/yoto/webhook
		v1.GET("/test-webhook", handler.TestWebhookHandler) // Test webhook simulation
		// v1.GET("/audio/intro", handler.GetRandomIntro) // Temporarily disabled - migrating to human voices
		// v1.POST("/update-card/:cardId", handler.SmartUpdateHandler)  // DEPRECATED: Use test-webhook
		v1.POST("/daily-update", handler.DailyUpdateHandler) // Scheduler trigger for global bird
		// v1.POST("/smart-update", handler.SmartUpdateHandler)  // DEPRECATED: Redundant
		v1.POST("/yoto/token/refresh", handler.HandleTokenRefresh)

		// Streaming endpoints for dynamic content
		v1.GET("/stream/intro", handler.StreamIntro)
		v1.GET("/stream/announcement", handler.StreamBirdAnnouncement)
		v1.GET("/stream/description", handler.StreamDescription)
		v1.GET("/stream/outro", handler.StreamOutro)
	}

	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "bird-song-explorer",
	})
}
