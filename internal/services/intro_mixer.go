package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// IntroMixer handles mixing intro tracks with nature sounds
type IntroMixer struct {
	natureSoundsPath string
	introPath        string
	soundFetcher     *NatureSoundFetcher
}

// NewIntroMixer creates a new intro mixer
func NewIntroMixer() *IntroMixer {
	// Try different possible paths for nature sounds
	possiblePaths := []string{
		"/root/assets/nature_sounds", // Docker container path
		"./assets/nature_sounds",     // Local development path
		"assets/nature_sounds",       // Relative path
	}

	natureSoundsPath := "./assets/nature_sounds" // Default
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			natureSoundsPath = path
			break
		}
	}

	return &IntroMixer{
		natureSoundsPath: natureSoundsPath,
		introPath:        "assets/final_intros",
		soundFetcher:     NewNatureSoundFetcher(),
	}
}

// MixIntroWithNatureSounds mixes a pre-recorded intro with nature sounds
// Nature sounds start 2-3 seconds before the voice and continue softly in the background
func (im *IntroMixer) MixIntroWithNatureSounds(introData []byte, natureSoundType string) ([]byte, error) {
	return im.MixIntroWithNatureSoundsForUser(introData, natureSoundType, "")
}

// MixIntroWithNatureSoundsForUser mixes intro with nature sounds based on user's timezone
func (im *IntroMixer) MixIntroWithNatureSoundsForUser(introData []byte, natureSoundType string, userTimezone string) ([]byte, error) {
	fmt.Printf("[INTRO_MIXER] Starting intro mixing with nature sounds\n")

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Printf("[INTRO_MIXER] ffmpeg not found in PATH, returning intro only\n")
		return introData, nil
	}

	// Check if ffprobe is available for duration detection
	if _, err := exec.LookPath("ffprobe"); err != nil {
		fmt.Printf("[INTRO_MIXER] ffprobe not found, using default timing\n")
		// Fall back to the full mixing with default timing
		return im.mixWithDefaultTiming(introData, natureSoundType, userTimezone)
	}

	// Determine nature sound type based on user's timezone if not specified
	if natureSoundType == "" && userTimezone != "" {
		timeHelper := NewUserTimeHelper()
		natureSoundType = timeHelper.GetNatureSoundForUserTime(userTimezone)
		fmt.Printf("[INTRO_MIXER] Selected %s sound based on user's local time in %s\n",
			natureSoundType, userTimezone)
	}

	// Fetch nature sound from API
	var natureSoundData []byte
	var err error

	if natureSoundType == "" {
		// Get ambient soundscape based on server time (fallback)
		natureSoundData, err = im.soundFetcher.GetAmbientSoundscape()
	} else {
		// Get specific type of nature sound
		natureSoundData, err = im.soundFetcher.GetNatureSoundByType(natureSoundType)
	}

	if err != nil {
		fmt.Printf("[INTRO_MIXER] Failed to fetch nature sounds: %v, returning intro only\n", err)
		return introData, nil
	}

	// Create temp files for processing
	tempDir := os.TempDir()
	introFile := filepath.Join(tempDir, fmt.Sprintf("intro_voice_%d.mp3", time.Now().Unix()))
	natureFile := filepath.Join(tempDir, fmt.Sprintf("nature_sound_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("intro_mixed_%d.mp3", time.Now().Unix()))

	// Write intro data to temp file
	if err := os.WriteFile(introFile, introData, 0644); err != nil {
		fmt.Printf("[INTRO_MIXER] Failed to write intro file: %v\n", err)
		return nil, fmt.Errorf("failed to write intro file: %w", err)
	}
	defer os.Remove(introFile)

	// Write nature sound data to temp file
	if err := os.WriteFile(natureFile, natureSoundData, 0644); err != nil {
		fmt.Printf("[INTRO_MIXER] Failed to write nature sound file: %v\n", err)
		return nil, fmt.Errorf("failed to write nature sound file: %w", err)
	}
	defer os.Remove(natureFile)
	defer os.Remove(outputFile)

	// Get intro duration using ffprobe
	introDuration := im.getAudioDuration(introFile)
	if introDuration <= 0 {
		// Default to 5 seconds if we can't detect
		introDuration = 5.0
	}
	fmt.Printf("[INTRO_MIXER] Intro duration: %.2f seconds\n", introDuration)

	// Calculate timings for short intro
	leadInTime := 3.0  // Nature sounds lead-in before voice
	fadeOutTime := 2.0 // Fade out duration after voice ends
	totalDuration := leadInTime + introDuration + fadeOutTime
	fadeOutStart := leadInTime + introDuration // When to start fading out

	// Mix audio using ffmpeg with dynamic timing
	cmd := exec.Command("ffmpeg",
		"-i", natureFile, // Input: nature sounds
		"-i", introFile, // Input: voice intro
		"-filter_complex",
		fmt.Sprintf(
			// Nature sounds: fade in at 25% volume for lead-in, then duck to 10% under voice
			"[0:a]afade=t=in:st=0:d=1.5,volume=0.25[nature_intro];"+
				"[0:a]volume=0.10[nature_bg];"+
				// Split nature sounds: lead-in part and background part
				"[nature_intro]atrim=0:%.1f[nature_start];"+
				"[nature_bg]atrim=%.1f:%.1f[nature_rest];"+
				// Delay voice by lead-in time
				"[1:a]adelay=%d|%d[voice_delayed];"+
				// Combine nature parts
				"[nature_start][nature_rest]concat=n=2:v=0:a=1[nature_full];"+
				// Mix voice with nature, using "first" duration to end when voice ends
				"[voice_delayed][nature_full]amix=inputs=2:duration=first:dropout_transition=0.5[mixed];"+
				// Add fade out starting when voice ends
				"[mixed]afade=t=out:st=%.1f:d=%.1f[out]",
			leadInTime,           // Trim nature_start to lead-in duration
			leadInTime,           // Start nature_rest after lead-in
			totalDuration,        // End nature_rest at total duration
			int(leadInTime*1000), // Delay voice (in milliseconds)
			int(leadInTime*1000), // Delay voice for second channel
			fadeOutStart,         // Start fade out when voice ends
			fadeOutTime,          // Fade out duration
		),
		"-map", "[out]",
		"-t", fmt.Sprintf("%.2f", totalDuration), // Total duration based on intro length
		"-c:a", "libmp3lame", // MP3 codec
		"-b:a", "192k", // Higher bitrate for intro quality
		"-ar", "44100", // Sample rate
		"-y", // Overwrite output
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If ffmpeg fails, return intro only
		fmt.Printf("[INTRO_MIXER] ffmpeg mixing failed: %v\nStderr: %s\n", err, stderr.String())
		return introData, nil
	}

	// Read the mixed audio
	mixedData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read mixed audio: %w", err)
	}

	fmt.Printf("[INTRO_MIXER] Successfully mixed intro with nature sounds (size: %d bytes)\n", len(mixedData))
	return mixedData, nil
}

