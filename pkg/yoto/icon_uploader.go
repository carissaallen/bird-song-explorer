package yoto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type IconUploader struct {
	client    *Client
	iconCache map[string]string
	cacheMu   sync.RWMutex
}

type IconUploadResponse struct {
	DisplayIcon struct {
		MediaID       string      `json:"mediaId"`
		UserID        string      `json:"userId"`
		DisplayIconID string      `json:"displayIconId"`
		URL           interface{} `json:"url"` // Can be string or empty object
		New           bool        `json:"new"`
	} `json:"displayIcon"`
}

// Cached icon IDs to avoid re-uploading
var (
	cachedBinocularsID string
	binocularsOnce     sync.Once
	cachedMeadowlarkID string
	meadowlarkOnce     sync.Once
	cachedHikingBootID string
	hikingBootOnce     sync.Once
)

func NewIconUploader(client *Client) *IconUploader {
	return &IconUploader{
		client:    client,
		iconCache: make(map[string]string),
	}
}

// UploadIcon uploads an icon file and returns the media ID
func (iu *IconUploader) UploadIcon(filePath string, filename string) (string, error) {
	// For animated GIFs, use the special uploader
	if strings.HasSuffix(strings.ToLower(filePath), ".gif") {
		return iu.UploadAnimatedGIF(filePath, filename)
	}

	// Check cache first
	iu.cacheMu.RLock()
	if cachedID, exists := iu.iconCache[filePath]; exists {
		iu.cacheMu.RUnlock()
		return cachedID, nil
	}
	iu.cacheMu.RUnlock()

	if err := iu.client.ensureAuthenticated(); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Read the icon file
	iconData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read icon file: %w", err)
	}

	// Determine content type based on file extension
	contentType := "image/png"
	if strings.HasSuffix(strings.ToLower(filePath), ".jpg") || strings.HasSuffix(strings.ToLower(filePath), ".jpeg") {
		contentType = "image/jpeg"
	}

	// Build URL with query parameters - use autoConvert=true for static images
	url := fmt.Sprintf("%s/media/displayIcons/user/me/upload?autoConvert=true&filename=%s",
		iu.client.baseURL, filename)

	// Create request with raw image data (as shown in Yoto docs)
	req, err := http.NewRequest("POST", url, bytes.NewReader(iconData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+iu.client.accessToken)
	req.Header.Set("Content-Type", contentType)

	// Send request
	resp, err := iu.client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload icon: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("upload failed: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var uploadResp IconUploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w - body: %s", err, string(body))
	}

	if uploadResp.DisplayIcon.MediaID == "" {
		return "", fmt.Errorf("no media ID in response: %s", string(body))
	}

	// Cache the result
	iu.cacheMu.Lock()
	iu.iconCache[filePath] = uploadResp.DisplayIcon.MediaID
	iu.cacheMu.Unlock()

	return uploadResp.DisplayIcon.MediaID, nil
}

// UploadIconNoCache uploads an icon WITHOUT caching - useful for bird icons that change
func (iu *IconUploader) UploadIconNoCache(filePath string, filename string) (string, error) {
	// For animated GIFs, use the special uploader
	if strings.HasSuffix(strings.ToLower(filePath), ".gif") {
		return iu.UploadAnimatedGIF(filePath, filename)
	}

	if err := iu.client.ensureAuthenticated(); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Read the icon file
	iconData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read icon file: %w", err)
	}

	// Determine content type based on file extension
	contentType := "image/png"
	if strings.HasSuffix(strings.ToLower(filePath), ".jpg") || strings.HasSuffix(strings.ToLower(filePath), ".jpeg") {
		contentType = "image/jpeg"
	}

	// Build URL with query parameters - use autoConvert=true for static images
	url := fmt.Sprintf("%s/media/displayIcons/user/me/upload?autoConvert=true&filename=%s",
		iu.client.baseURL, filename)

	// Create request with raw image data (as shown in Yoto docs)
	req, err := http.NewRequest("POST", url, bytes.NewReader(iconData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+iu.client.accessToken)
	req.Header.Set("Content-Type", contentType)

	// Send request
	resp, err := iu.client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload icon: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("upload failed: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var uploadResp IconUploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w - body: %s", err, string(body))
	}

	if uploadResp.DisplayIcon.MediaID == "" {
		return "", fmt.Errorf("no media ID in response: %s", string(body))
	}

	// DO NOT cache - return the media ID directly
	fmt.Printf("[ICON_UPLOADER] Uploaded icon (no cache) %s: %s\n", filename, uploadResp.DisplayIcon.MediaID)
	return uploadResp.DisplayIcon.MediaID, nil
}

