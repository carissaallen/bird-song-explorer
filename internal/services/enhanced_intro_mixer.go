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
	soundEffectsPath   string
	cacheDir          string
	selectedAmbience  string // Track which ambience was used for passing to next track
	narrationManager  *NarrationManager
}

// NewEnhancedIntroMixer creates a new enhanced intro mixer
func NewEnhancedIntroMixer(elevenLabsKey string) *EnhancedIntroMixer {
	return &EnhancedIntroMixer{
		soundEffectsPath: "sound_effects",
		cacheDir:        "./audio_cache/enhanced_intros",
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

	// Generate TTS with custom text
	voiceManager := config.NewVoiceManager()
	voice := VoiceConfig{VoiceID: voiceID}
	if voiceID == "" {
		// Use daily voice if not specified
		dailyVoice := voiceManager.GetDailyVoice()
		voice = VoiceConfig{
			Name:    dailyVoice.Name,
			VoiceID: dailyVoice.ID,
		}
	}

	eim.narrationManager.selectedVoice = voice
	ttsData, err := eim.narrationManager.generateAudio(text, voice)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TTS: %w", err)
	}

	// Mix the audio components
	mixedAudio, err := eim.mixAudioComponents(ambiencePath, chimePath, ttsData)
	if err != nil {
		return nil, fmt.Errorf("failed to mix audio components: %w", err)
	}

	return mixedAudio, nil
}

// mixAudioComponents combines ambience, chime, and voice into a single track
func (eim *EnhancedIntroMixer) mixAudioComponents(ambiencePath, chimePath string, ttsData []byte) ([]byte, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH")
	}

	// Create temp files
	tempDir := os.TempDir()
	ttsFile := filepath.Join(tempDir, fmt.Sprintf("tts_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("intro_enhanced_%d.mp3", time.Now().Unix()))

	// Write TTS data to temp file
	if err := os.WriteFile(ttsFile, ttsData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write TTS file: %w", err)
	}
	defer os.Remove(ttsFile)
	defer os.Remove(outputFile)

	// Get TTS duration
	ttsDuration := eim.getAudioDuration(ttsFile)
	if ttsDuration <= 0 {
		ttsDuration = 5.0 // Default fallback
	}

	// Calculate timings
	fadeInDuration := 2.5    // Ambience fade in duration
	chimeDelay := 2.0        // When chime plays (during fade in)
	voiceDelay := 3.0        // When voice starts (after fade in)
	fadeOutStart := voiceDelay + ttsDuration
	fadeOutDuration := 2.0
	totalDuration := voiceDelay + ttsDuration + fadeOutDuration

	// Build ffmpeg command for mixing
	cmd := exec.Command("ffmpeg",
		"-i", ambiencePath,  // Input: ambience
		"-i", chimePath,     // Input: chime
		"-i", ttsFile,       // Input: TTS voice
		"-filter_complex",
		fmt.Sprintf(
			// Ambience: fade in, then duck to 15%% volume under voice
			"[0:a]afade=t=in:st=0:d=%.1f,volume=0.35[ambience_fade];"+
			"[ambience_fade]volume=0.15:enable='gte(t,%.1f)'[ambience_ducked];"+
			// Chime: delay and set volume
			"[1:a]adelay=%d|%d,volume=0.6[chime_delayed];"+
			// Voice: delay
			"[2:a]adelay=%d|%d[voice_delayed];"+
			// Mix all three
			"[ambience_ducked][chime_delayed][voice_delayed]amix=inputs=3:duration=longest:dropout_transition=0.5[mixed];"+
			// Fade out at the end
			"[mixed]afade=t=out:st=%.1f:d=%.1f[out]",
			fadeInDuration,                // Ambience fade in duration
			voiceDelay,                    // When to duck ambience
			int(chimeDelay*1000),          // Chime delay (ms)
			int(chimeDelay*1000),          // Chime delay for second channel
			int(voiceDelay*1000),          // Voice delay (ms)
			int(voiceDelay*1000),          // Voice delay for second channel
			fadeOutStart,                  // When to start fade out
			fadeOutDuration,               // Fade out duration
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

	fmt.Printf("[ENHANCED_INTRO] Successfully created intro with %s ambience (size: %d bytes)\n", 
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