package yoto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// StreamingChapter represents a chapter with streaming tracks
type StreamingChapter struct {
	Key          string           `json:"key"`
	Title        string           `json:"title"`
	OverlayLabel string           `json:"overlayLabel,omitempty"`
	Tracks       []StreamingTrack `json:"tracks"`
	Display      Display          `json:"display,omitempty"`
}

// StreamingTrack represents a streaming audio track
type StreamingTrack struct {
	Key          string  `json:"key"`
	Title        string  `json:"title,omitempty"`
	TrackURL     string  `json:"trackUrl"`
	Type         string  `json:"type"`     // Must be "stream"
	Format       string  `json:"format"`   // e.g., "mp3"
	Duration     int     `json:"duration"` // Required, even for streaming
	OverlayLabel string  `json:"overlayLabel,omitempty"`
	Display      Display `json:"display,omitempty"`
}

// StreamingContent represents the content structure for streaming
type StreamingContent struct {
	Chapters []StreamingChapter `json:"chapters"`
}

// UpdateCardWithStreamingTracks updates a card to use streaming URLs
func (cm *ContentManager) UpdateCardWithStreamingTracks(cardID string, birdName string, baseURL string, sessionID string) error {
	if err := cm.client.ensureAuthenticated(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Get existing card to preserve metadata (userId, createdAt, etc.)
	existingCard, err := cm.client.GetCard(cardID)
	if err != nil {
		fmt.Printf("[STREAMING_UPDATE] Warning: Could not get existing card %s: %v\n", cardID, err)
	} else {
		fmt.Printf("[STREAMING_UPDATE] Successfully retrieved existing card %s\n", cardID)
		if existingCard != nil {
			fmt.Printf("  Existing userId: %s\n", existingCard.UserID)
			fmt.Printf("  Existing createdAt: %s\n", existingCard.CreatedAt)
			fmt.Printf("  Existing clientID: %s\n", existingCard.CreatedByClientID)
		}
	}

	// Use the provided session ID if available, otherwise generate one
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s_%d", cardID, time.Now().Unix())
	}

	encodedBirdName := url.QueryEscape(birdName)

	fmt.Printf("[STREAMING_UPDATE] Using dynamic icon URLs with baseURL: %s\n", baseURL)
	fmt.Printf("[STREAMING_UPDATE] Example icon URL: %s/assets/icons/hiking-boot.png\n", baseURL)

	chapters := []StreamingChapter{
		{
			Key:          "01",
			Title:        "Welcome, Explorers!",
			OverlayLabel: "1",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Welcome, Explorers!",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/intro?session=%s&bird=%s", baseURL, sessionID, encodedBirdName),
					Type:         "stream",
					Format:       "mp3",
					Duration:     30,
					OverlayLabel: "1",
					Display: Display{
						IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/radio_16x16.png",
					},
				},
			},
			Display: Display{
				IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/radio_16x16.png",
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
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/announcement?session=%s&bird=%s", baseURL, sessionID, encodedBirdName),
					Type:         "stream",
					Format:       "mp3",
					Duration:     10,
					OverlayLabel: "2",
					Display: Display{
						IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/binoculars_16x16.png",
					},
				},
			},
			Display: Display{
				IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/binoculars_16x16.png",
			},
		},
		{
			Key:          "03",
			Title:        "Bird Song",
			OverlayLabel: "3",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Bird Song",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/bird-song?session=%s&bird=%s", baseURL, sessionID, encodedBirdName),
					Type:         "stream",
					Format:       "mp3",
					Duration:     30,
					OverlayLabel: "3",
					Display: Display{
						IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/bird_16x16.png",
					},
				},
			},
			Display: Display{
				IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/bird_16x16.png",
			},
		},
		{
			Key:          "04",
			Title:        "Bird Explorer's Guide",
			OverlayLabel: "4",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Bird Explorer's Guide",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/description?session=%s&bird=%s", baseURL, sessionID, encodedBirdName),
					Type:         "stream",
					Format:       "mp3",
					Duration:     60,
					OverlayLabel: "4",
					Display: Display{
						IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/book_16x16.png",
					},
				},
			},
			Display: Display{
				IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/book_16x16.png",
			},
		},
		{
			Key:          "05",
			Title:        "Happy Exploring!",
			OverlayLabel: "5",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Happy Exploring!",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/outro?session=%s&bird=%s", baseURL, sessionID, encodedBirdName),
					Type:         "stream",
					Format:       "mp3",
					Duration:     20,
					OverlayLabel: "5",
					Display: Display{
						IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/hiking_boot_16x16.png",
					},
				},
			},
			Display: Display{
				IconUrl16x16: "https://raw.githubusercontent.com/carissaallen/bird-song-explorer/convert-to-streaming/assets/icons/hiking_boot_16x16.png",
			},
		},
	}

	// Extract metadata from existing card (if any)

	metadataMap := make(map[string]interface{})

	if existingCard != nil && existingCard.Metadata != nil {
		if cover, hasCover := existingCard.Metadata["cover"]; hasCover {
			metadataMap["cover"] = cover // Preserve the existing cover
		}
	}

	// Build the content update request according to Yoto API spec
	// Only cardId and content object are allowed at root level
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

	// Debug logging with full request
	fmt.Printf("[STREAMING_UPDATE] Sending request to Yoto API:\n")
	fmt.Printf("  cardId: %s\n", cardID)
	fmt.Printf("  title: Bird Song Explorer\n")
	fmt.Printf("  chapters: %d\n", len(chapters))
	fmt.Printf("  Request size: %d bytes\n", len(jsonData))

	// Log the actual JSON for debugging
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, jsonData, "", "  ")
	fmt.Printf("[STREAMING_UPDATE] Full request JSON:\n%s\n", prettyJSON.String())

	url := fmt.Sprintf("%s/content", cm.client.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cm.client.accessToken))

	// Execute the request
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

	// Debug logging
	responsePreview := string(body)
	if len(responsePreview) > 500 {
		responsePreview = responsePreview[:500] + "..."
	}
	fmt.Printf("Update response (%d): %s\n", resp.StatusCode, responsePreview)

	return nil
}
