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

	v1 := router.Group("/api/v1")
	{
		v1.POST("/daily-update", handler.DailyUpdateHandler) // Scheduler trigger for global bird
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
