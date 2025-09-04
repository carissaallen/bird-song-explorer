package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// EnhancedBirdAnnouncement handles mixing bird announcements with continuing ambience
type EnhancedBirdAnnouncement struct {
	enhancedMixer *EnhancedIntroMixer
	cacheDir      string
}

// NewEnhancedBirdAnnouncement creates a new enhanced bird announcement service
func NewEnhancedBirdAnnouncement(elevenLabsKey string) *EnhancedBirdAnnouncement {
	return &EnhancedBirdAnnouncement{
		enhancedMixer: NewEnhancedIntroMixer(elevenLabsKey),
		cacheDir:      "./audio_cache/enhanced_announcements",
	}
}

// GenerateAnnouncementWithAmbience creates Track 2 with continued ambience from Track 1
func (eba *EnhancedBirdAnnouncement) GenerateAnnouncementWithAmbience(
	birdName string,
	voiceID string,
	selectedAmbience string,
) ([]byte, error) {
	// Ensure cache directory exists
	os.MkdirAll(eba.cacheDir, 0755)

	// Generate TTS announcement without pauses for better pacing
	announcementText := fmt.Sprintf("Today's bird is the %s! Listen carefully to its unique song.", birdName)

	// Use narration manager for TTS
	narrationManager := NewNarrationManager(os.Getenv("ELEVENLABS_API_KEY"))
	narrationManager.selectedVoice = VoiceConfig{VoiceID: voiceID}
	ttsData, err := narrationManager.generateAudio(announcementText, narrationManager.selectedVoice)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TTS announcement: %w", err)
	}

	// If no ambience selected, return just the TTS
	if selectedAmbience == "" {
		return ttsData, nil
	}

	// Get the ambience file path based on selection
	var ambiencePath string
	switch selectedAmbience {
	case "forest":
		ambiencePath = "sound_effects/ambience/forest-ambience.mp3"
	case "jungle":
		ambiencePath = "sound_effects/ambience/jungle_sounds.mp3"
	case "morning":
		ambiencePath = "sound_effects/ambience/morning-birdsong.mp3"
	default:
		// Unknown ambience, return just TTS
		return ttsData, nil
	}

	// Check if ambience file exists
	if _, err := os.Stat(ambiencePath); err != nil {
		// File not found, return just TTS
		return ttsData, nil
	}

	// Mix the announcement with continuing ambience
	mixedAudio, err := eba.mixAnnouncementWithAmbience(ttsData, ambiencePath)
	if err != nil {
		// If mixing fails, return just the TTS
		return ttsData, nil
	}

	return mixedAudio, nil
}

