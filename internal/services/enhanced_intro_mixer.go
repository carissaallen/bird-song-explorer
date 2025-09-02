package services

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
)

// EnhancedIntroMixer handles creating dynamic intros with local sound effects
type EnhancedIntroMixer struct {
	soundEffectsPath string
	cacheDir         string
	selectedAmbience string // Track which ambience was used for passing to next track
	narrationManager *NarrationManager
}

// NewEnhancedIntroMixer creates a new enhanced intro mixer
func NewEnhancedIntroMixer(elevenLabsKey string) *EnhancedIntroMixer {
	return &EnhancedIntroMixer{
		soundEffectsPath: "sound_effects",
		cacheDir:         "./audio_cache/enhanced_intros",
		narrationManager: NewNarrationManager(elevenLabsKey),
	}
}

// AmbienceOption represents an available ambience track
type AmbienceOption struct {
	Name string
	Path string
}

// GetAvailableAmbiences returns the list of available ambience tracks
func (eim *EnhancedIntroMixer) GetAvailableAmbiences() []AmbienceOption {
	return []AmbienceOption{
		{Name: "forest", Path: "ambience/forest-ambience.mp3"},
		{Name: "jungle", Path: "ambience/jungle_sounds.mp3"},
		{Name: "morning", Path: "ambience/morning-birdsong.mp3"},
	}
}

// GenerateEnhancedIntro creates a complete intro with ambience, chime, and TTS
func (eim *EnhancedIntroMixer) GenerateEnhancedIntro(voiceID string) ([]byte, error) {
	// Ensure cache directory exists
	os.MkdirAll(eim.cacheDir, 0755)

	// Select a random ambience track
	ambiences := eim.GetAvailableAmbiences()
	rand.Seed(time.Now().UnixNano())
	selectedAmbience := ambiences[rand.Intn(len(ambiences))]
	eim.selectedAmbience = selectedAmbience.Name // Store for later use

	// Get paths for sound effects
	ambiencePath := filepath.Join(eim.soundEffectsPath, selectedAmbience.Path)
	chimePath := filepath.Join(eim.soundEffectsPath, "chimes/sparkle_chime.mp3")

	// Check if files exist
	if _, err := os.Stat(ambiencePath); err != nil {
		return nil, fmt.Errorf("ambience file not found: %s", ambiencePath)
	}
	if _, err := os.Stat(chimePath); err != nil {
		return nil, fmt.Errorf("chime file not found: %s", chimePath)
	}

	// Generate TTS intro using the narration manager
	eim.narrationManager.selectedVoice = VoiceConfig{VoiceID: voiceID}
	ttsData, err := eim.narrationManager.GenerateIntro()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TTS intro: %w", err)
	}

	// Mix the audio components
	mixedAudio, err := eim.mixAudioComponents(ambiencePath, chimePath, ttsData)
	if err != nil {
		return nil, fmt.Errorf("failed to mix audio components: %w", err)
	}

	return mixedAudio, nil
}

// GenerateEnhancedIntroWithText creates intro with specific text
func (eim *EnhancedIntroMixer) GenerateEnhancedIntroWithText(text string, voiceID string) ([]byte, error) {
	// This method is deprecated - use GenerateEnhancedIntroWithPreRecorded instead
	return eim.GenerateEnhancedIntroWithPreRecorded(voiceID)
}

// GenerateEnhancedIntroWithPreRecorded uses pre-recorded intro files with ambience overlay
func (eim *EnhancedIntroMixer) GenerateEnhancedIntroWithPreRecorded(voiceID string) ([]byte, error) {
	// Ensure cache directory exists
	os.MkdirAll(eim.cacheDir, 0755)

	// Select a random ambience track based on the day (for consistency)
	ambiences := eim.GetAvailableAmbiences()
	now := time.Now()
	daySeed := now.Year()*10000 + int(now.Month())*100 + now.Day()
	selectedAmbience := ambiences[daySeed%len(ambiences)]
	eim.selectedAmbience = selectedAmbience.Name

	// Get paths for sound effects
	ambiencePath := filepath.Join(eim.soundEffectsPath, selectedAmbience.Path)
	chimePath := filepath.Join(eim.soundEffectsPath, "chimes/sparkle_chime.mp3")

	// Check if files exist
	if _, err := os.Stat(ambiencePath); err != nil {
		return nil, fmt.Errorf("ambience file not found: %s", ambiencePath)
	}
	if _, err := os.Stat(chimePath); err != nil {
		return nil, fmt.Errorf("chime file not found: %s", chimePath)
	}

	// Get the voice name for selecting the right pre-recorded intro
	voiceManager := config.NewVoiceManager()
	var voiceName string
	if voiceID == "" {
		dailyVoice := voiceManager.GetDailyVoice()
		voiceName = dailyVoice.Name
		voiceID = dailyVoice.ID
	} else {
		// Look up voice name from ID using GetVoiceByID
		voice := voiceManager.GetVoiceByID(voiceID)
		if voice != nil {
			voiceName = voice.Name
		}
	}

	if voiceName == "" {
		voiceName = "Antoni" // Default fallback
	}

	// Select the pre-recorded intro file based on voice and day
	// Use a different seed component than voice selection for variety
	introIndex := (daySeed * 7) % 8 // 8 intros per voice
	introFileName := fmt.Sprintf("intro_%02d_%s.mp3", introIndex, voiceName)
	introPath := filepath.Join("final_intros", introFileName)

	// Check if the intro file exists
	if _, err := os.Stat(introPath); err != nil {
		// Try Antoni as fallback if voice-specific intro doesn't exist
		introFileName = fmt.Sprintf("intro_%02d_Antoni.mp3", introIndex)
		introPath = filepath.Join("final_intros", introFileName)
		if _, err := os.Stat(introPath); err != nil {
			return nil, fmt.Errorf("pre-recorded intro file not found: %s", introPath)
		}
	}

	// Read the pre-recorded intro file
	introData, err := os.ReadFile(introPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read intro file: %w", err)
	}

	fmt.Printf("[ENHANCED_INTRO] Using pre-recorded intro: %s\n", introFileName)

	// Mix the audio components with volume normalization
	mixedAudio, err := eim.mixAudioComponentsWithNormalization(ambiencePath, chimePath, introData)
	if err != nil {
		return nil, fmt.Errorf("failed to mix audio components: %w", err)
	}

	return mixedAudio, nil
}

