package yoto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type StreamingChapter struct {
	Key          string           `json:"key"`
	Title        string           `json:"title"`
	OverlayLabel string           `json:"overlayLabel,omitempty"`
	Tracks       []StreamingTrack `json:"tracks"`
	Display      Display          `json:"display,omitempty"`
}

type StreamingTrack struct {
	Key          string  `json:"key"`
	Title        string  `json:"title,omitempty"`
	TrackURL     string  `json:"trackUrl"`
	Type         string  `json:"type"`
	Format       string  `json:"format"`
	Duration     int     `json:"duration"`
	OverlayLabel string  `json:"overlayLabel,omitempty"`
	Display      Display `json:"display,omitempty"`
}

type StreamingContent struct {
	Chapters []StreamingChapter `json:"chapters"`
}

var defaultIconID = "yoto:#RSsi4eQvVffIDMHbq3cuKn0ebSg0X-3Y-ZxrAorxycY"

func (cm *ContentManager) uploadTrackIcon(iconPath string, iconName string) string {
	if cm.iconUploader == nil {
		fmt.Printf("[STREAMING_UPDATE] Icon uploader not initialized, using default icon\n")
		return defaultIconID
	}

	mediaID, err := cm.iconUploader.UploadIcon(iconPath, iconName)
	if err != nil {
		fmt.Printf("[STREAMING_UPDATE] Failed to upload icon %s: %v, using default\n", iconName, err)
		return defaultIconID
	}

	if !strings.HasPrefix(mediaID, "yoto:#") {
		mediaID = fmt.Sprintf("yoto:#%s", mediaID)
	}

	return mediaID
}

func (cm *ContentManager) uploadBirdIconNoCache(iconPath string, iconName string) string {
	if cm.iconUploader == nil {
		fmt.Printf("[STREAMING_UPDATE] Icon uploader not initialized, using default icon\n")
		return defaultIconID
	}

	mediaID, err := cm.iconUploader.UploadIconNoCache(iconPath, iconName)
	if err != nil {
		fmt.Printf("[STREAMING_UPDATE] Failed to upload bird icon %s: %v, using default\n", iconName, err)
		return defaultIconID
	}

	if !strings.HasPrefix(mediaID, "yoto:#") {
		mediaID = fmt.Sprintf("yoto:#%s", mediaID)
	}

	return mediaID
}

func (cm *ContentManager) UpdateCardWithStreamingTracks(cardID string, birdName string, baseURL string, sessionID string) error {
	if err := cm.client.ensureAuthenticated(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	existingCard, err := cm.client.GetCard(cardID)
	if err != nil {
		fmt.Printf("[STREAMING_UPDATE] Warning: Could not get existing card %s: %v\n", cardID, err)
	}

	if sessionID == "" {
		sessionID = fmt.Sprintf("%s_%d", cardID, time.Now().Unix())
	}

	fmt.Printf("[STREAMING_UPDATE] Updating card with session %s for bird: %s\n", sessionID, birdName)

	binocularsIcon := cm.uploadTrackIcon("./assets/icons/binoculars_16x16.png", "binoculars")
	musicIcon := cm.uploadTrackIcon("./assets/icons/music_16x16.png", "music")

	var birdIcon string
	if birdName != "" {
		birdDir := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
		birdSpecificIconPath := fmt.Sprintf("./assets/icons/%s.png", birdDir)

		// Try bird-specific icon first
		if _, err := os.Stat(birdSpecificIconPath); err == nil {
			fmt.Printf("[STREAMING_UPDATE] Uploading bird-specific icon: %s\n", birdSpecificIconPath)
			birdIcon = cm.uploadBirdIconNoCache(birdSpecificIconPath, birdDir)
		} else {
			// Fallback to generic bird icon
			fmt.Printf("[STREAMING_UPDATE] Bird-specific icon not found at %s, using generic bird icon\n", birdSpecificIconPath)
			birdIcon = cm.uploadTrackIcon("./assets/icons/bird_16x16.png", "bird")
		}
	} else {
		birdIcon = cm.uploadTrackIcon("./assets/icons/bird_16x16.png", "bird")
	}

	hikingBootIcon := cm.uploadTrackIcon("./assets/icons/hiking_boot_16x16.png", "hiking_boot")

	chapters := []StreamingChapter{
		{
			Key:          "01",
			Title:        "Welcome, Explorers!",
			OverlayLabel: "1",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Welcome, Explorers!",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/intro?session=%s", baseURL, sessionID),
					Type:         "stream",
					Format:       "mp3",
					Duration:     30,
					OverlayLabel: "1",
					Display: Display{
						Icon16x16: binocularsIcon,
					},
				},
			},
			Display: Display{
				Icon16x16: binocularsIcon,
			},
		},
		{
			Key:          "02",
			Title:        "Who's Singing Today?",
			OverlayLabel: "2",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Who's Singing Today?",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/announcement?session=%s", baseURL, sessionID),
					Type:         "stream",
					Format:       "mp3",
					Duration:     10,
					OverlayLabel: "2",
					Display: Display{
						Icon16x16: musicIcon,
					},
				},
			},
			Display: Display{
				Icon16x16: musicIcon,
			},
		},
		{
			Key:          "03",
			Title:        "Bird Explorer's Guide",
			OverlayLabel: "3",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Bird Explorer's Guide",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/description?session=%s", baseURL, sessionID),
					Type:         "stream",
					Format:       "mp3",
					Duration:     60,
					OverlayLabel: "3",
					Display: Display{
						Icon16x16: birdIcon,
					},
				},
			},
			Display: Display{
				Icon16x16: birdIcon,
			},
		},
		{
			Key:          "04",
			Title:        "Happy Exploring!",
			OverlayLabel: "4",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Happy Exploring!",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/outro?session=%s", baseURL, sessionID),
					Type:         "stream",
					Format:       "mp3",
					Duration:     20,
					OverlayLabel: "4",
					Display: Display{
						Icon16x16: hikingBootIcon,
					},
				},
			},
			Display: Display{
				Icon16x16: hikingBootIcon,
			},
		},
	}

	metadataMap := make(map[string]interface{})

	if existingCard != nil && existingCard.Metadata != nil {
		if cover, hasCover := existingCard.Metadata["cover"]; hasCover {
			metadataMap["cover"] = cover
		}
	}

	contentReq := map[string]interface{}{
		"cardId": cardID,
		"content": map[string]interface{}{
			"title":    "Bird Song Explorer",
			"chapters": chapters,
			"metadata": metadataMap,
		},
	}

	jsonData, err := json.Marshal(contentReq)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	url := fmt.Sprintf("%s/content", cm.client.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cm.client.accessToken))

	resp, err := cm.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to update card content (status %d): %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully updated card %s with streaming tracks for %s\n", cardID, birdName)
	return nil
}