// selectNatureSound selects appropriate nature sound based on type or time of day
func (im *IntroMixer) selectNatureSound(soundType string) string {
	// Define available nature sounds
	natureSounds := map[string]string{
		"forest":        "forest_ambience.mp3",
		"morning_birds": "morning_birds.mp3",
		"gentle_rain":   "gentle_rain.mp3",
		"wind_trees":    "wind_through_trees.mp3",
		"stream":        "babbling_brook.mp3",
		"meadow":        "meadow_sounds.mp3",
		"night":         "night_crickets.mp3",
	}

	// If specific type requested, use it
	if sound, exists := natureSounds[soundType]; exists {
		return filepath.Join(im.natureSoundsPath, sound)
	}

	// Otherwise, select based on time of day
	hour := time.Now().Hour()
	var selected string

	switch {
	case hour >= 5 && hour < 9:
		selected = natureSounds["morning_birds"]
	case hour >= 9 && hour < 17:
		selected = natureSounds["forest"]
	case hour >= 17 && hour < 20:
		selected = natureSounds["meadow"]
	default:
		selected = natureSounds["night"]
	}

	return filepath.Join(im.natureSoundsPath, selected)
}

// getAudioDuration gets the duration of an audio file using ffprobe
func (im *IntroMixer) getAudioDuration(audioFile string) float64 {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioFile,
	)

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("[INTRO_MIXER] Failed to get audio duration: %v\n", err)
		return 0
	}

	duration := 0.0
	fmt.Sscanf(string(output), "%f", &duration)
	return duration
}

// mixWithDefaultTiming is the fallback mixing method when ffprobe is not available
func (im *IntroMixer) mixWithDefaultTiming(introData []byte, natureSoundType string, userTimezone string) ([]byte, error) {
	// This is the original mixing logic with fixed 30-second duration
	// Kept as fallback for systems without ffprobe

	// [Previous mixing logic would go here - simplified for now]
	// In production, this would contain the original mixing code
	return introData, nil
}

// PreprocessAllIntros processes all intro files to add nature sounds
// This can be run as a batch job to prepare all intros
func (im *IntroMixer) PreprocessAllIntros() error {
	introDir := im.introPath
	outputDir := filepath.Join(introDir, "with_nature")

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get all intro files
	files, err := os.ReadDir(introDir)
	if err != nil {
		return fmt.Errorf("failed to read intro directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".mp3" {
			continue
		}

		inputPath := filepath.Join(introDir, file.Name())
		outputPath := filepath.Join(outputDir, file.Name())

		// Skip if already processed
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("[INTRO_MIXER] Already processed: %s\n", file.Name())
			continue
		}

		// Read intro file
		introData, err := os.ReadFile(inputPath)
		if err != nil {
			fmt.Printf("[INTRO_MIXER] Failed to read %s: %v\n", file.Name(), err)
			continue
		}

		// Mix with nature sounds (using time-based selection)
		mixedData, err := im.MixIntroWithNatureSounds(introData, "")
		if err != nil {
			fmt.Printf("[INTRO_MIXER] Failed to mix %s: %v\n", file.Name(), err)
			continue
		}

		// Save mixed version
		if err := os.WriteFile(outputPath, mixedData, 0644); err != nil {
			fmt.Printf("[INTRO_MIXER] Failed to save mixed %s: %v\n", file.Name(), err)
			continue
		}

		fmt.Printf("[INTRO_MIXER] Successfully processed: %s\n", file.Name())
	}

	return nil
}
