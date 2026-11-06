package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/pkg/yoto"
)

type PlaylistConfig struct {
	Title   string        `json:"title"`
	GCSBase string        `json:"gcs_base"`
	Tracks  []TrackConfig `json:"tracks"`
}

type TrackConfig struct {
	Title     string `json:"title"`
	AudioPath string `json:"audio_path"`
	Icon      string `json:"icon"`
}

func main() {
	cardID := flag.String("card-id", "dg4TT", "Yoto card ID to update")
	playlistFile := flag.String("playlist", "cmd/static_playlist/playlist.json", "Path to playlist JSON config")
	dryRun := flag.Bool("dry-run", false, "Print what would be done without making API calls")
	flag.Parse()

	// Load playlist config
	data, err := os.ReadFile(*playlistFile)
	if err != nil {
		log.Fatalf("Failed to read playlist file %s: %v", *playlistFile, err)
	}

	var playlist PlaylistConfig
	if err := json.Unmarshal(data, &playlist); err != nil {
		log.Fatalf("Failed to parse playlist file: %v", err)
	}

	fmt.Printf("Static Playlist Builder\n")
	fmt.Printf("  Card ID:  %s\n", *cardID)
	fmt.Printf("  Playlist: %s (%d tracks)\n", playlist.Title, len(playlist.Tracks))
	fmt.Printf("  Dry run:  %v\n\n", *dryRun)

	// Print track listing
	for i, track := range playlist.Tracks {
		audioURL := fmt.Sprintf("%s/%s", playlist.GCSBase, track.AudioPath)
		fmt.Printf("  %d. %s\n     %s\n     icon: %s\n", i+1, track.Title, audioURL, track.Icon)
	}
	fmt.Println()

	if *dryRun {
		fmt.Println("Dry run complete. No API calls made.")
		return
	}

	// Initialize Yoto client, pulling fresh tokens from Secret Manager via gcloud
	cfg := config.Load()

	gcpProject := os.Getenv("GCP_PROJECT")
	if gcpProject == "" {
		gcpProject = "yoto-bird-song-explorer"
	}

	accessToken := cfg.YotoAccessToken
	refreshToken := cfg.YotoRefreshToken

	if fresh, err := getSecretViaGcloud(gcpProject, "yoto-access-token"); err == nil {
		accessToken = fresh
		fmt.Println("  Loaded access token from Secret Manager")
	} else {
		fmt.Printf("  Warning: could not load access token from Secret Manager: %v\n", err)
		fmt.Println("  Falling back to .env token")
	}
	if fresh, err := getSecretViaGcloud(gcpProject, "yoto-refresh-token"); err == nil {
		refreshToken = fresh
		fmt.Println("  Loaded refresh token from Secret Manager")
	} else {
		fmt.Printf("  Warning: could not load refresh token from Secret Manager: %v\n", err)
		fmt.Println("  Falling back to .env token")
	}

	client := yoto.NewClient(cfg.YotoClientID, "", cfg.YotoAPIBaseURL)
	// Use a long expiry so the client tries the access token as-is first.
	// The token from Secret Manager was issued on the last Cloud Run refresh.
	client.SetTokens(accessToken, refreshToken, 86400)

	cm := client.NewContentManager()
	iconUploader := yoto.NewIconUploader(client)

	// Build chapters
	var chapters []yoto.StreamingChapter

	for i, track := range playlist.Tracks {
		key := fmt.Sprintf("%02d", i+1)
		overlayLabel := fmt.Sprintf("%d", i+1)
		audioURL := fmt.Sprintf("%s/%s", playlist.GCSBase, track.AudioPath)

		// Upload icon
		iconID := yoto.DefaultIconID
		if _, statErr := os.Stat(track.Icon); statErr == nil {
			iconName := strings.TrimSuffix(strings.TrimPrefix(track.Icon, "./assets/icons/"), ".png")
			mediaID, uploadErr := iconUploader.UploadIcon(track.Icon, iconName)
			if uploadErr != nil {
				fmt.Printf("  Warning: failed to upload icon %s: %v, using default\n", track.Icon, uploadErr)
			} else {
				iconID = yoto.FormatIconID(mediaID)
				fmt.Printf("  Uploaded icon: %s -> %s\n", track.Icon, iconID)
			}
		} else {
			fmt.Printf("  Warning: icon not found at %s, using default\n", track.Icon)
		}

		chapters = append(chapters, yoto.StreamingChapter{
			Key:          key,
			Title:        track.Title,
			OverlayLabel: overlayLabel,
			Tracks: []yoto.StreamingTrack{
				{
					Key:      "01",
					Title:    track.Title,
					TrackURL: audioURL,
					Type:     "stream",
					Format:   "mp3",
					Duration: 60,
					Display: yoto.Display{
						Icon16x16: iconID,
					},
				},
			},
			Display: yoto.Display{
				Icon16x16: iconID,
			},
		})
	}

	fmt.Printf("\nUpdating card %s with %d tracks...\n", *cardID, len(chapters))

	if err := cm.UpdateCardWithStaticPlaylist(*cardID, playlist.Title, chapters); err != nil {
		log.Fatalf("Failed to update card: %v", err)
	}

	fmt.Printf("Card %s updated successfully!\n", *cardID)
}

// getSecretViaGcloud fetches a secret from GCP Secret Manager using the gcloud CLI,
// which uses the user's authenticated gcloud credentials.
func getSecretViaGcloud(projectID, secretName string) (string, error) {
	out, err := exec.Command("gcloud", "secrets", "versions", "access", "latest",
		"--secret="+secretName, "--project="+projectID).Output()
	if err != nil {
		return "", fmt.Errorf("gcloud secrets access failed for %s: %w", secretName, err)
	}
	return strings.TrimSpace(string(out)), nil
}
