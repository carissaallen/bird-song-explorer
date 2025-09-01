package yoto

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultAuthURL = "https://login.yotoplay.com/oauth/token"
)

type Client struct {
	clientID     string
	clientSecret string
	baseURL      string
	authURL      string
	httpClient   *http.Client
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type Card struct {
	CardID            string                 `json:"cardId"`
	UserID            string                 `json:"userId"`
	CreatedByClientID string                 `json:"createdByClientId,omitempty"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description,omitempty"`
	Content           map[string]interface{} `json:"content,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
}

type Track struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title"`
	Duration int    `json:"duration"`
	URL      string `json:"url"`
	IconURL  string `json:"iconUrl,omitempty"`
	Order    int    `json:"order"`
}

type UpdateCardRequest struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	ImageURL    string  `json:"imageUrl,omitempty"`
	Tracks      []Track `json:"tracks,omitempty"`
}

func NewClient(clientID, clientSecret, baseURL string) *Client {
	// Create HTTP client that doesn't follow redirects automatically
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Log redirects but don't follow them automatically
			fmt.Printf("REDIRECT: From %s to %s\n", via[len(via)-1].URL, req.URL)
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			// Only follow redirects to trusted domains
			if req.URL.Host != "api.yotoplay.com" && req.URL.Host != "login.yotoplay.com" {
				return fmt.Errorf("refusing to redirect to untrusted host: %s", req.URL.Host)
			}
			return nil
		},
	}
	
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      baseURL,
		authURL:      defaultAuthURL,
		httpClient:   httpClient,
	}
}

// SetTokens allows setting pre-obtained tokens (e.g., from OAuth flow)
func (c *Client) SetTokens(accessToken, refreshToken string, expiresIn int) {
	c.accessToken = accessToken
	c.refreshToken = refreshToken
	c.tokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
}

// extractTokenExpiry extracts the expiry time from a JWT token
func extractTokenExpiry(token string) int64 {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0
	}

	if exp, ok := claims["exp"].(float64); ok {
		return int64(exp)
	}

	return 0
}

func (c *Client) authenticate() error {
	if accessToken := os.Getenv("YOTO_ACCESS_TOKEN"); accessToken != "" && c.accessToken == "" {
		c.accessToken = accessToken
		if refreshToken := os.Getenv("YOTO_REFRESH_TOKEN"); refreshToken != "" {
			c.refreshToken = refreshToken
		}
		if exp := extractTokenExpiry(accessToken); exp > 0 {
			c.tokenExpiry = time.Unix(exp, 0)
		} else {
			c.tokenExpiry = time.Now().Add(24 * time.Hour)
		}
	}

	if c.refreshToken != "" && time.Now().After(c.tokenExpiry) {
		return c.refreshAccessToken()
	}

	if c.accessToken != "" {
		return nil
	}

	return fmt.Errorf("no authentication method available - set YOTO_ACCESS_TOKEN and YOTO_REFRESH_TOKEN environment variables")
}

func (c *Client) refreshAccessToken() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.clientID)
	data.Set("refresh_token", c.refreshToken)

	req, err := http.NewRequest("POST", c.authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed: %d - %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		c.refreshToken = tokenResp.RefreshToken
	}
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

func (c *Client) ensureAuthenticated() error {
	if c.accessToken == "" {
		return c.authenticate()
	}

	if time.Now().After(c.tokenExpiry.Add(-5 * time.Minute)) {
		if c.refreshToken != "" {
			if err := c.refreshAccessToken(); err != nil {
				if c.accessToken == "" {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Client) GetCard(cardID string) (*Card, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	// Use the /content/{contentId} endpoint to get card content
	url := fmt.Sprintf("%s/content/%s", c.baseURL, cardID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get card: %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		Card Card `json:"card"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Card, nil
}

func (c *Client) UpdateCard(cardID string, update UpdateCardRequest) (*Card, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/content/%s", c.baseURL, cardID)

	jsonBody, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update card: %d - %s", resp.StatusCode, string(body))
	}

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}

	return &card, nil
}

func (c *Client) SearchLibrary(query string) ([]LibraryItem, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/library/search?q=%s", c.baseURL, query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("library search failed: %d - %s", resp.StatusCode, string(body))
	}

	var items []LibraryItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	return items, nil
}

type LibraryItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	IconURL  string `json:"iconUrl,omitempty"`
	AudioURL string `json:"audioUrl,omitempty"`
}

type DeviceConfig struct {
	Device struct {
		Config struct {
			GeoTimezone             string `json:"geoTimezone"`
			DayDisplayBrightness    string `json:"dayDisplayBrightness,omitempty"`
			NightDisplayBrightness  string `json:"nightDisplayBrightness,omitempty"`
			VolumeLevel             string `json:"volumeLevel,omitempty"`
			MaxVolumeLimit          string `json:"maxVolumeLimit,omitempty"`
			AmbientColour           string `json:"ambientColour,omitempty"`
			ClockFace               string `json:"clockFace,omitempty"`
			HourFormat              string `json:"hourFormat,omitempty"`
			HeadphonesVolumeLimited bool   `json:"headphonesVolumeLimited,omitempty"`
			BluetoothEnabled        string `json:"bluetoothEnabled,omitempty"`
			BtHeadphonesEnabled     bool   `json:"btHeadphonesEnabled,omitempty"`
		} `json:"config"`
		DeviceID         string `json:"deviceId"`
		DeviceType       string `json:"deviceType,omitempty"`
		DeviceFamily     string `json:"deviceFamily,omitempty"`
		Online           bool   `json:"online"`
		ReleaseChannelId string `json:"releaseChannelId,omitempty"`
		Mac              string `json:"mac,omitempty"`
		RegistrationCode string `json:"registrationCode,omitempty"`
	} `json:"device"`
}

func (c *Client) GetDeviceConfig(deviceID string) (*DeviceConfig, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/device-v2/%s/config", c.baseURL, deviceID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get device config: %d - %s", resp.StatusCode, string(body))
	}

	var config DeviceConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
