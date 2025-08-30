package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// HandleTokenRefresh manually triggers a token refresh for testing
func (h *Handler) HandleTokenRefresh(c *gin.Context) {
	refreshToken := os.Getenv("YOTO_REFRESH_TOKEN")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No refresh token available"})
		return
	}

	tokens, err := h.refreshTokens(refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh tokens"})
		return
	}

	h.yotoClient.SetTokens(tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn)

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Tokens refreshed successfully",
		"expires_in": tokens.ExpiresIn,
	})
}

func (h *Handler) refreshTokens(refreshToken string) (*OAuthTokenResponse, error) {
	tokenURL := "https://login.yotoplay.com/oauth/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", h.config.YotoClientID)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.RefreshToken == "" {
		tokenResp.RefreshToken = refreshToken
	}

	return &tokenResp, nil
}

// HandleOAuthCallback handles the OAuth callback from Yoto
func (h *Handler) HandleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing authorization code",
		})
		return
	}

	// For now, just return the code to the user
	// In production, this would exchange the code for tokens
	c.HTML(http.StatusOK, "", `
		<!DOCTYPE html>
		<html>
		<head>
			<title>OAuth Success</title>
			<style>
				body {
					font-family: system-ui, -apple-system, sans-serif;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
					margin: 0;
					background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
				}
				.container {
					background: white;
					padding: 2rem;
					border-radius: 8px;
					box-shadow: 0 4px 6px rgba(0,0,0,0.1);
					text-align: center;
					max-width: 500px;
				}
				h1 { color: #2d3748; }
				.code {
					background: #f7fafc;
					padding: 1rem;
					border-radius: 4px;
					margin: 1rem 0;
					font-family: monospace;
					word-break: break-all;
				}
				.success { color: #48bb78; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1 class="success">✅ Authorization Successful!</h1>
				<p>You've successfully authorized the Bird Song Explorer app.</p>
				<div class="code">
					<strong>Authorization Code:</strong><br>
					`+code+`
				</div>
				<p><small>State: `+state+`</small></p>
				<p>You can close this window and return to your terminal.</p>
			</div>
		</body>
		</html>
	`)
}
