package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/pkg/yoto"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Create Yoto client
	client := yoto.NewClient(cfg.YotoAccessToken, cfg.YotoRefreshToken, cfg.YotoClientID)
	
	// Create content manager
	cm := client.NewContentManager()

	// Sample data for testing
	cardID := "test-card-id"
	birdName := "American Robin"
	introURL := "https://example.com/intro.mp3"
	birdSongURL := "https://example.com/bird-song.mp3"
	birdDescription := "The American Robin is a migratory songbird with a reddish-orange breast."
	voiceID := "test-voice-id"

	// Build the request structure that would be sent
	chapters := []yoto.Chapter{
		{
			Key:          "01",
			Title:        "Introduction",
			OverlayLabel: "1",
			Tracks: []yoto.PlaylistTrack{
				{
					Key:          "01",
					Title:        "Welcome to Bird Song Explorer",
					TrackURL:     "yoto:#sample-intro-sha",
					Duration:     10,
					FileSize:     1024000,
					Channels:     2,
					Format:       "mp3",
					Type:         "audio",
					OverlayLabel: "1",
					Display: yoto.Display{
						Icon16x16: "yoto:#mmQkTUoEDBtnNVJNZy10GH3_c58aybuOeNoJv5pTo1Y",
					},
				},
			},
			Display: yoto.Display{
				Icon16x16: "yoto:#mmQkTUoEDBtnNVJNZy10GH3_c58aybuOeNoJv5pTo1Y",
			},
		},
		{
			Key:          "02",
			Title:        birdName,
			OverlayLabel: "2",
			Tracks: []yoto.PlaylistTrack{
				{
					Key:          "01",
					Title:        birdName + " Song",
					TrackURL:     "yoto:#sample-bird-sha",
					Duration:     30,
					FileSize:     3072000,
					Channels:     2,
					Format:       "mp3",
					Type:         "audio",
					OverlayLabel: "2",
					Display: yoto.Display{
						Icon16x16: "yoto:#R-60m21dr9Al8KQCy79k7lScYFRBBCvyYRbIZSDN_0Y",
					},
				},
			},
			Display: yoto.Display{
				Icon16x16: "yoto:#R-60m21dr9Al8KQCy79k7lScYFRBBCvyYRbIZSDN_0Y",
			},
		},
		{
			Key:          "03",
			Title:        "Today's Bird",
			OverlayLabel: "3",
			Tracks: []yoto.PlaylistTrack{
				{
					Key:          "01",
					Title:        "Today's Bird Announcement",
					TrackURL:     "yoto:#sample-announcement-sha",
					Duration:     5,
					FileSize:     512000,
					Channels:     2,
					Format:       "mp3",
					Type:         "audio",
					OverlayLabel: "3",
					Display: yoto.Display{
						Icon16x16: "yoto:#Cz1-d_jBfvwrbtt-CCyGS3T1mgASHQ8BDhzvtJ2J6Wg",
					},
				},
			},
			Display: yoto.Display{
				Icon16x16: "yoto:#Cz1-d_jBfvwrbtt-CCyGS3T1mgASHQ8BDhzvtJ2J6Wg",
			},
		},
		{
			Key:          "04",
			Title:        "Bird Explorer's Guide",
			OverlayLabel: "4",
			Tracks: []yoto.PlaylistTrack{
				{
					Key:          "01",
					Title:        "Bird Explorer's Guide",
					TrackURL:     "yoto:#sample-description-sha",
					Duration:     15,
					FileSize:     1536000,
					Channels:     2,
					Format:       "mp3",
					Type:         "audio",
					OverlayLabel: "4",
					Display: yoto.Display{
						Icon16x16: "yoto:#oCMXp05T6goR11wDmp2jr4KCEi8_i1KBfISZgKWyU48",
					},
				},
			},
			Display: yoto.Display{
				Icon16x16: "yoto:#oCMXp05T6goR11wDmp2jr4KCEi8_i1KBfISZgKWyU48",
			},
		},
		{
			Key:          "05",
			Title:        "See You Tomorrow!",
			OverlayLabel: "5",
			Tracks: []yoto.PlaylistTrack{
				{
					Key:          "01",
					Title:        "See You Tomorrow, Explorers!",
					TrackURL:     "yoto:#sample-outro-sha",
					Duration:     8,
					FileSize:     819200,
					Channels:     2,
					Format:       "mp3",
					Type:         "audio",
					OverlayLabel: "5",
					Display: yoto.Display{
						Icon16x16: "yoto:#kmmtUHk9_SEN1dTOSXJyeCjEVkxXmHwWDs36SMVqtYQ",
					},
				},
			},
			Display: yoto.Display{
				Icon16x16: "yoto:#kmmtUHk9_SEN1dTOSXJyeCjEVkxXmHwWDs36SMVqtYQ",
			},
		},
	}

	// Calculate totals
	totalDuration := 10 + 30 + 5 + 15 + 8  // Sum of all track durations
	totalSize := 1024000 + 3072000 + 512000 + 1536000 + 819200

	// Build the complete request
	updateReq := map[string]interface{}{
		"cardId":            cardID,
		"userId":            "sample-user-id",
		"createdByClientId": cfg.YotoClientID,
		"title":             "Bird Song Explorer",
		"content": map[string]interface{}{
			"chapters": chapters,
		},
		"metadata": map[string]interface{}{
			"media": map[string]interface{}{
				"duration":         totalDuration,
				"fileSize":         totalSize,
				"readableFileSize": float64(totalSize) / 1024 / 1024,
			},
			"cover": map[string]interface{}{
				"imageL": map[string]interface{}{
					"mediaId": "yoto:#sample-cover-image-id",
				},
			},
		},
		"createdAt": time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
	}

	// Convert to JSON with nice formatting
	jsonData, err := json.MarshalIndent(updateReq, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Save to file
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("yoto_content_request_sample_%s.json", timestamp)
	
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sample Yoto content request saved to: %s\n", filename)
	fmt.Printf("\nThis is the structure of the JSON request that would be sent to:\n")
	fmt.Printf("POST https://api.yotoplay.com/content\n")
	fmt.Printf("\nWith headers:\n")
	fmt.Printf("- Authorization: Bearer [access_token]\n")
	fmt.Printf("- Content-Type: application/json\n")
}