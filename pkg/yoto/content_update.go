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
	"github.com/callen/bird-song-explorer/pkg/ebird"
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
	// Call the location-aware version with zero coordinates to indicate no location available
	// The zero coordinates will trigger the no-location fallback behavior
	return cm.UpdateExistingCardContentWithDescriptionVoiceAndLocation(cardID, birdName, introURL, birdSongURL, birdDescription, voiceID, 0, 0)
}

// UpdateExistingCardContentWithDescriptionVoiceAndLocation updates an existing MYO card with location-aware content
func (cm *ContentManager) UpdateExistingCardContentWithDescriptionVoiceAndLocation(cardID string, birdName string, introURL string, birdSongURL string, birdDescription string, voiceID string, latitude, longitude float64) error {
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
	
	// Log Track 1 intro information
	if strings.Contains(introURL, "/intro_") {
		parts := strings.Split(introURL, "/")
		if len(parts) > 0 {
			fileName := parts[len(parts)-1]
			fmt.Printf("[TRACK_1_INTRO] Using pre-recorded intro file: %s\n", fileName)
		}
	} else if strings.Contains(introURL, "enhanced_intros") {
		fmt.Printf("[TRACK_1_INTRO] Using enhanced intro from cache\n")
	} else {
		fmt.Printf("[TRACK_1_INTRO] Using intro from URL: %s\n", introURL)
	}

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

	var descriptionSha string
	var descriptionInfo *TranscodeResponse
	var hasDescription bool

	if birdDescription != "" {
		var descriptionData []byte
		var err error

		// Check if we have valid location coordinates
		hasValidLocation := latitude != 0 || longitude != 0

		if hasValidLocation {
			locationName := cm.getLocationName(latitude, longitude)
			if locationName != "" {
				fmt.Printf("[CONTENT_UPDATE] Location available (%s at lat:%f, lng:%f) - using location-aware facts\n", 
					locationName, latitude, longitude)
			} else {
				fmt.Printf("[CONTENT_UPDATE] Location coordinates available (lat:%f, lng:%f) but couldn't determine city/state - using generic facts\n",
					latitude, longitude)
			}
		} else {
			fmt.Printf("[CONTENT_UPDATE] No location available (lat:%f, lng:%f) - using generic bird facts without sighting claims\n",
				latitude, longitude)
		}

		// Choose generator based on environment variable
		generatorType := os.Getenv("BIRD_FACT_GENERATOR")
		if generatorType == "" {
			generatorType = "basic" // Default to basic generator
		}
		
		fmt.Printf("[CONTENT_UPDATE] Using %s facts generator\n", generatorType)
		
		if generatorType == "enhanced" {
			// Use enhanced generator (formerly V4)
			descriptionData, err = cm.generateEnhancedBirdDescriptionModular(birdDescription, birdName, voiceID, latitude, longitude)
		} else {
			// Use basic generator (standard)
			descriptionData, err = cm.generateBirdDescriptionWithLocation(birdDescription, birdName, voiceID, latitude, longitude)
		}

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
	
	// Log Track 2 voice information
	voiceManager := config.NewVoiceManager()
	voiceName := "Unknown"
	for _, voice := range voiceManager.GetAvailableVoices() {
		if voice.ID == voiceID {
			voiceName = voice.Name
			break
		}
	}
	fmt.Printf("[TRACK_2_ANNOUNCEMENT] Using voice: %s (ID: %s) for bird announcement\n", voiceName, voiceID)

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
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":         0.40,
			"similarity_boost":  0.90,
			"use_speaker_boost": true,
			"speed":             1.0,
			"style":             0,
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

	fmt.Printf("[TRACK_5_OUTRO] Using pre-recorded outro for voice: %s (ID: %s) on %s\n", voiceName, voiceID, dayOfWeek.String())

	// Use OutroIntegration to get pre-recorded outro with volume boost
	outroIntegration := services.NewOutroIntegration()
	return outroIntegration.GenerateOutroWithAmbience(voiceName, dayOfWeek, cm.ambienceData, "")
}

// getLocationName converts coordinates to a kid-friendly location name
func (cm *ContentManager) getLocationName(latitude, longitude float64) string {
	ebirdAPIKey := os.Getenv("EBIRD_API_KEY")
	if ebirdAPIKey == "" {
		return ""
	}

	ebirdClient := ebird.NewClient(ebirdAPIKey)
	hotspots, err := ebirdClient.GetNearbyHotspots(latitude, longitude, 25)
	if err != nil || len(hotspots) == 0 {
		return ""
	}

	cities := make(map[string]int)
	states := make(map[string]int)

	for _, hotspot := range hotspots {
		parts := strings.Split(hotspot.LocationName, ", ")
		if len(parts) >= 2 {
			// Last part is usually state/province/country
			statePart := strings.TrimSpace(parts[len(parts)-1])
			if isValidLocationName(statePart) {
				states[statePart]++
			}

			// Second to last is usually city
			if len(parts) >= 2 {
				cityPart := strings.TrimSpace(parts[len(parts)-2])
				// Skip if it looks like a street address
				if !strings.Contains(cityPart, "St") && !strings.Contains(cityPart, "Ave") &&
					!strings.Contains(cityPart, "Rd") && !containsNumbers(cityPart) {
					cities[cityPart]++
				}
			}
		}
	}

	// Get most common city and state
	city := getMostCommonLocation(cities)
	state := getMostCommonLocation(states)

	if city != "" && state != "" {
		return fmt.Sprintf("%s, %s", city, state)
	} else if state != "" {
		return state
	}

	return ""
}

func isValidLocationName(s string) bool {
	s = strings.TrimSpace(s)

	validUSStates := []string{"Oregon", "Washington", "California", "Nevada", "Idaho",
		"Arizona", "Utah", "Colorado", "Montana", "Wyoming", "New Mexico", "Texas",
		"Alaska", "Hawaii", "Florida", "Georgia", "North Carolina", "South Carolina",
		"Virginia", "West Virginia", "Maryland", "Delaware", "New Jersey", "New York",
		"Connecticut", "Rhode Island", "Massachusetts", "Vermont", "New Hampshire", "Maine",
		"Pennsylvania", "Ohio", "Indiana", "Illinois", "Michigan", "Wisconsin", "Minnesota",
		"Iowa", "Missouri", "Arkansas", "Louisiana", "Mississippi", "Alabama", "Tennessee",
		"Kentucky", "Kansas", "Nebraska", "Oklahoma", "North Dakota", "South Dakota"}

	for _, state := range validUSStates {
		if strings.EqualFold(s, state) {
			return true
		}
	}

	if len(s) == 2 {
		s = strings.ToUpper(s)
		validAbbr := []string{"OR", "WA", "CA", "NV", "ID", "AZ", "UT", "CO", "MT", "WY",
			"NM", "TX", "AK", "HI", "FL", "GA", "NC", "SC", "VA", "WV", "MD", "DE", "NJ",
			"NY", "CT", "RI", "MA", "VT", "NH", "ME", "PA", "OH", "IN", "IL", "MI", "WI",
			"MN", "IA", "MO", "AR", "LA", "MS", "AL", "TN", "KY", "KS", "NE", "OK", "ND", "SD", "DC"}
		for _, abbr := range validAbbr {
			if s == abbr {
				return true
			}
		}
	}

	canadianRegions := []string{
		"Ontario", "Quebec", "British Columbia", "Alberta", "Manitoba", "Saskatchewan",
		"Nova Scotia", "New Brunswick", "Newfoundland and Labrador", "Prince Edward Island",
		"Northwest Territories", "Yukon", "Nunavut",
		// Common abbreviations
		"ON", "QC", "BC", "AB", "MB", "SK", "NS", "NB", "NL", "PE", "NT", "YT", "NU",
	}

	for _, region := range canadianRegions {
		if strings.EqualFold(s, region) {
			return true
		}
	}

	countries := []string{
		"Canada", "United Kingdom", "UK", "England", "Scotland", "Wales", "Northern Ireland",
		"Australia", "New Zealand", "Ireland", "South Africa", "India", "Singapore",
		"Mexico", "Costa Rica", "Panama", "Colombia", "Ecuador", "Peru", "Brazil", "Argentina",
		"Kenya", "Tanzania", "Uganda", "South Africa", "Botswana", "Namibia",
		"Japan", "Thailand", "Malaysia", "Indonesia", "Philippines",
		"Spain", "France", "Germany", "Italy", "Netherlands", "Belgium", "Switzerland",
		"Norway", "Sweden", "Finland", "Denmark", "Iceland",
	}

	for _, country := range countries {
		if strings.EqualFold(s, country) {
			return true
		}
	}

	australianRegions := []string{
		"New South Wales", "Victoria", "Queensland", "Western Australia",
		"South Australia", "Tasmania", "Australian Capital Territory", "Northern Territory",
		"NSW", "VIC", "QLD", "WA", "SA", "TAS", "ACT", "NT",
	}

	for _, region := range australianRegions {
		if strings.EqualFold(s, region) {
			return true
		}
	}

	ukRegions := []string{
		"Greater London", "South East", "South West", "West Midlands", "North West",
		"North East", "Yorkshire and the Humber", "East Midlands", "East of England",
	}

	for _, region := range ukRegions {
		if strings.EqualFold(s, region) {
			return true
		}
	}

	locationKeywords := []string{"Province", "Prefecture", "State", "Territory", "Region", "Department"}
	for _, keyword := range locationKeywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}

	if len(s) > 2 && !containsNumbers(s) && !strings.Contains(s, "/") && !strings.Contains(s, "\\") {
		if len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z' {
			return true
		}
	}

	return false
}

