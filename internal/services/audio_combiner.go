package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/callen/bird-song-explorer/internal/config"
)

// AudioCombiner handles combining intro audio with bird name announcement
type AudioCombiner struct {
	elevenLabsKey string
	cacheDir      string
	httpClient    *http.Client
}

// NewAudioCombiner creates a new audio combiner
func NewAudioCombiner(elevenLabsKey string) *AudioCombiner {
	cacheDir := "./audio_cache/combined"
	os.MkdirAll(cacheDir, 0755)

	return &AudioCombiner{
		elevenLabsKey: elevenLabsKey,
		cacheDir:      cacheDir,
		httpClient:    &http.Client{},
	}
}

// GetIntroWithBirdName returns a URL to an intro audio file that includes the bird name
func (ac *AudioCombiner) GetIntroWithBirdName(introURL string, birdName string, voiceID string) (string, error) {
	// Check if we have a cached version
	cacheKey := ac.getCacheKey(introURL, birdName)
	cachedPath := filepath.Join(ac.cacheDir, cacheKey+".mp3")

	if _, err := os.Stat(cachedPath); err == nil {
		// Return the cached file URL
		return "/audio/cache/combined/" + cacheKey + ".mp3", nil
	}

	// Generate the bird name announcement
	announcementAudio, err := ac.generateBirdAnnouncement(birdName, voiceID)
	if err != nil {
		return "", fmt.Errorf("failed to generate bird announcement: %w", err)
	}

	// Download the intro audio
	introAudio, err := ac.downloadAudio(introURL)
	if err != nil {
		return "", fmt.Errorf("failed to download intro: %w", err)
	}

	// Combine the audio files (simple concatenation for now)
	// In production, you might want to use ffmpeg for proper audio processing
	combined := append(introAudio, announcementAudio...)

	// Save to cache
	if err := os.WriteFile(cachedPath, combined, 0644); err != nil {
		return "", fmt.Errorf("failed to save combined audio: %w", err)
	}

	return "/audio/cache/combined/" + cacheKey + ".mp3", nil
}

// generateBirdAnnouncement creates TTS audio saying "Today's bird is the [bird name]!"
func (ac *AudioCombiner) generateBirdAnnouncement(birdName string, voiceID string) ([]byte, error) {
	// Use provided voice ID or get daily voice
	if voiceID == "" {
		voiceManager := config.NewVoiceManager()
		dailyVoice := voiceManager.GetDailyVoice()
		voiceID = dailyVoice.ID
	}

	text := fmt.Sprintf("Today's bird is the %s!", birdName)

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.5,
			"speed":            0.95, // Slightly slower speed for kids (95% of normal)
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
	req.Header.Set("xi-api-key", ac.elevenLabsKey)

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ElevenLabs API error: %d - %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// downloadAudio downloads audio from a URL
func (ac *AudioCombiner) downloadAudio(audioURL string) ([]byte, error) {
	resp, err := ac.httpClient.Get(audioURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download audio: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// getCacheKey generates a cache key for the combined audio
func (ac *AudioCombiner) getCacheKey(introURL string, birdName string) string {
	// Extract intro filename
	parts := strings.Split(introURL, "/")
	introFile := parts[len(parts)-1]
	introName := strings.TrimSuffix(introFile, filepath.Ext(introFile))

	// Create a filesystem-safe bird name
	safeBirdName := strings.ReplaceAll(birdName, " ", "_")
	safeBirdName = strings.ReplaceAll(safeBirdName, "'", "")
	safeBirdName = strings.ReplaceAll(safeBirdName, ",", "")

	return fmt.Sprintf("%s_%s", introName, safeBirdName)
}
