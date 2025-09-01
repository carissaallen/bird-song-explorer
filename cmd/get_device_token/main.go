package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	deviceCodeURL = "https://login.yotoplay.com/oauth/device/code"
	tokenURL      = "https://login.yotoplay.com/oauth/token"
	scope         = "profile offline_access"
	audience      = "https://api.yotoplay.com"
)

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type ErrorResponse struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.ErrorDescription)
}

func main() {
	clientID := os.Getenv("YOTO_CLIENT_ID")
	if clientID == "" {
		log.Fatal("YOTO_CLIENT_ID environment variable is required")
	}

	fmt.Println("🔐 Yoto Device Authorization Flow")
	fmt.Println("==================================")
	fmt.Printf("Client ID: %s\n\n", clientID)

	// Step 1: Initialize device authorization
	deviceResp, err := initializeDeviceAuth(clientID)
	if err != nil {
		log.Fatalf("Failed to initialize device auth: %v", err)
	}

	fmt.Println("📱 Please visit this URL to authorize:")
	fmt.Printf("   %s\n\n", deviceResp.VerificationURI)
	fmt.Println("Enter this code:")
	fmt.Printf("   %s\n\n", deviceResp.UserCode)
	fmt.Println("Or visit this URL directly:")
	fmt.Printf("   %s\n\n", deviceResp.VerificationURIComplete)
	fmt.Printf("Code expires in %d seconds\n\n", deviceResp.ExpiresIn)

	// Step 2: Poll for token
	fmt.Println("⏳ Waiting for authorization...")
	tokens, err := pollForToken(clientID, deviceResp.DeviceCode, deviceResp.Interval)
	if err != nil {
		log.Fatalf("Failed to get tokens: %v", err)
	}

	// Display tokens
	fmt.Println("\n✅ Successfully obtained tokens!")
	fmt.Println("==================================")
	fmt.Printf("\nAccess Token (first 50 chars):\n%s...\n", tokens.AccessToken[:min(50, len(tokens.AccessToken))])
	fmt.Printf("\nRefresh Token (first 50 chars):\n%s...\n", tokens.RefreshToken[:min(50, len(tokens.RefreshToken))])
	fmt.Printf("\nExpires in: %d seconds\n", tokens.ExpiresIn)

	// Generate Cloud Run commands
	generateCloudRunCommands(tokens)
}

func initializeDeviceAuth(clientID string) (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", scope)
	data.Set("audience", audience)

	resp, err := http.PostForm(deviceCodeURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization failed (status %d): %s", resp.StatusCode, string(body))
	}

	var deviceResp DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceResp); err != nil {
		return nil, err
	}

	return &deviceResp, nil
}

func pollForToken(clientID, deviceCode string, interval int) (*TokenResponse, error) {
	if interval < 5 {
		interval = 5 // minimum 5 seconds
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tokens, err := tryTokenExchange(clientID, deviceCode)
			if err == nil {
				return tokens, nil
			}

			// Check if it's a retryable error
			if errResp, ok := err.(*ErrorResponse); ok {
				switch errResp.ErrorCode {
				case "authorization_pending":
					fmt.Print(".")
					continue
				case "slow_down":
					// Increase polling interval
					ticker.Stop()
					interval += 5
					ticker = time.NewTicker(time.Duration(interval) * time.Second)
					fmt.Printf("\n(Slowing down, new interval: %d seconds)\n", interval)
					continue
				case "expired_token":
					return nil, fmt.Errorf("device code expired, please restart")
				default:
					return nil, fmt.Errorf("token exchange failed: %s", errResp.ErrorDescription)
				}
			}

			return nil, err
		}
	}
}

func tryTokenExchange(clientID, deviceCode string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)
	data.Set("client_id", clientID)
	data.Set("audience", audience)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		var tokenResp TokenResponse
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			return nil, err
		}
		return &tokenResp, nil
	}

	// Parse error response
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil, fmt.Errorf("unexpected error response: %s", string(body))
	}
	return nil, &errResp
}

func generateCloudRunCommands(tokenResp *TokenResponse) {
	fmt.Println("\n🚀 Cloud Run Update Commands")
	fmt.Println("==================================")
	fmt.Printf(`
# Update Cloud Run with tokens:
gcloud run services update bird-song-explorer \
  --region us-central1 \
  --update-env-vars \
    "YOTO_ACCESS_TOKEN=%s,\
YOTO_REFRESH_TOKEN=%s" \
  --quiet

`, tokenResp.AccessToken, tokenResp.RefreshToken)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}