func containsNumbers(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func getMostCommonLocation(locations map[string]int) string {
	maxCount := 0
	result := ""
	for loc, count := range locations {
		if count > maxCount {
			maxCount = count
			result = loc
		}
	}
	return result
}

// getGenericBirdFact returns an interesting generic fact based on bird characteristics
func getGenericBirdFact(birdName string, existingFact string) string {
	lowerName := strings.ToLower(birdName)

	// Size-based facts
	if strings.Contains(lowerName, "eagle") || strings.Contains(lowerName, "hawk") ||
		strings.Contains(lowerName, "owl") {
		return "Birds of prey have incredible eyesight - they can spot tiny movements from far away!"
	}

	if strings.Contains(lowerName, "hummingbird") {
		return "Hummingbirds are the only birds that can fly backwards and their hearts beat over one thousand two-hundred times per minute!"
	}

	// Water birds
	if strings.Contains(lowerName, "duck") || strings.Contains(lowerName, "goose") ||
		strings.Contains(lowerName, "swan") {
		return "Water birds have special oil glands that keep their feathers waterproof!"
	}

	// Songbirds
	if strings.Contains(lowerName, "robin") || strings.Contains(lowerName, "sparrow") ||
		strings.Contains(lowerName, "finch") || strings.Contains(lowerName, "warbler") {
		return "Songbirds learn their songs by listening to their parents, just like you learned to talk!"
	}

	// Colorful birds
	if strings.Contains(lowerName, "cardinal") || strings.Contains(lowerName, "blue") ||
		strings.Contains(lowerName, "gold") {
		return "Bright colors help birds recognize their own species and attract mates!"
	}

	// Nocturnal birds
	if strings.Contains(lowerName, "owl") || strings.Contains(lowerName, "nightjar") {
		return "Night birds have special feathers that let them fly almost silently!"
	}

	// Migration-related
	if strings.Contains(lowerName, "swallow") || strings.Contains(lowerName, "crane") ||
		strings.Contains(lowerName, "arctic") {
		return "Some birds travel thousands of miles each year, using the stars and Earth's magnetic field to navigate!"
	}

	// Intelligence
	if strings.Contains(lowerName, "crow") || strings.Contains(lowerName, "raven") ||
		strings.Contains(lowerName, "jay") {
		return "These birds are super smart - they can use tools and even recognize human faces!"
	}

	// Tropical birds
	if strings.Contains(lowerName, "parrot") || strings.Contains(lowerName, "toucan") ||
		strings.Contains(lowerName, "macaw") {
		return "Tropical birds often live in the rainforest canopy, rarely coming down to the ground!"
	}

	// Shore birds
	if strings.Contains(lowerName, "gull") || strings.Contains(lowerName, "tern") ||
		strings.Contains(lowerName, "sandpiper") {
		return "Shore birds have long legs and beaks perfect for finding food in sand and shallow water!"
	}

	// Default interesting facts
	defaultFacts := []string{
		"Birds are the only animals with feathers - no other creature has them!",
		"A bird's bones are hollow, making them light enough to fly!",
		"Birds can see colors that humans can't even imagine!",
		"Most birds have excellent memories and can remember hundreds of food hiding spots!",
		"Baby birds can eat their own body weight in food every day!",
		"Birds help plants grow by spreading seeds in their droppings!",
		"Some birds can sleep with one half of their brain while the other stays awake!",
		"Birds existed alongside dinosaurs - they're living dinosaurs themselves!",
	}

	// Pick a random default fact
	rand.Seed(time.Now().UnixNano())
	return defaultFacts[rand.Intn(len(defaultFacts))]
}

// generateBirdDescriptionWithLocation creates audio narration with location awareness
func (cm *ContentManager) generateBirdDescriptionWithLocation(description string, birdName string, voiceID string, latitude, longitude float64) ([]byte, error) {
	hasValidLocation := latitude != 0 || longitude != 0

	if hasValidLocation {
		// Use location-aware description
		return cm.generateBirdDescriptionWithSightings(description, birdName, voiceID, latitude, longitude)
	} else {
		// Use generic description without sighting claims
		return cm.generateBirdDescription(description, birdName, voiceID)
	}
}

// generateBirdDescriptionWithSightings creates audio narration with location-specific sighting information
func (cm *ContentManager) generateBirdDescriptionWithSightings(description string, birdName string, voiceID string, latitude, longitude float64) ([]byte, error) {
	// Check if we have ElevenLabs API key
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		return nil, fmt.Errorf("no ElevenLabs API key configured")
	}

	// Voice ID must be provided by the caller
	if voiceID == "" {
		return nil, fmt.Errorf("voice ID is required for description generation")
	}
	
	// Log Track 4 voice information for location-aware generator
	voiceManager := config.NewVoiceManager()
	voiceName := "Unknown"
	for _, voice := range voiceManager.GetAvailableVoices() {
		if voice.ID == voiceID {
			voiceName = voice.Name
			break
		}
	}
	fmt.Printf("[TRACK_4_DESCRIPTION] Using voice: %s (ID: %s) for bird description (location-aware)\n", voiceName, voiceID)

	// Extract components from the Wikipedia description
	scientificName := ""
	simpleFact := description

	// Look for scientific name in parentheses
	if strings.Contains(description, "(") && strings.Contains(description, ")") {
		start := strings.Index(description, "(")
		end := strings.Index(description, ")")
		if start < end && end-start < 50 { // Scientific names are usually short
			potentialName := description[start+1 : end]
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
			simpleFact = strings.TrimSpace(parts[0]) + ". " + strings.TrimSpace(parts[1]) + "."
		}
	}

	// Remove overly technical content
	if strings.Contains(strings.ToLower(simpleFact), "derived from") ||
		strings.Contains(strings.ToLower(simpleFact), "greek") ||
		strings.Contains(strings.ToLower(simpleFact), "latin") {
		// Just say it's a type of bird if too technical
		simpleFact = fmt.Sprintf("The %s is an amazing bird!", birdName)
	}

	// Build the final text WITH location claims (for valid location case)
	var descriptionText string

	// Get location name for kid-friendly references
	locationName := cm.getLocationName(latitude, longitude)
	locationReference := "your area"
	if locationName != "" {
		// Use the actual location name instead of "your area"
		locationReference = locationName
	}

	// Try to get actual sighting data if we have eBird API key
	ebirdAPIKey := os.Getenv("EBIRD_API_KEY")
	hasSightings := false

	if ebirdAPIKey != "" {
		// Check for actual sightings using eBird API
		ebirdClient := ebird.NewClient(ebirdAPIKey)
		observations, err := ebirdClient.GetRecentObservations(latitude, longitude, 30)
		if err == nil {
			// Check if this specific bird has been seen
			for _, obs := range observations {
				if strings.EqualFold(obs.CommonName, birdName) {
					hasSightings = true
					if locationName != "" {
						fmt.Printf("[BIRD_DESCRIPTION] Verified sighting of %s in %s\n", birdName, locationName)
					} else {
						fmt.Printf("[BIRD_DESCRIPTION] Verified sighting of %s at provided location\n", birdName)
					}
					break
				}
			}
		}
	}

	// Use random selection for variety
	rand.Seed(time.Now().UnixNano())

	if hasSightings {
		// We have verified sightings - use location name in the text
		if scientificName != "" {
			templates := []string{
				"The %s has a special scientific name: %s. Recently, people spotted them in %s! %s Isn't that cool? Maybe you'll see one today!",
				"Scientists call the %s by its special name: %s. Guess what? Bird watchers have spotted them in %s recently! Here's a feathered fact: %s How exciting is that?",
				"The %s goes by the scientific name %s. Amazing news - they've been seen in %s! Did you know? %s Nature never stops surprising us!",
				"The %s is called %s in science books. It was spotted in %s not too long ago! What a discovery! %s",
				"Meet the %s, scientifically known as %s! Good news for explorers: there have been sightings in %s. Fun fact: %s Isn't nature incredible?",
			}
			descriptionText = fmt.Sprintf(templates[rand.Intn(len(templates))],
				birdName, scientificName, locationReference, simpleFact)
		} else {
			templates := []string{
				"Bird watchers have spotted the %s in %s! Did you know? %s Isn't that cool? Nature is full of wonderful surprises!",
				"Exciting news! %s has been spotted in %s recently! Here's something cool: %s",
				"Bird alert! The %s has been seen flying in %s! Fun fact for you: %s Nature is so fascinating!",
				"Guess what? %s are active in %s right now! Did you know? %s What an incredible bird!",
			}
			descriptionText = fmt.Sprintf(templates[rand.Intn(len(templates))],
				birdName, locationReference, simpleFact)
		}
	} else {
		{
			fmt.Printf("[BIRD_DESCRIPTION] No verified sightings at provided location - using generic facts\n")
			// Fall back to generic text without location references
			if scientificName != "" {
				templates := []string{
					"Want to sound like a scientist? The %s is called %s. Pretty cool, right? %s",
					"The fancy science name for the %s is %s. Try saying it out loud! Here's a feathered fact: %s",
					"The %s has a special scientific name: %s. Fun fact: %s. Nature is so fascinating!",
					"Here's something amazing: the %s is called %s in science language. Did you know? %s",
					"Scientists call the %s by the name %s. Here's a cool fact: %s Nature is full of surprises!",
					"The %s is scientifically known as %s. Fun fact: %s Keep exploring - birds are all around us!",
				}
				descriptionText = fmt.Sprintf(templates[rand.Intn(len(templates))],
					birdName, scientificName, simpleFact)
			} else {
				templates := []string{
					"The %s is one of nature's wonders. Here's why: %s Birds are everywhere, waiting to be discovered!",
					"Every bird has a story. The %s' story goes like this: %s Isn't that cool?",
					"Let's learn about the %s! Here's something special: %s Nature has so many treasures to find!",
					"Discover the wonderful %s! Fun fact: %s Keep your eyes and ears open for bird adventures!",
					"The %s is more than just a bird-it's a clue to how nature works. Did you know? %s Every bird has its own special story!",
				}
				descriptionText = fmt.Sprintf(templates[rand.Intn(len(templates))],
					birdName, simpleFact)
			}
		}
	}

	// Log the text that will be spoken
	fmt.Printf("[BIRD_DESCRIPTION] Track 4 text (location-aware): %s\n", descriptionText)

	// Generate speech using ElevenLabs
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     descriptionText,
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":         0.40,
			"similarity_boost":  0.90,
			"use_speaker_boost": true,
			"speed":             1.0,
			"style":             0,
		},
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
	
	// Log Track 4 voice information for basic generator
	voiceManager := config.NewVoiceManager()
	voiceName := "Unknown"
	for _, voice := range voiceManager.GetAvailableVoices() {
		if voice.ID == voiceID {
			voiceName = voice.Name
			break
		}
	}
	fmt.Printf("[TRACK_4_DESCRIPTION] Using voice: %s (ID: %s) for bird description (basic generator)\n", voiceName, voiceID)

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
			potentialName := description[start+1 : end]
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
			simpleFact = strings.TrimSpace(parts[0]) + ". " + strings.TrimSpace(parts[1]) + "."
		}
	}

	// Remove overly technical content
	if strings.Contains(strings.ToLower(simpleFact), "derived from") ||
		strings.Contains(strings.ToLower(simpleFact), "greek") ||
		strings.Contains(strings.ToLower(simpleFact), "latin") {
		// Just say it's a type of bird if too technical
		simpleFact = fmt.Sprintf("The %s is an amazing bird!", birdName)
	}

	// Build enhanced generic text for global audience
	var descriptionText string

	additionalFact := getGenericBirdFact(birdName, simpleFact)

	if scientificName != "" {
		// Format with scientific name and enhanced fact
		descriptionText = fmt.Sprintf("The scientific name for the %s is %s. Did you know? %s %s Birds are found all over the world, each one perfectly adapted to its home!",
			birdName, scientificName, simpleFact, additionalFact)
	} else {
		// Enhanced version without scientific name
		descriptionText = fmt.Sprintf("Let me tell you about the amazing %s! Did you know? %s %s Every bird has its own special story. Listen carefully to learn its unique song!",
			birdName, simpleFact, additionalFact)
	}

	// Log the text that will be spoken
	fmt.Printf("[BIRD_DESCRIPTION] Track 4 text: %s\n", descriptionText)

	// Generate speech using ElevenLabs
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     descriptionText,
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":         0.40,
			"similarity_boost":  0.90,
			"use_speaker_boost": true,
			"speed":             1.0,
			"style":             0,
		},
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

