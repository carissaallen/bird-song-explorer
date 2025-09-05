package yoto

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

// UpdateContentRequest represents the request body for updating card content
type UpdateContentRequest struct {
	CardID            string   `json:"cardId"` // Use cardId for POST requests
	UserID            string   `json:"userId"`
	CreatedByClientID string   `json:"createdByClientId"`
	Title             string   `json:"title"`
	Content           Content  `json:"content"`
	Metadata          Metadata `json:"metadata"`
	CreatedAt         string   `json:"createdAt,omitempty"`
	UpdatedAt         string   `json:"updatedAt"`
}

// Radio icon media IDs (different colored radios)
var radioIcons = []string{
	"yoto:#mmQkTUoEDBtnNVJNZy10GH3_c58aybuOeNoJv5pTo1Y",
	"yoto:#1bB-31IMz24mz4XnEpCGU15q-Cu_gfUWUdH4ioZvjyI",
	"yoto:#tal2wCH_bPOitYdPZQUvUSPGK5YobvV-rCXPH7sVGIg",
	"yoto:#lVHrb3TZ1RZSPUNDhEWts1j1IrZYvsIb0J9BP_8ISLA",
	"yoto:#auOIXmLNhMt0W4rv0SMeNTZtvZ91O5xIsGRSA5VlE2s",
	"yoto:#nIGf1CHb9WEDO8uNV7uHdFK-Y2fLovO8EM-ULiBXT94",
}

// Book icon media IDs (different colored books for description track)
var bookIcons = []string{
	"yoto:#oCMXp05T6goR11wDmp2jr4KCEi8_i1KBfISZgKWyU48",
	"yoto:#ByDf_m-0HtI6EogcFHoIGGqigQC2OT-4WsHjL4B53C0",
	"yoto:#iZoAZGVVtrcfEPfnXftcag0itN17SjSnxSV9pD4aXHA",
	"yoto:#99i8i93d17yfLxD9cDjXfgSL6B9A0XfIkJo513VYJ0U",
	"yoto:#kmmtUHk9_SEN1dTOSXJyeCjEVkxXmHwWDs36SMVqtYQ",
}

// getRandomRadioIcon returns a random radio icon from the available options
func getRandomRadioIcon() string {
	rand.Seed(time.Now().UnixNano())
	return radioIcons[rand.Intn(len(radioIcons))]
}

// getRandomBookIcon returns a random book icon from the available options
func getRandomBookIcon() string {
	rand.Seed(time.Now().UnixNano())
	return bookIcons[rand.Intn(len(bookIcons))]
}

// UpdateExistingCardContentWithDescriptionAndVoice updates an existing MYO card with new content including bird description and specific voice
func (cm *ContentManager) UpdateExistingCardContentWithDescriptionAndVoice(cardID string, birdName string, introURL string, birdSongURL string, birdDescription string, voiceID string) error {
	// For now, call the location-aware version with zero coordinates
	return cm.UpdateExistingCardContentWithDescriptionVoiceAndLocation(cardID, birdName, introURL, birdSongURL, birdDescription, voiceID, 0, 0)
}

