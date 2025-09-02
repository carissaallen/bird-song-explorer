package yoto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type ContentManager struct {
	client               *Client
	uploader             *AudioUploader
	iconUploader         *IconUploader
	iconSearcher         *IconSearcher
	lastIntroText        string // Store intro text for transitions (see: 'previous_track')
	lastAnnouncementText string // Store announcement text for transitions (see: 'previous_track')
	lastDescriptionText  string // Store description text for transitions (see: 'previous_track')
	selectedAmbience     string // Store which ambience was used in intro for continuity
	ambienceData         []byte // Store ambience audio data for Track 2 and outro
}

type CreateContentResponse struct {
	CardID string `json:"cardId"` // The API returns cardId, not contentId
	Status string `json:"status"`
}

type UpdateCardContentRequest struct {
	ContentID string `json:"contentId"`
}

const defaultBirdIcon = "yoto:#R-60m21dr9Al8KQCy79k7lScYFRBBCvyYRbIZSDN_0Y"

// Radio icon media IDs for introduction tracks
var radioIconsManager = []string{
	"yoto:#mmQkTUoEDBtnNVJNZy10GH3_c58aybuOeNoJv5pTo1Y",
	"yoto:#1bB-31IMz24mz4XnEpCGU15q-Cu_gfUWUdH4ioZvjyI",
	"yoto:#tal2wCH_bPOitYdPZQUvUSPGK5YobvV-rCXPH7sVGIg",
	"yoto:#lVHrb3TZ1RZSPUNDhEWts1j1IrZYvsIb0J9BP_8ISLA",
	"yoto:#auOIXmLNhMt0W4rv0SMeNTZtvZ91O5xIsGRSA5VlE2s",
	"yoto:#nIGf1CHb9WEDO8uNV7uHdFK-Y2fLovO8EM-ULiBXT94",
}

// getRandomRadioIconManager returns a random radio icon from the available options
func getRandomRadioIconManager() string {
	rand.Seed(time.Now().UnixNano())
	return radioIconsManager[rand.Intn(len(radioIconsManager))]
}

func NewContentManager(client *Client) *ContentManager {
	return &ContentManager{
		client:       client,
		uploader:     NewAudioUploader(client),
		iconUploader: NewIconUploader(client),
		iconSearcher: NewIconSearcher(client),
	}
}

// NewContentManager creates a new content manager (method on Client for convenience)
func (c *Client) NewContentManager() *ContentManager {
	return NewContentManager(c)
}

// CreateBirdPlaylist creates a new playlist with intro and bird song
func (cm *ContentManager) CreateBirdPlaylist(birdName string, introURL string, birdSongURL string) (string, error) {
	if err := cm.client.ensureAuthenticated(); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	introSha, introInfo, err := cm.uploader.UploadAudioFromURL(introURL, "Bird Song Explorer Intro")
	if err != nil {
		return "", fmt.Errorf("failed to upload intro: %w", err)
	}

	birdSongSha, birdInfo, err := cm.uploader.UploadAudioFromURL(birdSongURL, birdName+" Song")
	if err != nil {
		return "", fmt.Errorf("failed to upload bird song: %w", err)
	}

	radioIcon := getRandomRadioIconManager()

	tracks := []PlaylistTrack{
		{
			Key:          "01",
			Title:        "Welcome to Bird Song Explorer",
			TrackURL:     fmt.Sprintf("yoto:#%s", introSha),
			Duration:     introInfo.GetDuration(),
			FileSize:     introInfo.GetFileSize(),
			Channels:     introInfo.GetChannels(),
			Format:       introInfo.Transcode.TranscodedInfo.Format,
			Type:         "audio",
			OverlayLabel: "1",
			Display: Display{
				Icon16x16: radioIcon,
			},
		},
		{
			Key:          "02",
			Title:        birdName + " Song",
			TrackURL:     fmt.Sprintf("yoto:#%s", birdSongSha),
			Duration:     birdInfo.GetDuration(),
			FileSize:     birdInfo.GetFileSize(),
			Channels:     birdInfo.GetChannels(),
			Format:       birdInfo.Transcode.TranscodedInfo.Format,
			Type:         "audio",
			OverlayLabel: "2",
			Display: Display{
				Icon16x16: defaultBirdIcon,
			},
		},
	}

	totalDuration := introInfo.GetDuration() + birdInfo.GetDuration()
	totalSize := introInfo.GetFileSize() + birdInfo.GetFileSize()

	chapters := []Chapter{
		{
			Key:          "01",
			Title:        "Today's Bird: " + birdName,
			OverlayLabel: "1",
			Tracks:       tracks,
			Display: Display{
				Icon16x16: radioIcon,
			},
		},
	}

	content := PlaylistContent{
		Title: "Bird Song Explorer - " + birdName,
		Content: Content{
			Chapters: chapters,
		},
		Metadata: Metadata{
			Media: MediaInfo{
				Duration:         totalDuration,
				FileSize:         totalSize,
				ReadableFileSize: float64(totalSize) / 1024 / 1024,
			},
		},
	}

	contentID, err := cm.createContent(content)
	if err != nil {
		return "", fmt.Errorf("failed to create playlist: %w", err)
	}

	return contentID, nil
}

// UpdateCardContent updates a MYO card with new content
func (cm *ContentManager) UpdateCardContent(cardID string, contentID string) error {
	if err := cm.client.ensureAuthenticated(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	url := fmt.Sprintf("%s/content/%s", cm.client.baseURL, cardID)

	updateReq := UpdateCardContentRequest{
		ContentID: contentID,
	}

	jsonData, err := json.Marshal(updateReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+cm.client.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cm.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update card: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func (cm *ContentManager) createContent(content PlaylistContent) (string, error) {
	url := fmt.Sprintf("%s/content", cm.client.baseURL)

	jsonData, err := json.Marshal(content)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+cm.client.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cm.client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create content: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result CreateContentResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Create content response: %s\n", string(body))
		return "", err
	}

	if result.CardID != "" {
		return result.CardID, nil
	}

	var altResult map[string]interface{}
	if err := json.Unmarshal(body, &altResult); err == nil {
		if cardID, ok := altResult["cardId"].(string); ok {
			return cardID, nil
		}
	}

	return "", fmt.Errorf("no card ID in response: %s", string(body))
}
