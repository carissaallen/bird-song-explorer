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

	// Serve static audio files
	router.Static("/audio/intros", "./final_intros")
	router.Static("/audio/cache", "./audio_cache") // This already serves everything under audio_cache including dynamic_intros

	v1 := router.Group("/api/v1")
	{
		v1.GET("/bird-of-day", handler.GetBirdOfDay)
		v1.POST("/yoto/webhook", handler.HandleYotoWebhook)
		v1.GET("/audio/intro", handler.GetRandomIntro)
		v1.POST("/update-card/:cardId", handler.UpdateCardManually)
		v1.POST("/daily-update", handler.DailyUpdateHandler)
	}

	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "bird-song-explorer",
	})
}