// GetBinocularsIcon uploads the binoculars icon once and returns its ID
func (iu *IconUploader) GetBinocularsIcon() (string, error) {
	var uploadErr error

	binocularsOnce.Do(func() {
		// Try different possible locations for the binoculars icon
		possiblePaths := []string{
			"./assets/icons/binoculars.png",
			"assets/icons/binoculars.png",
			"/root/assets/icons/binoculars.png", // Docker working directory
			"./binoculars.png",
			"binoculars.png",
		}

		var iconPath string
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				iconPath = path
				break
			} else {
			}
		}

		if iconPath == "" {
			uploadErr = fmt.Errorf("binoculars.png not found in any expected location")
			return
		}

		mediaID, err := iu.UploadIcon(iconPath, "binoculars")
		if err != nil {
			uploadErr = err
			return
		}

		cachedBinocularsID = mediaID
	})

	if uploadErr != nil {
		return "", uploadErr
	}

	if cachedBinocularsID == "" {
		return "", fmt.Errorf("failed to get binoculars icon ID")
	}

	return cachedBinocularsID, nil
}

// GetMeadowlarkIcon uploads the meadowlark icon once and returns its ID
func (iu *IconUploader) GetMeadowlarkIcon() (string, error) {
	// Check if already cached
	if cachedMeadowlarkID != "" {
		return cachedMeadowlarkID, nil
	}

	var uploadErr error

	meadowlarkOnce.Do(func() {
		// Try different possible locations for the meadowlark icon
		possiblePaths := []string{
			"./assets/icons/meadowlark_fly.gif",
			"assets/icons/meadowlark_fly.gif",
			"/root/assets/icons/meadowlark_fly.gif", // Docker working directory
			"./meadowlark_fly.gif",
			"meadowlark_fly.gif",
		}

		var iconPath string
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				iconPath = path
				break
			}
		}

		if iconPath == "" {
			uploadErr = fmt.Errorf("meadowlark_fly.gif not found in any expected location")
			return
		}

		mediaID, err := iu.UploadIcon(iconPath, "meadowlark")
		if err != nil {
			uploadErr = err
			return
		}

		cachedMeadowlarkID = mediaID
	})

	if uploadErr != nil {
		return "", uploadErr
	}

	if cachedMeadowlarkID == "" {
		return "", fmt.Errorf("failed to get meadowlark icon ID - cache is empty")
	}

	return cachedMeadowlarkID, nil
}

// GetHikingBootIcon uploads the hiking boot icon once and returns its ID
func (iu *IconUploader) GetHikingBootIcon() (string, error) {
	// Check if already cached
	if cachedHikingBootID != "" {
		return cachedHikingBootID, nil
	}

	var uploadErr error

	hikingBootOnce.Do(func() {
		// Try different possible locations for the hiking boot icon
		possiblePaths := []string{
			"./assets/icons/hiking-boot.png",
			"assets/icons/hiking-boot.png",
			"/root/assets/icons/hiking-boot.png", // Docker working directory
			"./hiking-boot.png",
			"hiking-boot.png",
		}

		var iconPath string
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				iconPath = path
				break
			}
		}

		if iconPath == "" {
			uploadErr = fmt.Errorf("hiking-boot.png not found in any expected location")
			return
		}

		mediaID, err := iu.UploadIcon(iconPath, "hiking_boot")
		if err != nil {
			uploadErr = err
			return
		}

		cachedHikingBootID = mediaID
	})

	if uploadErr != nil {
		return "", uploadErr
	}

	if cachedHikingBootID == "" {
		return "", fmt.Errorf("failed to get hiking boot icon ID - cache is empty")
	}

	return cachedHikingBootID, nil
}

// FormatIconID formats a media ID for use in content
func FormatIconID(mediaID string) string {
	if mediaID == "" {
		return ""
	}
	if len(mediaID) > 6 && mediaID[:6] == "yoto:#" {
		return mediaID
	}
	return fmt.Sprintf("yoto:#%s", mediaID)
}

// UploadAnimatedGIF uploads an animated GIF icon with autoConvert=false
func (iu *IconUploader) UploadAnimatedGIF(filePath string, filename string) (string, error) {
	// Check cache first
	iu.cacheMu.RLock()
	if cachedID, exists := iu.iconCache[filePath]; exists {
		iu.cacheMu.RUnlock()
		return cachedID, nil
	}
	iu.cacheMu.RUnlock()

	if err := iu.client.ensureAuthenticated(); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Read the GIF file
	gifData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read GIF file: %w", err)
	}

	// Build URL with autoConvert=false for animated GIFs
	url := fmt.Sprintf("%s/media/displayIcons/user/me/upload?autoConvert=false&filename=%s",
		iu.client.baseURL, filename)

	// Create request with GIF data
	req, err := http.NewRequest("POST", url, bytes.NewReader(gifData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", iu.client.accessToken))
	req.Header.Set("Content-Type", "image/gif")

	// Make request
	resp, err := iu.client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload animated GIF: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var uploadResp struct {
		MediaID string `json:"mediaId"`
	}
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Cache the result
	iu.cacheMu.Lock()
	iu.iconCache[filePath] = uploadResp.MediaID
	iu.cacheMu.Unlock()

	fmt.Printf("[ICON_UPLOADER] Successfully uploaded animated GIF %s: %s\n", filename, uploadResp.MediaID)
	return uploadResp.MediaID, nil
}
