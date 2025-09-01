package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	
	"github.com/gin-gonic/gin"
)

// TestYotoUploadURL tests getting an upload URL from Yoto API
func (h *Handler) TestYotoUploadURL(c *gin.Context) {
	// Test 1: Try to get upload URL
	url := fmt.Sprintf("%s/media/transcode/audio/uploadUrl", h.config.YotoAPIBaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create request", "details": err.Error()})
		return
	}
	
	// Get the access token that's configured
	accessToken := h.config.YotoAccessToken
	envToken := os.Getenv("YOTO_ACCESS_TOKEN")
	
	if accessToken == "" && envToken == "" {
		c.JSON(500, gin.H{"error": "No access token configured"})
		return
	}
	
	// Use the token from config, or fall back to env
	if accessToken == "" {
		accessToken = envToken
	}
	
	// Show first/last few chars of token for debugging
	tokenPreview := ""
	if len(accessToken) > 20 {
		tokenPreview = accessToken[:10] + "..." + accessToken[len(accessToken)-10:]
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to make request",
			"details": err.Error(),
			"url": url,
			"token_preview": tokenPreview,
		})
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	c.JSON(resp.StatusCode, gin.H{
		"test": "get_upload_url",
		"url": url,
		"status_code": resp.StatusCode,
		"response_body": string(body),
		"token_preview": tokenPreview,
		"card_id": h.config.YotoCardID,
		"headers_sent": map[string]string{
			"Authorization": "Bearer " + tokenPreview,
			"Accept": "application/json",
		},
	})
}

// TestYotoCard tests getting card information
func (h *Handler) TestYotoCard(c *gin.Context) {
	cardID := h.config.YotoCardID
	if cardID == "" {
		c.JSON(400, gin.H{"error": "No card ID configured"})
		return
	}
	
	url := fmt.Sprintf("%s/content/%s", h.config.YotoAPIBaseURL, cardID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create request", "details": err.Error()})
		return
	}
	
	accessToken := h.config.YotoAccessToken
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to make request", "details": err.Error()})
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	c.JSON(resp.StatusCode, gin.H{
		"test": "get_card",
		"url": url,
		"card_id": cardID,
		"status_code": resp.StatusCode,
		"response_body": string(body),
	})
}