// mixAnnouncementWithAmbience combines the announcement with background ambience
func (eba *EnhancedBirdAnnouncement) mixAnnouncementWithAmbience(ttsData []byte, ambiencePath string) ([]byte, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH")
	}

	// Create temp files
	tempDir := os.TempDir()
	ttsFile := filepath.Join(tempDir, fmt.Sprintf("announcement_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("announcement_mixed_%d.mp3", time.Now().Unix()))

	// Write TTS data to temp file
	if err := os.WriteFile(ttsFile, ttsData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write TTS file: %w", err)
	}
	defer os.Remove(ttsFile)
	defer os.Remove(outputFile)

	// Get TTS duration
	ttsDuration := eba.getAudioDuration(ttsFile)
	if ttsDuration <= 0 {
		ttsDuration = 3.0 // Default fallback
	}

	// Calculate timings
	// Ambience continues from Track 1 at low volume
	// Let the voice complete fully before any fade out
	fadeInDuration := 0.1                    // Very quick fade in to prevent cutoff
	fadeOutStart := ttsDuration + 0.5        // Start fading AFTER voice completes
	fadeOutDuration := 1.0                   // Fade duration for ambience
	totalDuration := ttsDuration + 2.0       // Add 2 seconds after voice for fade

	// Build ffmpeg command for mixing with volume normalization
	cmd := exec.Command("ffmpeg",
		"-i", ambiencePath, // Input: continuing ambience
		"-i", ttsFile, // Input: TTS announcement
		"-filter_complex",
		fmt.Sprintf(
			// Ambience: start at 15%% volume (matching intro/outro), fade out gradually
			"[0:a]afade=t=in:st=0:d=%.1f,volume=0.15[ambience_low];"+
				"[ambience_low]afade=t=out:st=%.1f:d=%.1f[ambience_fade];"+
				// Boost TTS volume to match dynamic track levels with 0.1s fade-in to prevent cutoff
				"[1:a]afade=t=in:st=0:d=0.1,volume=2.2[voice_boosted];"+
				// Mix boosted voice with fading ambience
				"[voice_boosted][ambience_fade]amix=inputs=2:duration=first:dropout_transition=0.5[mixed];"+
				// Final fade out after voice completes - no loudnorm to preserve dynamics
				"[mixed]afade=t=out:st=%.1f:d=%.1f[out]",
			fadeInDuration,  // Quick fade in
			fadeOutStart,    // When to start fade out (after voice)
			fadeOutDuration, // Fade out duration
			fadeOutStart,    // Final fade start (same as ambience fade)
			fadeOutDuration, // Final fade duration
		),
		"-map", "[out]",
		"-t", fmt.Sprintf("%.2f", totalDuration),
		"-c:a", "libmp3lame",
		"-b:a", "192k",
		"-ar", "44100",
		"-y",
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg mixing failed: %v\nStderr: %s", err, stderr.String())
	}

	// Read the mixed audio
	mixedData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read mixed audio: %w", err)
	}

	fmt.Printf("[ENHANCED_ANNOUNCEMENT] Successfully created announcement with %s ambience and volume normalization (size: %d bytes)\n",
		filepath.Base(ambiencePath), len(mixedData))
	return mixedData, nil
}

// GenerateAnnouncementFromAudioData creates Track 2 using raw ambience audio data
func (eba *EnhancedBirdAnnouncement) GenerateAnnouncementFromAudioData(
	birdName string,
	voiceID string,
	ambienceData []byte,
) ([]byte, error) {
	// Ensure cache directory exists
	os.MkdirAll(eba.cacheDir, 0755)

	// Generate TTS announcement without pauses for better pacing
	announcementText := fmt.Sprintf("Today's bird is the %s! Listen carefully to its unique song.", birdName)

	// Use narration manager for TTS
	narrationManager := NewNarrationManager(os.Getenv("ELEVENLABS_API_KEY"))
	narrationManager.selectedVoice = VoiceConfig{VoiceID: voiceID}
	ttsData, err := narrationManager.generateAudio(announcementText, narrationManager.selectedVoice)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TTS announcement: %w", err)
	}

	// If no ambience data, return just the TTS
	if len(ambienceData) == 0 {
		return ttsData, nil
	}

	// Create temp file for ambience
	tempDir := os.TempDir()
	ambienceFile := filepath.Join(tempDir, fmt.Sprintf("ambience_%d.mp3", time.Now().Unix()))
	if err := os.WriteFile(ambienceFile, ambienceData, 0644); err != nil {
		return ttsData, nil // Return TTS only if we can't write ambience
	}
	defer os.Remove(ambienceFile)

	// Mix the announcement with continuing ambience
	mixedAudio, err := eba.mixAnnouncementWithAmbience(ttsData, ambienceFile)
	if err != nil {
		// If mixing fails, return just the TTS
		return ttsData, nil
	}

	return mixedAudio, nil
}

// getAudioDuration gets the duration of an audio file using ffprobe
func (eba *EnhancedBirdAnnouncement) getAudioDuration(audioFile string) float64 {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioFile,
	)

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("[ENHANCED_ANNOUNCEMENT] Failed to get audio duration: %v\n", err)
		return 0
	}

	duration := 0.0
	fmt.Sscanf(string(output), "%f", &duration)
	return duration
}