// generateEnhancedBirdDescriptionModular creates location-aware audio narration using the modular enhanced fact generator
func (cm *ContentManager) generateEnhancedBirdDescriptionModular(description string, birdName string, voiceID string, latitude, longitude float64) ([]byte, error) {
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		// Fall back to basic description if no TTS available
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	ebirdAPIKey := os.Getenv("EBIRD_API_KEY")
	if ebirdAPIKey == "" {
		// Fall back to basic description if no eBird API
		return cm.generateBirdDescription(description, birdName, voiceID)
	}

	// Log Track 4 voice information
	voiceManager := config.NewVoiceManager()
	voiceName := "Unknown"
	for _, voice := range voiceManager.GetAvailableVoices() {
		if voice.ID == voiceID {
			voiceName = voice.Name
			break
		}
	}
	fmt.Printf("[TRACK_4_DESCRIPTION] Using voice: %s (ID: %s) for bird description\n", voiceName, voiceID)
	
	// Use the modular fact generator interface
	factGen := services.NewFactGenerator("enhanced", ebirdAPIKey)
	
	bird := &models.Bird{
		CommonName:     birdName,
		ScientificName: "", // Could be extracted from Wikipedia if needed
		Family:         "",
		AudioURL:       "", // Not needed for description generation
		Description:    description,
	}

	enhancedScript := factGen.GenerateFactScript(bird, latitude, longitude)
	fmt.Printf("[ENHANCED_FACTS] Generated script (%s generator): %d characters\n", factGen.GetGeneratorType(), len(enhancedScript))
	fmt.Printf("[ENHANCED_FACTS] Track 4 text: %s\n", enhancedScript)

	// If the script is too short or empty, fall back to basic
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

	// Generate TTS audio
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     enhancedScript,
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":         0.40,
			"similarity_boost":  0.90,
			"use_speaker_boost": true,
			"speed":             1.0,
			"style":             0,
		},
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

	// Store the description text for the next track (outro)
	cm.lastDescriptionText = enhancedScript

	return audioData, nil
}

// generateEnhancedBirdDescription creates location-aware audio narration using V4 fact generator
// DEPRECATED: Use generateEnhancedBirdDescriptionModular instead
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
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":         0.40,
			"similarity_boost":  0.90,
			"use_speaker_boost": true,
			"speed":             1.0,
			"style":             0,
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
		ambiencePath := filepath.Join("assets/sound_effects", selectedAmbience.Path)
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
