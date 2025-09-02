package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
)

type AudioManager struct {
	introDir      string
	cacheDir      string
	introURLs     []string // URLs where intros are hosted
	enhancedMixer *EnhancedIntroMixer
	elevenLabsKey string
}

func NewAudioManager() *AudioManager {
	// Get ElevenLabs API key from environment
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	return &AudioManager{
		introDir:      "final_intros",
		cacheDir:      "audio_cache",
		introURLs:     []string{}, // Will be populated with hosted URLs
		enhancedMixer: NewEnhancedIntroMixer(elevenLabsKey),
		elevenLabsKey: elevenLabsKey,
	}
}

// GetRandomIntroURL returns a URL to an intro that matches the daily voice
// The intro is selected deterministically based on the day to ensure consistency
// Also returns the voice ID to ensure other tracks use the same voice
func (am *AudioManager) GetRandomIntroURL(baseURL string) (string, string) {
	// Get the daily voice
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()

	// Check if we should use enhanced intro (with local sound effects)
	if am.shouldUseEnhancedIntro() {
		return am.getEnhancedIntroURL(baseURL, dailyVoice.ID)
	}

	// All available intro files for all voices
	// These should match the voices defined in config/voices.go
	allIntros := []string{
		// Amelia (British)
		"intro_00_Amelia.mp3", "intro_01_Amelia.mp3", "intro_02_Amelia.mp3", "intro_03_Amelia.mp3",
		"intro_04_Amelia.mp3", "intro_05_Amelia.mp3", "intro_06_Amelia.mp3", "intro_07_Amelia.mp3",
		// Antoni (American)
		"intro_00_Antoni.mp3", "intro_01_Antoni.mp3", "intro_02_Antoni.mp3", "intro_03_Antoni.mp3",
		"intro_04_Antoni.mp3", "intro_05_Antoni.mp3", "intro_06_Antoni.mp3", "intro_07_Antoni.mp3",
		// Charlotte (Australian)
		"intro_00_Charlotte.mp3", "intro_01_Charlotte.mp3", "intro_02_Charlotte.mp3", "intro_03_Charlotte.mp3",
		"intro_04_Charlotte.mp3", "intro_05_Charlotte.mp3", "intro_06_Charlotte.mp3", "intro_07_Charlotte.mp3",
		// Danielle
		"intro_00_Danielle.mp3", "intro_01_Danielle.mp3", "intro_02_Danielle.mp3", "intro_03_Danielle.mp3",
		"intro_04_Danielle.mp3", "intro_05_Danielle.mp3", "intro_06_Danielle.mp3", "intro_07_Danielle.mp3",
		// Drake (Canadian)
		"intro_00_Drake.mp3", "intro_01_Drake.mp3", "intro_02_Drake.mp3", "intro_03_Drake.mp3",
		"intro_04_Drake.mp3", "intro_05_Drake.mp3", "intro_06_Drake.mp3", "intro_07_Drake.mp3",
		// Hope
		"intro_00_Hope.mp3", "intro_01_Hope.mp3", "intro_02_Hope.mp3", "intro_03_Hope.mp3",
		"intro_04_Hope.mp3", "intro_05_Hope.mp3", "intro_06_Hope.mp3", "intro_07_Hope.mp3",
		// Peter (Irish)
		"intro_00_Peter.mp3", "intro_01_Peter.mp3", "intro_02_Peter.mp3", "intro_03_Peter.mp3",
		"intro_04_Peter.mp3", "intro_05_Peter.mp3", "intro_06_Peter.mp3", "intro_07_Peter.mp3",
		// Rory
		"intro_00_Rory.mp3", "intro_01_Rory.mp3", "intro_02_Rory.mp3", "intro_03_Rory.mp3",
		"intro_04_Rory.mp3", "intro_05_Rory.mp3", "intro_06_Rory.mp3", "intro_07_Rory.mp3",
		// Sally (Southern U.S.)
		"intro_00_Sally.mp3", "intro_01_Sally.mp3", "intro_02_Sally.mp3", "intro_03_Sally.mp3",
		"intro_04_Sally.mp3", "intro_05_Sally.mp3", "intro_06_Sally.mp3", "intro_07_Sally.mp3",
		// Stuart
		"intro_00_Stuart.mp3", "intro_01_Stuart.mp3", "intro_02_Stuart.mp3", "intro_03_Stuart.mp3",
		"intro_04_Stuart.mp3", "intro_05_Stuart.mp3", "intro_06_Stuart.mp3", "intro_07_Stuart.mp3",
	}

	// Filter intros by voice name
	var voiceIntros []string
	for _, intro := range allIntros {
		// Check if the intro filename contains the voice name
		if strings.Contains(intro, dailyVoice.Name) {
			voiceIntros = append(voiceIntros, intro)
		}
	}

	// If no intros found for this voice (e.g., Sarah), fall back to Antoni
	if len(voiceIntros) == 0 {
		fmt.Printf("Warning: No pre-recorded intros for voice %s, using Antoni\n", dailyVoice.Name)
		// Default to Antoni intros if voice has no pre-recorded intros
		for _, intro := range allIntros {
			if strings.Contains(intro, "Antoni") {
				voiceIntros = append(voiceIntros, intro)
			}
		}
	}

	// Select an intro deterministically based on the day
	// This ensures the same intro is used throughout the day for consistency
	now := time.Now()
	daySeed := now.Year()*10000 + int(now.Month())*100 + now.Day()
	// Use a different seed component to avoid always picking the same intro number as voice
	introIndex := (daySeed * 7) % len(voiceIntros)
	selected := voiceIntros[introIndex]

	// Return intro URL and voice ID to ensure consistency across all tracks
	return fmt.Sprintf("%s/audio/intros/%s", baseURL, selected), dailyVoice.ID
}

