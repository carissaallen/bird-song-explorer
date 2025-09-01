package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// HandleDebugOAuth helps debug OAuth flow issues
func (h *Handler) HandleDebugOAuth(c *gin.Context) {
	// Log all query parameters
	queryParams := c.Request.URL.Query()
	
	// Check for OAuth callback parameters
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	errorDesc := c.Query("error_description")
	
	// Build authorization URL for testing
	authURL := "https://api.yotoplay.com/authorize"
	redirectURI := "https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"
	
	authFullURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&state=debug_test&scope=offline_access",
		authURL, h.config.YotoClientID, url.QueryEscape(redirectURI))
	
	// Return debug information
	c.JSON(http.StatusOK, gin.H{
		"debug_info": map[string]interface{}{
			"request_method": c.Request.Method,
			"request_url":    c.Request.URL.String(),
			"query_params":   queryParams,
			"has_code":       code != "",
			"code":           code,
			"state":          state,
			"error":          errorParam,
			"error_desc":     errorDesc,
			"client_id":      h.config.YotoClientID,
			"redirect_uri":   redirectURI,
		},
		"auth_url": authFullURL,
		"instructions": []string{
			"1. Visit the auth_url above to start OAuth flow",
			"2. After authorization, check what parameters you receive",
			"3. If no 'code' parameter, check Yoto Developer Dashboard settings",
			"4. Ensure redirect URI matches exactly: " + redirectURI,
		},
	})
}

// HandleWebhookDebug handles both webhooks and OAuth with detailed logging
func (h *Handler) HandleWebhookDebug(c *gin.Context) {
	// Log the request details
	fmt.Printf("=== Webhook/OAuth Debug ===\n")
	fmt.Printf("Method: %s\n", c.Request.Method)
	fmt.Printf("URL: %s\n", c.Request.URL.String())
	fmt.Printf("Query Params: %v\n", c.Request.URL.Query())
	fmt.Printf("Headers: %v\n", c.Request.Header)
	
	// Check if this is an OAuth callback
	if code := c.Query("code"); code != "" {
		fmt.Printf("OAuth callback detected with code: %s\n", code)
		h.HandleOAuthCallback(c)
		return
	}
	
	// Check for OAuth error
	if errorParam := c.Query("error"); errorParam != "" {
		errorDesc := c.Query("error_description")
		fmt.Printf("OAuth error: %s - %s\n", errorParam, errorDesc)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errorParam,
			"error_description": errorDesc,
			"message": "OAuth authorization failed",
		})
		return
	}
	
	// Otherwise, handle as webhook
	h.HandleYotoWebhook(c)
}