// UpdateExistingCardContentWithDescriptionVoiceAndLocation updates an existing MYO card with location-aware content
func (cm *ContentManager) UpdateExistingCardContentWithDescriptionVoiceAndLocation(cardID string, birdName string, introURL string, birdSongURL string, birdDescription string, voiceID string, latitude, longitude float64) error {
	// Voice ID must be provided explicitly
	if voiceID == "" {
		return fmt.Errorf("voice ID must be provided explicitly")
	}

	if err := cm.client.ensureAuthenticated(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Get existing card to preserve metadata
	existingCard, err := cm.client.GetCard(cardID)
	if err != nil {
	}

	// Extract intro text from the URL if it's a pre-recorded intro
	cm.extractIntroTextFromURL(introURL)

	// Check if this is an enhanced intro and capture ambience data
	cm.captureAmbienceFromEnhancedIntro(introURL)

	introSha, introInfo, err := cm.uploader.UploadAudioFromURL(introURL, "Bird Song Explorer Intro")
	if err != nil {
		return fmt.Errorf("failed to upload intro: %w", err)
	}

	var announcementSha string
	var announcementInfo *TranscodeResponse
	var hasAnnouncement bool

	announcementData, err := cm.generateBirdAnnouncement(birdName, voiceID)
	if err != nil {
		hasAnnouncement = false
	} else {
		announcementSha, announcementInfo, err = cm.uploader.UploadAudioData(announcementData, "Bird Announcement")
		if err != nil {
			hasAnnouncement = false
		} else {
			hasAnnouncement = true
		}
	}

	var binocularsIcon string
	if hasAnnouncement {
		binocularsID, err := cm.iconUploader.GetBinocularsIcon()
		if err != nil {
			binocularsIcon = "yoto:#Cz1-d_jBfvwrbtt-CCyGS3T1mgASHQ8BDhzvtJ2J6Wg" // Previously uploaded binoculars
		} else {
			binocularsIcon = FormatIconID(binocularsID)
		}
	}

	birdSongSha, birdInfo, err := cm.uploader.UploadAudioFromURL(birdSongURL, birdName+" Song")
	if err != nil {
		return fmt.Errorf("failed to upload bird song: %w", err)
	}

	// No longer storing bird song data for outro - we'll use ambience instead

	var descriptionSha string
	var descriptionInfo *TranscodeResponse
	var hasDescription bool

	if birdDescription != "" {
		// Use enhanced fact generator if we have location data
		var descriptionData []byte
		var err error

		// Use standard generator which sounds better with TTS
		// The simpler, more conversational text works better than complex V4 facts
		fmt.Printf("[CONTENT_UPDATE] Using standard facts generator for better TTS quality\n")
		descriptionData, err = cm.generateBirdDescription(birdDescription, birdName, voiceID)

		if err != nil {
			hasDescription = false
		} else {
			descriptionSha, descriptionInfo, err = cm.uploader.UploadAudioData(descriptionData, "Bird Description")
			if err != nil {
				hasDescription = false
			} else {
				hasDescription = true
			}
		}
	}

	totalDuration := introInfo.GetDuration() + birdInfo.GetDuration()
	totalSize := introInfo.GetFileSize() + birdInfo.GetFileSize()

	if hasAnnouncement {
		totalDuration += announcementInfo.GetDuration()
		totalSize += announcementInfo.GetFileSize()
	}

	if hasDescription {
		totalDuration += descriptionInfo.GetDuration()
		totalSize += descriptionInfo.GetFileSize()
	}

	// Generate outro
	var outroSha string
	var outroInfo *TranscodeResponse
	var hasOutro bool

	outroData, err := cm.generateOutro(birdName, voiceID)
	if err != nil {
		// Log the error but don't fail the entire update
		fmt.Printf("Warning: Failed to generate outro: %v\n", err)
		hasOutro = false
	} else if outroData == nil || len(outroData) == 0 {
		fmt.Printf("Warning: Outro data is empty\n")
		hasOutro = false
	} else {
		outroSha, outroInfo, err = cm.uploader.UploadAudioData(outroData, "See You Tomorrow Explorers")
		if err != nil {
			fmt.Printf("Warning: Failed to upload outro: %v\n", err)
			hasOutro = false
		} else {
			hasOutro = true
			totalDuration += outroInfo.GetDuration()
			totalSize += outroInfo.GetFileSize()
		}
	}

	var hikingBootIcon string
	if hasOutro {
		if cm.iconUploader != nil {
			hikingBootID, err := cm.iconUploader.GetHikingBootIcon()
			if err != nil {
				// Use a star or other adventure-themed icon as fallback
				hikingBootIcon = "yoto:#kmmtUHk9_SEN1dTOSXJyeCjEVkxXmHwWDs36SMVqtYQ" // Using a book icon as fallback
			} else {
				hikingBootIcon = FormatIconID(hikingBootID)
			}
		}
	}

	radioIcon := getRandomRadioIcon()

	chapters := []Chapter{
		{
			Key:          "01",
			Title:        "Introduction",
			OverlayLabel: "1",
			Tracks: []PlaylistTrack{
				{
					Key:          "01",
					Title:        "Introduction",
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
			},
			Display: Display{
				Icon16x16: radioIcon,
			},
		},
	}

	if hasAnnouncement {
		chapters = append(chapters, Chapter{
			Key:          "02",
			Title:        "Today's Bird",
			OverlayLabel: "2",
			Tracks: []PlaylistTrack{
				{
					Key:          "01",
					Title:        "Today's Bird",
					TrackURL:     fmt.Sprintf("yoto:#%s", announcementSha),
					Duration:     announcementInfo.GetDuration(),
					FileSize:     announcementInfo.GetFileSize(),
					Channels:     announcementInfo.GetChannels(),
					Format:       announcementInfo.Transcode.TranscodedInfo.Format,
					Type:         "audio",
					OverlayLabel: "2",
					Display: Display{
						Icon16x16: binocularsIcon,
					},
				},
			},
			Display: Display{
				Icon16x16: binocularsIcon,
			},
		})
	}

	birdChapterKey := "02"
	birdOverlayLabel := "2"
	if hasAnnouncement {
		birdChapterKey = "03"
		birdOverlayLabel = "3"
	}

	var birdIcon string

	if cm.iconUploader != nil {
		if meadowlarkID, err := cm.iconUploader.GetMeadowlarkIcon(); err == nil {
			birdIcon = FormatIconID(meadowlarkID)
		} else {
			// Only fall back to the static bird icon if meadowlark upload fails
			birdIcon = defaultBirdIcon
		}
	} else {
		// If no icon uploader, use the static fallback
		birdIcon = defaultBirdIcon
	}

	if cm.iconSearcher != nil {
		if searchedIcon, err := cm.iconSearcher.SearchBirdIcon(birdName); err == nil && searchedIcon != "" {
			birdIcon = searchedIcon
		}
	}

	chapters = append(chapters, Chapter{
		Key:          birdChapterKey,
		Title:        birdName,
		OverlayLabel: birdOverlayLabel,
		Tracks: []PlaylistTrack{
			{
				Key:          "01",
				Title:        birdName,
				TrackURL:     fmt.Sprintf("yoto:#%s", birdSongSha),
				Duration:     birdInfo.GetDuration(),
				FileSize:     birdInfo.GetFileSize(),
				Channels:     birdInfo.GetChannels(),
				Format:       birdInfo.Transcode.TranscodedInfo.Format,
				Type:         "audio",
				OverlayLabel: birdOverlayLabel,
				Display: Display{
					Icon16x16: birdIcon,
				},
			},
		},
		Display: Display{
			Icon16x16: birdIcon,
		},
	})

	if hasDescription {
		descChapterKey := "03"
		descOverlayLabel := "3"
		if hasAnnouncement {
			descChapterKey = "04"
			descOverlayLabel = "4"
		}

		bookIcon := getRandomBookIcon()

		chapters = append(chapters, Chapter{
			Key:          descChapterKey,
			Title:        "Bird Explorer's Guide",
			OverlayLabel: descOverlayLabel,
			Tracks: []PlaylistTrack{
				{
					Key:          "01",
					Title:        "Bird Explorer's Guide",
					TrackURL:     fmt.Sprintf("yoto:#%s", descriptionSha),
					Duration:     descriptionInfo.GetDuration(),
					FileSize:     descriptionInfo.GetFileSize(),
					Channels:     descriptionInfo.GetChannels(),
					Format:       descriptionInfo.Transcode.TranscodedInfo.Format,
					Type:         "audio",
					OverlayLabel: descOverlayLabel,
					Display: Display{
						Icon16x16: bookIcon,
					},
				},
			},
			Display: Display{
				Icon16x16: bookIcon,
			},
		})
	}

	// Log chapter count for debugging
	fmt.Printf("Building content update: %d chapters (hasAnnouncement: %v, hasDescription: %v, hasOutro: %v)\n",
		len(chapters), hasAnnouncement, hasDescription, hasOutro)

	// Add outro chapter if generated successfully
	if hasOutro {
		outroChapterKey := "04"
		outroOverlayLabel := "4"
		if hasAnnouncement && hasDescription {
			outroChapterKey = "05"
			outroOverlayLabel = "5"
		} else if hasAnnouncement || hasDescription {
			outroChapterKey = "04"
			outroOverlayLabel = "4"
		} else {
			outroChapterKey = "03"
			outroOverlayLabel = "3"
		}

		chapters = append(chapters, Chapter{
			Key:          outroChapterKey,
			Title:        "See You Tomorrow!",
			OverlayLabel: outroOverlayLabel,
			Tracks: []PlaylistTrack{
				{
					Key:          "01",
					Title:        "See You Tomorrow, Explorers!",
					TrackURL:     fmt.Sprintf("yoto:#%s", outroSha),
					Duration:     outroInfo.GetDuration(),
					FileSize:     outroInfo.GetFileSize(),
					Channels:     outroInfo.GetChannels(),
					Format:       outroInfo.Transcode.TranscodedInfo.Format,
					Type:         "audio",
					OverlayLabel: outroOverlayLabel,
					Display: Display{
						Icon16x16: hikingBootIcon,
					},
				},
			},
			Display: Display{
				Icon16x16: hikingBootIcon,
			},
		})
	}

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

	metadata := Metadata{
		Media: MediaInfo{
			Duration:         totalDuration,
			FileSize:         totalSize,
			ReadableFileSize: float64(totalSize) / 1024 / 1024,
		},
	}

	if existingCard != nil && existingCard.Metadata != nil {
		if cover, hasCover := existingCard.Metadata["cover"]; hasCover {
			metadataMap := map[string]interface{}{
				"media": map[string]interface{}{
					"duration":         totalDuration,
					"fileSize":         totalSize,
					"readableFileSize": float64(totalSize) / 1024 / 1024,
				},
				"cover": cover, // Preserve the existing cover
			}

			updateReq := map[string]interface{}{
				"cardId":            cardID, // Use cardId for POST requests
				"userId":            userID,
				"createdByClientId": clientID,
				"title":             "Bird Song Explorer", // Keep original title
				"content": map[string]interface{}{
					"chapters": chapters,
				},
				"metadata":  metadataMap,
				"createdAt": createdAt,
				"updatedAt": now,
			}

			return cm.sendUpdateRequest(cardID, updateReq, birdName)
		}
	}

	updateReq := UpdateContentRequest{
		CardID:            cardID, // Include CardID for POST requests
		UserID:            userID,
		CreatedByClientID: clientID,
		Title:             "Bird Song Explorer", // Keep original title
		Content: Content{
			Chapters: chapters,
		},
		Metadata:  metadata,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}

	return cm.sendUpdateRequest(cardID, updateReq, birdName)
}

// sendUpdateRequest sends the update request to the Yoto API
func (cm *ContentManager) sendUpdateRequest(cardID string, updateReq interface{}, birdName string) error {
	// Use POST request to /content for updating existing content
	url := fmt.Sprintf("%s/content", cm.client.baseURL)

	jsonData, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cm.client.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cm.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Log success for debugging
	fmt.Printf("Content update response (%d): Updated %s, body length: %d\n", resp.StatusCode, birdName, len(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to update content: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// generateBirdAnnouncement creates a short audio announcing the bird
// Returns the audio data directly instead of a URL
func (cm *ContentManager) generateBirdAnnouncement(birdName string, voiceID string) ([]byte, error) {
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		return nil, fmt.Errorf("no ElevenLabs API key configured")
	}

	// Voice ID must be provided by the caller
	if voiceID == "" {
		return nil, fmt.Errorf("voice ID is required for announcement generation")
	}

	// Check if we have ambience data from Track 1 (enhanced intro)
	if cm.selectedAmbience != "" && len(cm.ambienceData) > 0 {
		// Use enhanced announcement with continuing ambience
		enhancedAnnouncement := services.NewEnhancedBirdAnnouncement(elevenLabsKey)
		announcementData, err := enhancedAnnouncement.GenerateAnnouncementFromAudioData(
			birdName,
			voiceID,
			cm.ambienceData,
		)
		if err == nil {
			// Store the announcement text for transitions (without breaks)
			cm.lastAnnouncementText = fmt.Sprintf("Today's bird is the %s! Listen carefully to its unique song.", birdName)
			return announcementData, nil
		}
		// Fall through to standard generation if enhanced fails
	}

	// Add pauses between sentences using ElevenLabs break syntax
	// The <break time="1.0s" /> syntax creates natural pauses that the AI understands
	announcement := fmt.Sprintf("Today's bird is the %s! <break time=\"1.0s\" /> Listen carefully to its unique song.", birdName)

	// Generate speech using ElevenLabs
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     announcement,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.30, // Low for good emotional range while maintaining stability
			"similarity_boost": 0.95, // Very high similarity to original voice
			"speed":            0.95, // Faster, more energetic pace
		},
		"previous_text": cm.lastIntroText,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", elevenLabsKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ElevenLabs API error: %d - %s", resp.StatusCode, string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Store the announcement text for the next track
	cm.lastAnnouncementText = announcement

	return audioData, nil
}

// generateOutro uses pre-recorded outro files with ambient sounds
func (cm *ContentManager) generateOutro(birdName string, voiceID string) ([]byte, error) {
	if voiceID == "" {
		return nil, fmt.Errorf("voice ID is required for outro generation")
	}

	now := time.Now()
	dayOfWeek := now.Weekday()

	// Get voice name from ID
	voiceManager := config.NewVoiceManager()
	voiceName := ""
	for _, voice := range voiceManager.GetAvailableVoices() {
		if voice.ID == voiceID {
			voiceName = voice.Name
			break
		}
	}
	if voiceName == "" {
		return nil, fmt.Errorf("voice not found for ID: %s", voiceID)
	}

	fmt.Printf("Using pre-recorded outro for %s on %s\n", voiceName, dayOfWeek.String())

	// Use OutroIntegration to get pre-recorded outro with volume boost
	outroIntegration := services.NewOutroIntegration()
	return outroIntegration.GenerateOutroWithAmbience(voiceName, dayOfWeek, cm.ambienceData, "")
}

// generateBirdDescription creates audio narration for the bird's Wikipedia description
// Uses the same voice as introduction for consistency
func (cm *ContentManager) generateBirdDescription(description string, birdName string, voiceID string) ([]byte, error) {
	// Check if we have ElevenLabs API key
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		return nil, fmt.Errorf("no ElevenLabs API key configured")
	}

	// Voice ID must be provided by the caller
	if voiceID == "" {
		return nil, fmt.Errorf("voice ID is required for description generation")
	}

	// Extract components from the Wikipedia description
	// The first sentence often contains the scientific name in a format like:
	// "The ring-necked duck (Aythya collaris) is a diving duck..."
	scientificName := ""
	simpleFact := description
	
	// Look for scientific name in parentheses
	if strings.Contains(description, "(") && strings.Contains(description, ")") {
		start := strings.Index(description, "(")
		end := strings.Index(description, ")")
		if start < end && end-start < 50 { // Scientific names are usually short
			potentialName := description[start+1:end]
			// Check if it looks like a scientific name (two words, capitalized)
			words := strings.Fields(potentialName)
			if len(words) == 2 && strings.Title(words[0]) == words[0] {
				scientificName = potentialName
				// Remove the scientific name from the description for the fact
				simpleFact = description[:start] + description[end+1:]
			}
		}
	}
	
	// Clean up and simplify the fact
	simpleFact = strings.TrimSpace(simpleFact)
	parts := strings.Split(simpleFact, ".")
	if len(parts) > 0 {
		// Use the first sentence about what the bird IS
		if strings.Contains(strings.ToLower(parts[0]), "is a") {
			simpleFact = strings.TrimSpace(parts[0]) + "."
		} else if len(parts) > 1 {
			simpleFact = strings.TrimSpace(parts[0]) + "." + strings.TrimSpace(parts[1]) + "."
		}
	}
	
	// Remove overly technical content
	if strings.Contains(strings.ToLower(simpleFact), "derived from") ||
	   strings.Contains(strings.ToLower(simpleFact), "greek") ||
	   strings.Contains(strings.ToLower(simpleFact), "latin") {
		// Just say it's a type of bird if too technical
		simpleFact = fmt.Sprintf("The %s is an amazing bird!", birdName)
	}
	
	// Build the final text with the requested format
	var descriptionText string
	if scientificName != "" {
		// Format: "[Scientific name]. [Location/sightings]. Did you know? [Fact] Isn't that amazing?"
		descriptionText = fmt.Sprintf("The scientific name for the %s is %s. There have been recent sightings in your area! Did you know? %s Isn't that amazing? Nature is full of wonderful surprises!", 
			birdName, scientificName, simpleFact)
	} else {
		// If no scientific name found, still mention sightings
		descriptionText = fmt.Sprintf("There have been recent sightings of %s in your area! Did you know? %s Isn't that amazing? Nature is full of wonderful surprises!", 
			birdName, simpleFact)
	}
	
	// Log the text that will be spoken
	fmt.Printf("[BIRD_DESCRIPTION] Track 4 text: %s\n", descriptionText)

	// Generate speech using ElevenLabs
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     descriptionText,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.30, // Low for good emotional range while maintaining stability
			"similarity_boost": 0.95, // Very high similarity to original voice
			"speed":            0.95, // Faster, more energetic pace
		},
		// No previous_text since Track 3 (bird song) is between Track 2 and 4
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", elevenLabsKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ElevenLabs API error: %d - %s", resp.StatusCode, string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Store the description text for the next track (outro)
	cm.lastDescriptionText = descriptionText

	return audioData, nil
}

// generateEnhancedBirdDescription creates location-aware audio narration using V4 fact generator
func (cm *ContentManager) generateEnhancedBirdDescription(description string, birdName string, voiceID string, latitude, longitude float64) ([]byte, error) {
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		// Fall back to regular description if no TTS available
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	ebirdAPIKey := os.Getenv("EBIRD_API_KEY")
	if ebirdAPIKey == "" {
		// Fall back to regular description if no eBird API
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	factGen := services.NewImprovedFactGeneratorV4(ebirdAPIKey)

	bird := &models.Bird{
		CommonName:     birdName,
		ScientificName: "", // Could be extracted from Wikipedia if needed
		Family:         "",
		AudioURL:       "", // Not needed for description generation
		Description:    description,
	}

	enhancedScript := factGen.GenerateExplorersGuideScriptWithLocation(bird, latitude, longitude)
	fmt.Printf("[ENHANCED_FACTS] Generated script: %d characters\n", len(enhancedScript))

	// If the script is too short or empty, fall back to regular
	if len(enhancedScript) < 100 {
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	// Limit script length to prevent excessive TTS costs
	maxLength := 2500
	if len(enhancedScript) > maxLength {
		enhancedScript = enhancedScript[:maxLength]
		// Find the last complete sentence
		lastPeriod := strings.LastIndex(enhancedScript, ".")
		if lastPeriod > 0 {
			enhancedScript = enhancedScript[:lastPeriod+1]
		}
	}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     enhancedScript,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.30, // Low for good emotional range while maintaining stability
			"similarity_boost": 0.95, // Very high similarity to original voice
			"speed":            0.95, // Faster, more energetic pace
		},
		// No previous_text since Track 3 (bird song) is between Track 2 and 4
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("[ENHANCED_FACTS] Failed to marshal request: %v\n", err)
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[ENHANCED_FACTS] Failed to create request: %v\n", err)
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", elevenLabsKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ENHANCED_FACTS] TTS request failed: %v\n", err)
		return cm.generateBirdDescription(description, birdName, voiceID)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[ENHANCED_FACTS] TTS API error: %d\n", resp.StatusCode)
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[ENHANCED_FACTS] Failed to read audio data: %v\n", err)
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	return audioData, nil
}

// extractIntroTextFromURL extracts the intro text from a pre-recorded intro URL
// This maps the intro filename to the actual text used in the recording
func (cm *ContentManager) extractIntroTextFromURL(introURL string) {
	// Map of intro texts used in the pre-recorded files
	introTexts := []string{
		"Welcome, nature detectives! Time to discover an amazing bird from your neighborhood.",
		"Hello, bird explorers! Today's special bird is waiting to sing for you.",
		"Ready for an adventure? Let's meet today's featured bird from your area!",
		"Welcome back, little listeners! A wonderful bird is calling just for you.",
		"Hello, young scientists! Let's explore the amazing birds living near you.",
		"Calling all bird lovers! Your daily bird discovery awaits.",
		"Time for today's bird adventure! Listen closely to nature's music.",
		"Welcome to your daily bird journey! Let's discover who's singing today.",
	}

	// Extract the intro number from the URL (e.g., intro_00_Amelia.mp3 -> 0)
	if strings.Contains(introURL, "/intro_") {
		// Find the intro number in the filename
		for i := 0; i < len(introTexts); i++ {
			introPattern := fmt.Sprintf("intro_%02d_", i)
			if strings.Contains(introURL, introPattern) {
				cm.lastIntroText = introTexts[i]
				return
			}
		}
	}

	// If not a pre-recorded intro or pattern not found, use a default
	cm.lastIntroText = "Welcome to Bird Song Explorer!"
}

// captureAmbienceFromEnhancedIntro checks if the intro is enhanced and captures ambience data
func (cm *ContentManager) captureAmbienceFromEnhancedIntro(introURL string) {
	// Check if this is an enhanced intro (from cache)
	if !strings.Contains(introURL, "/audio/cache/enhanced_intros/") {
		// Not an enhanced intro, clear ambience data
		cm.selectedAmbience = ""
		cm.ambienceData = nil
		return
	}

	// Try to get the ambience data from the audio manager
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		return
	}

	// Create an enhanced intro mixer to get ambience info
	mixer := services.NewEnhancedIntroMixer(elevenLabsKey)

	// Determine which ambience was selected based on the day (same logic as intro generation)
	ambiences := mixer.GetAvailableAmbiences()
	now := time.Now()
	daySeed := now.Year()*10000 + int(now.Month())*100 + now.Day()
	selectedAmbienceIdx := daySeed % len(ambiences)

	if selectedAmbienceIdx < len(ambiences) {
		selectedAmbience := ambiences[selectedAmbienceIdx]
		cm.selectedAmbience = selectedAmbience.Name

		// Try to read the ambience file
		ambiencePath := filepath.Join("sound_effects", selectedAmbience.Path)
		if data, err := os.ReadFile(ambiencePath); err == nil {
			cm.ambienceData = data
			fmt.Printf("[CONTENT_UPDATE] Captured %s ambience for Track 2 (%d bytes)\n",
				cm.selectedAmbience, len(cm.ambienceData))
		}
	}
}

