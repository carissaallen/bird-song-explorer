package yoto

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

	// Search for bird icon
	iconMediaID := ""
	if cm.iconSearcher != nil {
		iconResult, err := cm.iconSearcher.SearchBirdIcon(birdName)
		if err == nil && iconResult != "" {
			iconMediaID = FormatIconID(iconResult)
			fmt.Printf("Found icon for %s: %s\n", birdName, iconMediaID)
		}
	}

	if iconMediaID == "" {
		// Use meadowlark as default bird icon
		iconMediaID = "yoto:#OOKWbJLOXojHvDuWdJLWs91LVP0yA9s8FBX0fQ4xP7Y"
		fmt.Printf("No specific icon found for %s on yotoicons.com, will use meadowlark default\n", birdName)
	}

	// Use the provided session ID if available, otherwise generate one
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s_%d", cardID, time.Now().Unix())
	}

	// URL encode the bird name for use in query parameters
	encodedBirdName := url.QueryEscape(birdName)

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
						Icon16x16: getRandomRadioIcon(),
					},
				},
			},
			Display: Display{
				Icon16x16: getRandomRadioIcon(),
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
						Icon16x16: "yoto:#Cz1-d_jBfvwrbtt-CCyGS3T1mgASHQ8BDhzvtJ2J6Wg", // Binoculars
					},
				},
			},
			Display: Display{
				Icon16x16: "yoto:#Cz1-d_jBfvwrbtt-CCyGS3T1mgASHQ8BDhzvtJ2J6Wg", // Binoculars
			},
		},
		{
			Key:          "03",
			Title:        fmt.Sprintf("%s Song", birdName),
			OverlayLabel: "3",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        fmt.Sprintf("%s Song", birdName),
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/bird-song?session=%s&bird=%s", baseURL, sessionID, encodedBirdName),
					Type:         "stream",
					Format:       "mp3",
					Duration:     30,
					OverlayLabel: "3",
					Display: Display{
						Icon16x16: iconMediaID,
					},
				},
			},
			Display: Display{
				Icon16x16: iconMediaID,
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
						Icon16x16: getRandomBookIcon(),
					},
				},
			},
			Display: Display{
				Icon16x16: getRandomBookIcon(),
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
						Icon16x16: "yoto:#kmmtUHk9_SEN1dTOSXJyeCjEVkxXmHwWDs36SMVqtYQ", // Hiking boot
					},
				},
			},
			Display: Display{
				Icon16x16: "yoto:#kmmtUHk9_SEN1dTOSXJyeCjEVkxXmHwWDs36SMVqtYQ", // Hiking boot
			},
		},
	}

	// Extract user ID and metadata from existing card
	userID := ""
	createdAt := ""
	clientID := cm.client.clientID

	if existingCard != nil {
		userID = existingCard.UserID
		createdAt = existingCard.CreatedAt
		if existingCard.CreatedByClientID != "" {
			clientID = existingCard.CreatedByClientID
		}
	}

	// If no userID from card, try to extract from JWT token
	if userID == "" && cm.client.accessToken != "" {
		parts := strings.Split(cm.client.accessToken, ".")
		if len(parts) >= 2 {
			payload, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err == nil {
				var tokenData map[string]interface{}
				if err := json.Unmarshal(payload, &tokenData); err == nil {
					if sub, ok := tokenData["sub"].(string); ok {
						userID = sub
					}
				}
			}
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if createdAt == "" {
		createdAt = now
	}

	metadataMap := make(map[string]interface{})

	if existingCard != nil && existingCard.Metadata != nil {
		if cover, hasCover := existingCard.Metadata["cover"]; hasCover {
			metadataMap["cover"] = cover // Preserve the existing cover
		}
	}

	// Build the content update request
	// Include all fields needed for updating existing card content
	contentReq := map[string]interface{}{
		"cardId":            cardID,
		"userId":            userID,
		"createdByClientId": clientID,
		"title":             "Bird Song Explorer",
		"content": map[string]interface{}{
			"chapters": chapters,
		},
		"createdAt": createdAt,
		"updatedAt": now,
	}

	// Only add metadata if we have something to preserve (like cover)
	if len(metadataMap) > 0 {
		contentReq["metadata"] = metadataMap
	}

	jsonData, err := json.Marshal(contentReq)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	// Debug logging
	fmt.Printf("[STREAMING_UPDATE] Sending request to Yoto API:\n")
	fmt.Printf("  cardId: %s\n", cardID)
	fmt.Printf("  userId: %s\n", userID)
	fmt.Printf("  clientID: %s\n", clientID)
	fmt.Printf("  createdAt: %s\n", createdAt)
	fmt.Printf("  updatedAt: %s\n", now)
	fmt.Printf("  Request size: %d bytes\n", len(jsonData))

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
