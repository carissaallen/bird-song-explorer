package services

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
)

// StaticOutroManager uses pre-recorded outro files instead of TTS
type StaticOutroManager struct {
	outroDir     string
	useStatic    bool
	voiceManager *config.VoiceManager
}

// NewStaticOutroManager creates a manager for pre-recorded outros
func NewStaticOutroManager() *StaticOutroManager {
	return &StaticOutroManager{
		outroDir:     "assets/final_outros",
		useStatic:    true, // Can be toggled via env var
		voiceManager: config.NewVoiceManager(),
	}
}

// GetOutroURL returns a URL to a pre-recorded outro file
func (som *StaticOutroManager) GetOutroURL(voiceName string, dayOfWeek time.Weekday, baseURL string) (string, error) {
	outroType := som.getOutroType(dayOfWeek)

	// Find available outros of this type for this voice
	pattern := filepath.Join(som.outroDir, fmt.Sprintf("outro_%s_*_%s.mp3", outroType, voiceName))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("no outros found for %s/%s", outroType, voiceName)
	}

	// Select one deterministically based on date
	// This ensures the same outro is used throughout the day
	now := time.Now()
	daySeed := now.Year()*10000 + int(now.Month())*100 + now.Day()
	outroIndex := daySeed % len(matches)
	selectedFile := matches[outroIndex]

	// Return URL to the outro file
	filename := filepath.Base(selectedFile)
	return fmt.Sprintf("%s/audio/outros/%s", baseURL, filename), nil
}

// GetOutroWithBirdSongURL creates an outro mixed with bird song
// The bird song is mixed in at runtime since it varies per bird
func (som *StaticOutroManager) GetOutroWithBirdSongURL(
	voiceName string,
	dayOfWeek time.Weekday,
	birdSongData []byte,
	baseURL string,
) (string, error) {
	// Get the pre-recorded outro
	outroURL, err := som.GetOutroURL(voiceName, dayOfWeek, baseURL)
	if err != nil {
		return "", err
	}

	// In production, this would:
	// 1. Download the outro file
	// 2. Mix it with the bird song using AudioMixer
	// 3. Cache the result
	// 4. Return the URL to the mixed version

	// For now, return the outro URL (mixing would happen here)
	return outroURL, nil
}

// getOutroType determines which type of outro to use based on the day
func (som *StaticOutroManager) getOutroType(dayOfWeek time.Weekday) string {
	switch dayOfWeek {
	case time.Monday, time.Friday:
		return "joke"
	case time.Tuesday, time.Thursday:
		return "teaser"
	case time.Wednesday:
		return "wisdom"
	case time.Saturday:
		return "challenge"
	case time.Sunday:
		return "funfact"
	default:
		return "teaser"
	}
}

// CountAvailableOutros returns how many outros are available
func (som *StaticOutroManager) CountAvailableOutros() map[string]int {
	counts := make(map[string]int)
	types := []string{"joke", "wisdom", "teaser", "challenge", "funfact"}
	// Get all configured voices
	voices := []string{}
	for _, voice := range som.voiceManager.GetAvailableVoices() {
		voices = append(voices, voice.Name)
	}

	for _, voice := range voices {
		for _, outroType := range types {
			pattern := filepath.Join(som.outroDir, fmt.Sprintf("outro_%s_*_%s.mp3", outroType, voice))
			matches, _ := filepath.Glob(pattern)
			key := fmt.Sprintf("%s_%s", voice, outroType)
			counts[key] = len(matches)
		}
	}

	return counts
}

// MixOutroWithBirdSong mixes a pre-recorded outro with bird song
func (som *StaticOutroManager) MixOutroWithBirdSong(outroData []byte, birdSongData []byte) ([]byte, error) {
	// Use the existing AudioMixer to combine outro speech with bird song
	mixer := NewAudioMixer()

	// Mix at lower volume since it's background for the outro
	// Bird song plays softly under the outro speech
	return mixer.MixOutroWithNatureSounds(outroData, birdSongData)
}