// mixAudioComponents combines ambience, chime, and voice into a single track
func (eim *EnhancedIntroMixer) mixAudioComponents(ambiencePath, chimePath string, ttsData []byte) ([]byte, error) {
	// Deprecated - use mixAudioComponentsWithNormalization instead
	return eim.mixAudioComponentsWithNormalization(ambiencePath, chimePath, ttsData)
}

// mixAudioComponentsWithNormalization combines ambience, chime, and voice with volume normalization
func (eim *EnhancedIntroMixer) mixAudioComponentsWithNormalization(ambiencePath, chimePath string, introData []byte) ([]byte, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH")
	}

	// Create temp files
	tempDir := os.TempDir()
	introFile := filepath.Join(tempDir, fmt.Sprintf("intro_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("intro_enhanced_%d.mp3", time.Now().Unix()))

	// Write intro data to temp file
	if err := os.WriteFile(introFile, introData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write intro file: %w", err)
	}
	defer os.Remove(introFile)
	defer os.Remove(outputFile)

	// Get intro duration
	introDuration := eim.getAudioDuration(introFile)
	if introDuration <= 0 {
		introDuration = 5.0 // Default fallback
	}

	// Calculate timings
	fadeInDuration := 2.5 // Ambience fade in duration
	chimeDelay := 2.0     // When chime plays (during fade in)
	voiceDelay := 3.0     // When voice starts (after fade in)
	fadeOutStart := voiceDelay + introDuration
	fadeOutDuration := 2.0
	totalDuration := voiceDelay + introDuration + fadeOutDuration

	// Build ffmpeg command for mixing with normalization
	// Note: We boost the intro volume to match dynamic TTS levels
	cmd := exec.Command("ffmpeg",
		"-i", ambiencePath, // Input: ambience
		"-i", chimePath, // Input: chime
		"-i", introFile, // Input: pre-recorded intro
		"-filter_complex",
		fmt.Sprintf(
			// Ambience: fade in, then duck to 15%% volume under voice
			"[0:a]afade=t=in:st=0:d=%.1f,volume=0.35[ambience_fade];"+
				"[ambience_fade]volume=0.15:enable='gte(t,%.1f)'[ambience_ducked];"+
				// Chime: delay and set volume
				"[1:a]adelay=%d|%d,volume=0.6[chime_delayed];"+
				// Intro: apply normalization to match dynamic TTS volume, then delay
				// Boost by ~6dB to match dynamic TTS levels
				"[2:a]volume=2.0,adelay=%d|%d[voice_delayed];"+
				// Mix all three
				"[ambience_ducked][chime_delayed][voice_delayed]amix=inputs=3:duration=longest:dropout_transition=0.5[mixed];"+
				// Apply loudnorm for consistent output levels
				"[mixed]loudnorm=I=-16:TP=-1.5:LRA=11[normalized];"+
				// Fade out at the end
				"[normalized]afade=t=out:st=%.1f:d=%.1f[out]",
			fadeInDuration,       // Ambience fade in duration
			voiceDelay,           // When to duck ambience
			int(chimeDelay*1000), // Chime delay (ms)
			int(chimeDelay*1000), // Chime delay for second channel
			int(voiceDelay*1000), // Voice delay (ms)
			int(voiceDelay*1000), // Voice delay for second channel
			fadeOutStart,         // When to start fade out
			fadeOutDuration,      // Fade out duration
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

	fmt.Printf("[ENHANCED_INTRO] Successfully created intro with %s ambience and volume normalization (size: %d bytes)\n",
		eim.selectedAmbience, len(mixedData))
	return mixedData, nil
}

// GetSelectedAmbience returns the ambience track used in the last generated intro
func (eim *EnhancedIntroMixer) GetSelectedAmbience() string {
	return eim.selectedAmbience
}

// GetAmbienceForBackground returns audio data for background ambience in next track
func (eim *EnhancedIntroMixer) GetAmbienceForBackground() ([]byte, error) {
	if eim.selectedAmbience == "" {
		return nil, fmt.Errorf("no ambience selected yet")
	}

	// Find the ambience file
	var ambiencePath string
	for _, amb := range eim.GetAvailableAmbiences() {
		if amb.Name == eim.selectedAmbience {
			ambiencePath = filepath.Join(eim.soundEffectsPath, amb.Path)
			break
		}
	}

	if ambiencePath == "" {
		return nil, fmt.Errorf("ambience path not found for: %s", eim.selectedAmbience)
	}

	// Read and return the ambience file
	return os.ReadFile(ambiencePath)
}

// getAudioDuration gets the duration of an audio file using ffprobe
func (eim *EnhancedIntroMixer) getAudioDuration(audioFile string) float64 {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioFile,
	)

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("[ENHANCED_INTRO] Failed to get audio duration: %v\n", err)
		return 0
	}

	duration := 0.0
	fmt.Sscanf(string(output), "%f", &duration)
	return duration
}