// shouldUseEnhancedIntro checks if we should use the enhanced intro with local sound effects
func (am *AudioManager) shouldUseEnhancedIntro() bool {
	// Check if sound_effects directory exists
	if _, err := os.Stat("sound_effects"); err != nil {
		return false
	}
	// Check if we have the required sound files
	requiredFiles := []string{
		"sound_effects/ambience/forest-ambience.mp3",
		"sound_effects/ambience/jungle_sounds.mp3",
		"sound_effects/ambience/morning-birdsong.mp3",
		"sound_effects/chimes/sparkle_chime.mp3",
	}
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); err != nil {
			return false
		}
	}
	return true
}

// getEnhancedIntroURL generates and caches an enhanced intro, returning its URL
func (am *AudioManager) getEnhancedIntroURL(baseURL string, voiceID string) (string, string) {
	// Generate cache key based on date and voice
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	cacheKey := fmt.Sprintf("enhanced_intro_%s_%s", dateStr, voiceID)
	cachePath := filepath.Join(am.cacheDir, "enhanced_intros", cacheKey+".mp3")

	// Check if we have a cached version
	if _, err := os.Stat(cachePath); err == nil {
		// Return cached URL
		return fmt.Sprintf("%s/audio/cache/enhanced_intros/%s.mp3", baseURL, cacheKey), voiceID
	}

	// Generate enhanced intro using pre-recorded files
	introData, err := am.enhancedMixer.GenerateEnhancedIntroWithPreRecorded(voiceID)
	if err != nil {
		fmt.Printf("Failed to generate enhanced intro: %v, falling back to standard\n", err)
		// Fall back to standard intro
		return am.getStandardIntroURL(baseURL, voiceID)
	}

	// Ensure cache directory exists
	os.MkdirAll(filepath.Dir(cachePath), 0755)

	// Save to cache
	if err := os.WriteFile(cachePath, introData, 0644); err != nil {
		fmt.Printf("Failed to cache enhanced intro: %v\n", err)
	}

	return fmt.Sprintf("%s/audio/cache/enhanced_intros/%s.mp3", baseURL, cacheKey), voiceID
}

