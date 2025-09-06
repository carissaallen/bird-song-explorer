package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
)

type NarrationManager struct {
	elevenLabsKey string
	selectedVoice VoiceConfig
	introTexts    []string
	voiceManager  *config.VoiceManager
}

type VoiceConfig struct {
	Name    string
	VoiceID string
}

func NewNarrationManager(elevenLabsKey string) *NarrationManager {
	return &NarrationManager{
		elevenLabsKey: elevenLabsKey,
		voiceManager:  config.NewVoiceManager(),
		introTexts: []string{
			"Welcome, nature detectives! Time to discover an amazing bird from your neighborhood.",
			"Hello, bird explorers! Today's special bird is waiting to sing for you.",
			"Ready for an adventure? Let's meet today's featured bird from your area!",
			"Welcome back, little listeners! A wonderful bird is calling just for you.",
			"Hello, young scientists! Let's explore the amazing birds living near you.",
			"Calling all bird lovers! Your daily bird discovery awaits.",
			"Time for today's bird adventure! Listen closely to nature's music.",
			"Welcome to your daily bird journey! Let's discover who's singing today.",
		},
	}
}

// SelectDailyVoice returns the daily voice using voice rotation
func (nm *NarrationManager) SelectDailyVoice() VoiceConfig {
	voice := nm.voiceManager.GetDailyVoice()
	nm.selectedVoice = VoiceConfig{
		Name:    voice.Name,
		VoiceID: voice.ID,
	}

	return nm.selectedVoice
}

// GetRandomVoice returns the daily voice (no longer random, for consistency)
func (nm *NarrationManager) GetRandomVoice() VoiceConfig {
	voice := nm.voiceManager.GetDailyVoice()
	nm.selectedVoice = VoiceConfig{
		Name:    voice.Name,
		VoiceID: voice.ID,
	}
	return nm.selectedVoice
}

// GenerateIntro creates intro audio with the selected voice
func (nm *NarrationManager) GenerateIntro() ([]byte, error) {
	// Ensure voice is selected
	if nm.selectedVoice.VoiceID == "" {
		nm.SelectDailyVoice()
	}

	// Select random intro text
	rand.Seed(time.Now().UnixNano())
	introText := nm.introTexts[rand.Intn(len(nm.introTexts))]

	return nm.generateAudio(introText, nm.selectedVoice)
}

// GenerateBirdIntro creates a bird-specific intro with the same voice
func (nm *NarrationManager) GenerateBirdIntro(birdName string) ([]byte, error) {
	// Ensure voice is selected
	if nm.selectedVoice.VoiceID == "" {
		nm.SelectDailyVoice()
	}

	introTemplates := []string{
		"Today's featured friend is the %s! Let's hear their beautiful song.",
		"Listen closely! The amazing %s has something special to share with you.",
		"Get ready to meet the wonderful %s from your neighborhood!",
		"Your bird discovery today is the %s! What an incredible creature!",
	}

	rand.Seed(time.Now().UnixNano())
	template := introTemplates[rand.Intn(len(introTemplates))]
	text := fmt.Sprintf(template, birdName)

	return nm.generateAudio(text, nm.selectedVoice)
}

// GenerateFact creates fact narration with the same voice
func (nm *NarrationManager) GenerateFact(fact string) ([]byte, error) {
	// Ensure voice is selected
	if nm.selectedVoice.VoiceID == "" {
		nm.SelectDailyVoice()
	}

	return nm.generateAudio(fact, nm.selectedVoice)
}

// GenerateAllFactsForBird generates all fact narrations with consistent voice
func (nm *NarrationManager) GenerateAllFactsForBird(facts []string) ([][]byte, error) {
	// Ensure voice is selected
	if nm.selectedVoice.VoiceID == "" {
		nm.SelectDailyVoice()
	}

	audioFiles := make([][]byte, 0, len(facts))

	for _, fact := range facts {
		audio, err := nm.generateAudio(fact, nm.selectedVoice)
		if err != nil {
			return nil, fmt.Errorf("failed to generate fact audio: %w", err)
		}
		audioFiles = append(audioFiles, audio)
	}

	return audioFiles, nil
}

// generateAudio creates audio using ElevenLabs with specified voice
func (nm *NarrationManager) generateAudio(text string, voice VoiceConfig) ([]byte, error) {
	if nm.elevenLabsKey == "" {
		return nil, fmt.Errorf("ElevenLabs API key not configured")
	}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voice.VoiceID)

	payload := map[string]interface{}{
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

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("xi-api-key", nm.elevenLabsKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ElevenLabs API error: %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetSelectedVoiceName returns the name of the currently selected voice
func (nm *NarrationManager) GetSelectedVoiceName() string {
	if nm.selectedVoice.Name == "" {
		nm.SelectDailyVoice()
	}
	return nm.selectedVoice.Name
}