// addFadeInToAudio adds a fade-in effect to audio data
func (cm *ContentManager) addFadeInToAudio(audioData []byte, fadeSeconds float64) ([]byte, error) {
	// Check if ffmpeg is available
	cmd := exec.Command("which", "ffmpeg")
	if err := cmd.Run(); err != nil {
		// ffmpeg not available, return original audio
		return audioData, nil
	}

	// Create temp files
	tempDir := os.TempDir()
	inputFile := filepath.Join(tempDir, fmt.Sprintf("input_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.mp3", time.Now().Unix()))

	// Write input data
	if err := os.WriteFile(inputFile, audioData, 0644); err != nil {
		return audioData, nil
	}
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// Apply fade-in using ffmpeg
	ffmpegCmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-af", fmt.Sprintf("afade=t=in:st=0:d=%.1f", fadeSeconds),
		"-c:a", "libmp3lame",
		"-b:a", "192k",
		"-ar", "44100",
		"-y",
		outputFile,
	)

	var stderr bytes.Buffer
	ffmpegCmd.Stderr = &stderr

	if err := ffmpegCmd.Run(); err != nil {
		fmt.Printf("[CONTENT_UPDATE] Failed to add fade-in: %v\n", err)
		return audioData, nil
	}

	// Read the output
	fadedData, err := os.ReadFile(outputFile)
	if err != nil {
		return audioData, nil
	}

	fmt.Printf("[CONTENT_UPDATE] Added %.1fs fade-in to audio\n", fadeSeconds)
	return fadedData, nil
}
