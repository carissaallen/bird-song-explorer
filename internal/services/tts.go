package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/callen/bird-song-explorer/internal/config"
)

type TTSService struct {
	elevenLabsKey string
	openAIKey     string
}

func NewTTSService(elevenLabsKey, openAIKey string) *TTSService {
	return &TTSService{
		elevenLabsKey: elevenLabsKey,
		openAIKey:     openAIKey,
	}
}

const introScript = `Good morning, little explorers! Welcome to Bird Song Explorer. 
Today's special bird is calling just for you. Listen carefully!`

func (s *TTSService) GenerateIntroAudio() ([]byte, error) {
	if s.elevenLabsKey != "" {
		return s.generateElevenLabsAudio(introScript)
	}
	if s.openAIKey != "" {
		return s.generateOpenAIAudio(introScript)
	}
	return nil, fmt.Errorf("no TTS API key configured")
}

func (s *TTSService) generateElevenLabsAudio(text string) ([]byte, error) {
	// Use daily voice for consistency
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()
	voiceID := dailyVoice.ID
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	payload := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.3,  // Lower for more emotional range
			"similarity_boost": 0.85, // High similarity to original voice
			"speed":            0.92, // Good pace for kids
			"use_speaker_boost": true, // Enhance voice clarity
		},
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("xi-api-key", s.elevenLabsKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ElevenLabs API error: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (s *TTSService) generateOpenAIAudio(text string) ([]byte, error) {
	url := "https://api.openai.com/v1/audio/speech"

	payload := map[string]string{
		"model": "tts-1",
		"input": text,
		"voice": "nova", // or "alloy" for younger sound
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.openAIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (s *TTSService) GenerateBirdFactAudio(fact string) ([]byte, error) {
	if s.elevenLabsKey != "" {
		return s.generateElevenLabsAudio(fact)
	}
	if s.openAIKey != "" {
		return s.generateOpenAIAudio(fact)
	}
	return nil, fmt.Errorf("no TTS API key configured")
}
