package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
)

// DynamicIntroGenerator generates intro audio with bird name announcements
type DynamicIntroGenerator struct {
	elevenLabsKey string
	cacheDir      string
	httpClient    *http.Client
}

// IntroScript represents different intro variations
var IntroScripts = []string{
	"Hello, nature explorers! Are you ready for today's bird adventure? Today's bird is the %s!",
	"Welcome back to Bird Song Explorer! Let's discover an amazing bird from your neighborhood. Today's bird is the %s!",
	"Good day, young birders! Time to learn about a wonderful feathered friend. Today's bird is the %s!",
	"Hey there, bird detectives! Ready to solve today's mystery? Today's bird is the %s!",
	"Welcome, nature lovers! Let's listen to something special from the bird world. Today's bird is the %s!",
	"Hello, friends! Get your ears ready for an incredible bird song. Today's bird is the %s!",
	"Greetings, explorers! Time for another bird discovery. Today's bird is the %s!",
	"Hi there! Let's go on a sound adventure with our feathered friends. Today's bird is the %s!",
}

// Removed hardcoded voice IDs - now using centralized voice manager

// NewDynamicIntroGenerator creates a new dynamic intro generator
func NewDynamicIntroGenerator(elevenLabsKey string) *DynamicIntroGenerator {
	cacheDir := "./audio_cache/dynamic_intros"
	os.MkdirAll(cacheDir, 0755)

	return &DynamicIntroGenerator{
		elevenLabsKey: elevenLabsKey,
		cacheDir:      cacheDir,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// GenerateIntroWithBirdName creates a complete intro that includes the bird name
func (dig *DynamicIntroGenerator) GenerateIntroWithBirdName(birdName string) (string, error) {
	// Check cache first
	cacheKey := dig.getCacheKey(birdName)
	cachedPath := filepath.Join(dig.cacheDir, cacheKey+".mp3")

	if _, err := os.Stat(cachedPath); err == nil {
		// Return the cached file path
		return cachedPath, nil
	}

	// Select voice based on day (for consistency throughout the day)
	voiceID := dig.selectDailyVoice()

	// Select a random intro script
	script := IntroScripts[rand.Intn(len(IntroScripts))]
	text := fmt.Sprintf(script, birdName)

	// Generate the audio
	audioData, err := dig.generateSpeech(text, voiceID)
	if err != nil {
		return "", fmt.Errorf("failed to generate speech: %w", err)
	}

	// Save to cache
	if err := os.WriteFile(cachedPath, audioData, 0644); err != nil {
		return "", fmt.Errorf("failed to save audio: %w", err)
	}

	return cachedPath, nil
}

// GetDynamicIntroURL generates and returns a URL for the dynamic intro
func (dig *DynamicIntroGenerator) GetDynamicIntroURL(birdName string, baseURL string) (string, error) {
	localPath, err := dig.GenerateIntroWithBirdName(birdName)
	if err != nil {
		return "", err
	}

	// Convert local path to URL path
	filename := filepath.Base(localPath)
	return fmt.Sprintf("%s/audio/cache/dynamic_intros/%s", baseURL, filename), nil
}

// generateSpeech calls ElevenLabs API to generate speech
func (dig *DynamicIntroGenerator) generateSpeech(text string, voiceID string) ([]byte, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.75,
			"similarity_boost": 0.75,
			"style":            0.5,
			"speed":            0.90, // Slower speed for kids (90% of normal)
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
	req.Header.Set("xi-api-key", dig.elevenLabsKey)

	resp, err := dig.httpClient.Do(req)
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

// selectDailyVoice selects a voice based on the current day
func (dig *DynamicIntroGenerator) selectDailyVoice() string {
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()
	return dailyVoice.ID
}

// getCacheKey generates a cache key for the intro
func (dig *DynamicIntroGenerator) getCacheKey(birdName string) string {
	// Include date and voice in cache key for daily variation
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()
	voice := strings.ToLower(dailyVoice.Name)

	// Create a filesystem-safe bird name
	safeBirdName := strings.ReplaceAll(birdName, " ", "_")
	safeBirdName = strings.ReplaceAll(safeBirdName, "'", "")
	safeBirdName = strings.ReplaceAll(safeBirdName, ",", "")
	safeBirdName = strings.ToLower(safeBirdName)

	return fmt.Sprintf("%s_%s_%s", dateStr, voice, safeBirdName)
}