// getStandardIntroURL returns a standard pre-recorded intro URL (fallback)
func (am *AudioManager) getStandardIntroURL(baseURL string, voiceID string) (string, string) {
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()

	// All available intro files for all voices (8 intros per voice)
	allIntros := []string{
		// Amelia (British)
		"intro_00_Amelia.mp3", "intro_01_Amelia.mp3", "intro_02_Amelia.mp3", "intro_03_Amelia.mp3",
		"intro_04_Amelia.mp3", "intro_05_Amelia.mp3", "intro_06_Amelia.mp3", "intro_07_Amelia.mp3",
		// Antoni (American)
		"intro_00_Antoni.mp3", "intro_01_Antoni.mp3", "intro_02_Antoni.mp3", "intro_03_Antoni.mp3",
		"intro_04_Antoni.mp3", "intro_05_Antoni.mp3", "intro_06_Antoni.mp3", "intro_07_Antoni.mp3",
		// Charlotte (Australian)
		"intro_00_Charlotte.mp3", "intro_01_Charlotte.mp3", "intro_02_Charlotte.mp3", "intro_03_Charlotte.mp3",
		"intro_04_Charlotte.mp3", "intro_05_Charlotte.mp3", "intro_06_Charlotte.mp3", "intro_07_Charlotte.mp3",
		// Danielle
		"intro_00_Danielle.mp3", "intro_01_Danielle.mp3", "intro_02_Danielle.mp3", "intro_03_Danielle.mp3",
		"intro_04_Danielle.mp3", "intro_05_Danielle.mp3", "intro_06_Danielle.mp3", "intro_07_Danielle.mp3",
		// Drake (Canadian)
		"intro_00_Drake.mp3", "intro_01_Drake.mp3", "intro_02_Drake.mp3", "intro_03_Drake.mp3",
		"intro_04_Drake.mp3", "intro_05_Drake.mp3", "intro_06_Drake.mp3", "intro_07_Drake.mp3",
		// Hope
		"intro_00_Hope.mp3", "intro_01_Hope.mp3", "intro_02_Hope.mp3", "intro_03_Hope.mp3",
		"intro_04_Hope.mp3", "intro_05_Hope.mp3", "intro_06_Hope.mp3", "intro_07_Hope.mp3",
		// Peter (Irish)
		"intro_00_Peter.mp3", "intro_01_Peter.mp3", "intro_02_Peter.mp3", "intro_03_Peter.mp3",
		"intro_04_Peter.mp3", "intro_05_Peter.mp3", "intro_06_Peter.mp3", "intro_07_Peter.mp3",
		// Rory
		"intro_00_Rory.mp3", "intro_01_Rory.mp3", "intro_02_Rory.mp3", "intro_03_Rory.mp3",
		"intro_04_Rory.mp3", "intro_05_Rory.mp3", "intro_06_Rory.mp3", "intro_07_Rory.mp3",
		// Sally (Southern U.S.)
		"intro_00_Sally.mp3", "intro_01_Sally.mp3", "intro_02_Sally.mp3", "intro_03_Sally.mp3",
		"intro_04_Sally.mp3", "intro_05_Sally.mp3", "intro_06_Sally.mp3", "intro_07_Sally.mp3",
		// Stuart
		"intro_00_Stuart.mp3", "intro_01_Stuart.mp3", "intro_02_Stuart.mp3", "intro_03_Stuart.mp3",
		"intro_04_Stuart.mp3", "intro_05_Stuart.mp3", "intro_06_Stuart.mp3", "intro_07_Stuart.mp3",
	}

	// Filter intros by voice name
	var voiceIntros []string
	for _, intro := range allIntros {
		if strings.Contains(intro, dailyVoice.Name) {
			voiceIntros = append(voiceIntros, intro)
		}
	}

	// If no intros found for this voice, fall back to Antoni
	if len(voiceIntros) == 0 {
		for _, intro := range allIntros {
			if strings.Contains(intro, "Antoni") {
				voiceIntros = append(voiceIntros, intro)
			}
		}
	}

	// Select an intro deterministically based on the day
	now := time.Now()
	daySeed := now.Year()*10000 + int(now.Month())*100 + now.Day()
	introIndex := (daySeed * 7) % len(voiceIntros)
	selected := voiceIntros[introIndex]

	return fmt.Sprintf("%s/audio/intros/%s", baseURL, selected), voiceID
}

// GetEnhancedMixer returns the enhanced intro mixer for use by other services
func (am *AudioManager) GetEnhancedMixer() *EnhancedIntroMixer {
	return am.enhancedMixer
}

// DownloadAndCacheBirdSong downloads bird song from Xeno-canto and caches it
func (am *AudioManager) DownloadAndCacheBirdSong(audioURL string, birdName string) (string, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(am.cacheDir, 0755); err != nil {
		return "", err
	}

	// Generate cache filename
	safeFileName := sanitizeFileName(birdName)
	cacheFile := filepath.Join(am.cacheDir, fmt.Sprintf("%s_%d.mp3", safeFileName, time.Now().Unix()))

	// Check if we already have a recent cache
	existingFiles, _ := filepath.Glob(filepath.Join(am.cacheDir, fmt.Sprintf("%s_*.mp3", safeFileName)))
	for _, file := range existingFiles {
		info, err := os.Stat(file)
		if err == nil && time.Since(info.ModTime()) < 24*time.Hour {
			return file, nil // Use cached file if less than 24 hours old
		}
	}

	// Download the audio file
	resp, err := http.Get(audioURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Create the cache file
	out, err := os.Create(cacheFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Copy the audio data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	// Clean up old cache files for this bird
	for _, file := range existingFiles {
		if file != cacheFile {
			os.Remove(file)
		}
	}

	return cacheFile, nil
}

// TrimAudioToLength trims audio file to specified duration (15-30 seconds)
func (am *AudioManager) TrimAudioToLength(inputFile string, duration int) (string, error) {
	// This would use ffmpeg to trim the audio
	// For now, we'll just return the original file
	// In production, you'd execute: ffmpeg -i input.mp3 -t 30 -acodec copy output.mp3
	return inputFile, nil
}

func sanitizeFileName(name string) string {
	// Replace spaces and special characters
	safe := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			safe += string(r)
		} else if r == ' ' {
			safe += "_"
		}
	}
	return safe
}
