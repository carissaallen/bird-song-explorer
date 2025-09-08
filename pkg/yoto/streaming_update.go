package yoto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Type         string  `json:"type"`   // Must be "stream"
	Format       string  `json:"format"` // e.g., "mp3"
	Duration     int     `json:"duration"` // Required, even for streaming
	OverlayLabel string  `json:"overlayLabel,omitempty"`
	Display      Display `json:"display,omitempty"`
}

// StreamingContent represents the content structure for streaming
type StreamingContent struct {
	Chapters []StreamingChapter `json:"chapters"`
}


// UpdateCardWithStreamingTracks updates a card to use streaming URLs
func (cm *ContentManager) UpdateCardWithStreamingTracks(cardID string, birdName string, baseURL string) error {
	if err := cm.client.ensureAuthenticated(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
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

	// Use icon or fallback to radio icons
	if iconMediaID == "" {
		iconMediaID = getRandomRadioIcon()
	}

	// Create streaming chapters with session parameter
	// Session will be created on first track request
	chapters := []StreamingChapter{
		{
			Key:          "01",
			Title:        "Introduction",
			OverlayLabel: "1",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Introduction",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/intro", baseURL),
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
			Title:        "Today's Bird",
			OverlayLabel: "2",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Bird Announcement",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/announcement", baseURL),
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
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/bird-song", baseURL),
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
			Title:        "Bird Description",
			OverlayLabel: "4",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "Bird Description",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/description", baseURL),
					Type:         "stream",
					Format:       "mp3",
					Duration:     60,
					OverlayLabel: "4",
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
			Key:          "05",
			Title:        "See You Tomorrow",
			OverlayLabel: "5",
			Tracks: []StreamingTrack{
				{
					Key:          "01",
					Title:        "See You Tomorrow Explorers",
					TrackURL:     fmt.Sprintf("%s/api/v1/stream/outro", baseURL),
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

	// Build the content creation request (without CardID)
	contentReq := map[string]interface{}{
		"title": fmt.Sprintf("Bird Song Explorer - %s", birdName),
		"content": map[string]interface{}{
			"chapters": chapters,
		},
		"metadata": map[string]interface{}{
			"description": fmt.Sprintf("Streaming playlist for %s", birdName),
		},
	}

	// Marshal the request
	jsonData, err := json.Marshal(contentReq)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	// Make the API request
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
		return fmt.Errorf("failed to create streaming content (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse the response to get the content ID
	var contentResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &contentResp); err != nil {
		return fmt.Errorf("failed to parse content response: %w", err)
	}

	fmt.Printf("Successfully created streaming content with ID %s for %s\n", contentResp.ID, birdName)
	
	// Now associate the content with the card
	associateURL := fmt.Sprintf("%s/card/%s/content/%s", cm.client.baseURL, cardID, contentResp.ID)
	associateReq, err := http.NewRequest("POST", associateURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create associate request: %w", err)
	}
	
	associateReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cm.client.accessToken))
	
	associateResp, err := cm.client.httpClient.Do(associateReq)
	if err != nil {
		return fmt.Errorf("failed to associate content with card: %w", err)
	}
	defer associateResp.Body.Close()
	
	associateBody, _ := io.ReadAll(associateResp.Body)
	
	if associateResp.StatusCode != http.StatusOK && associateResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to associate content with card (status %d): %s", associateResp.StatusCode, string(associateBody))
	}
	
	fmt.Printf("Successfully associated streaming content with card %s\n", cardID)

	return nil
}