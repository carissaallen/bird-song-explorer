package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
)

// GenerateBirdAnnouncement generates TTS audio for bird announcement
func (am *AudioManager) GenerateBirdAnnouncement(birdName string, voiceID string) ([]byte, error) {
	announcementText := fmt.Sprintf("Today's bird! %s!", birdName)
	
	// Generate TTS audio using ElevenLabs
	return am.generateElevenLabsAudio(announcementText, voiceID)
}

// GenerateLocationAwareDescription generates location-aware bird description audio
func (am *AudioManager) GenerateLocationAwareDescription(bird *models.Bird, voiceID string, latitude, longitude float64) ([]byte, error) {
	// Generate fact script based on location
	var factScript string
	
	// Check which generator to use
	generatorType := os.Getenv("BIRD_FACT_GENERATOR")
	if generatorType == "" {
		generatorType = "enhanced" // Default to enhanced for streaming
	}
	
	// Get eBird API key
	ebirdAPIKey := os.Getenv("EBIRD_API_KEY")
	
	log.Printf("[AUDIO_STREAMING] Using %s facts generator for %s at (%f, %f)", 
		generatorType, bird.CommonName, latitude, longitude)
	
	if generatorType == "enhanced" {
		// Use enhanced generator
		generator := NewEnhancedFactGenerator(ebirdAPIKey)
		factScript = generator.GenerateFactScript(bird, latitude, longitude)
	} else {
		// Use basic generator
		generator := NewImprovedFactGeneratorV4(ebirdAPIKey)
		factScript = generator.GenerateExplorersGuideScriptWithLocation(bird, latitude, longitude)
	}
	
	log.Printf("[AUDIO_STREAMING] Generated fact script: %d characters", len(factScript))
	
	// Generate TTS audio
	return am.generateElevenLabsAudio(factScript, voiceID)
}

// GenerateOutro generates outro audio
func (am *AudioManager) GenerateOutro(birdName string, voiceID string) ([]byte, error) {
	// Check if we should use pre-recorded outros
	useStaticOutros := os.Getenv("USE_STATIC_OUTROS")
	
	var outroText string
	if useStaticOutros == "true" {
		// For streaming, we can't use pre-recorded files directly
		// Generate a simple TTS outro instead
		outroText = "See you tomorrow, explorers! Keep watching and listening for birds!"
	} else {
		// Generate dynamic outro
		outroText = fmt.Sprintf("That was the amazing %s! See you tomorrow for another bird adventure, explorers!", birdName)
	}
	
	// Generate TTS audio
	return am.generateElevenLabsAudio(outroText, voiceID)
}

// generateElevenLabsAudio generates audio using ElevenLabs API
func (am *AudioManager) generateElevenLabsAudio(text string, voiceID string) ([]byte, error) {
	if am.elevenLabsKey == "" {
		return nil, fmt.Errorf("no ElevenLabs API key configured")
	}
	
	// Generate speech using ElevenLabs
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	
	requestBody := map[string]interface{}{
		"text":     text,
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
	req.Header.Set("xi-api-key", am.elevenLabsKey)
	
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
	
	return audioData, nil
}