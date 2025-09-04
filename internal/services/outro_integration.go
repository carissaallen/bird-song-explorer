package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// OutroIntegration handles the complete outro flow with pre-recorded files
type OutroIntegration struct {
	staticManager *StaticOutroManager
	audioMixer    *AudioMixer
	useStatic     bool
}

// NewOutroIntegration creates a new outro integration service
func NewOutroIntegration() *OutroIntegration {
	// Check environment variable
	useStatic := os.Getenv("USE_STATIC_OUTROS") != "false" // Default to true
	
	return &OutroIntegration{
		staticManager: NewStaticOutroManager(),
		audioMixer:    NewAudioMixer(),
		useStatic:     useStatic,
	}
}

// GenerateOutroWithAmbience creates the complete outro track with ambient sounds
func (oi *OutroIntegration) GenerateOutroWithAmbience(
	voiceName string,
	dayOfWeek time.Weekday,
	ambienceData []byte,
	baseURL string,
) ([]byte, error) {
	
	if !oi.useStatic {
		// Fall back to dynamic TTS generation (old method)
		return oi.generateDynamicOutro(voiceName, dayOfWeek, ambienceData)
	}
	
	// Get the pre-recorded outro file path
	outroPath, err := oi.getStaticOutroPath(voiceName, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get outro path: %w", err)
	}
	
	// Read the pre-recorded outro
	outroData, err := os.ReadFile(outroPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read outro file: %w", err)
	}
	
	fmt.Printf("[OUTRO] Using pre-recorded outro: %s\n", filepath.Base(outroPath))
	
	// Mix with ambient sounds if available
	if ambienceData != nil && len(ambienceData) > 0 {
		fmt.Printf("[OUTRO] Mixing with ambient sounds (%d bytes)\n", len(ambienceData))
		mixedAudio, err := oi.mixOutroWithAmbience(outroData, ambienceData)
		if err != nil {
			fmt.Printf("[OUTRO] Mixing failed, applying volume boost only: %v\n", err)
			// Apply volume boost even if mixing fails
			return oi.applyVolumeBoost(outroData)
		}
		fmt.Printf("[OUTRO] Successfully mixed outro with ambient sounds\n")
		return mixedAudio, nil
	}
	
	// Apply volume boost to match intro track even without ambient sounds
	fmt.Printf("[OUTRO] No ambient sounds, applying volume boost to outro\n")
	return oi.applyVolumeBoost(outroData)
}

// getStaticOutroPath selects the appropriate pre-recorded outro file
func (oi *OutroIntegration) getStaticOutroPath(voiceName string, dayOfWeek time.Weekday) (string, error) {
	outroType := oi.getOutroType(dayOfWeek)
	
	// Find available outros of this type for this voice
	pattern := filepath.Join("final_outros", fmt.Sprintf("outro_%s_*_%s.mp3", outroType, voiceName))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("no outros found for %s/%s (pattern: %s)", outroType, voiceName, pattern)
	}
	
	// Select one deterministically based on date
	now := time.Now()
	daySeed := now.Year()*10000 + int(now.Month())*100 + now.Day()
	outroIndex := daySeed % len(matches)
	selectedFile := matches[outroIndex]
	
	fmt.Printf("[OUTRO] Selected: %s (type: %s, voice: %s, index: %d of %d)\n", 
		filepath.Base(selectedFile), outroType, voiceName, outroIndex, len(matches))
	
	return selectedFile, nil
}

// getOutroType determines which type of outro to use based on the day
func (oi *OutroIntegration) getOutroType(dayOfWeek time.Weekday) string {
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

// generateDynamicOutro is the fallback to TTS generation (old method)
func (oi *OutroIntegration) generateDynamicOutro(voiceName string, dayOfWeek time.Weekday, ambienceData []byte) ([]byte, error) {
	// This would call the original TTS-based outro generation
	// Kept as fallback if needed
	return nil, fmt.Errorf("dynamic outro generation not implemented - use static outros")
}

// GetOutroURL returns a URL to serve the pre-recorded outro
func (oi *OutroIntegration) GetOutroURL(voiceName string, dayOfWeek time.Weekday, baseURL string) (string, error) {
	outroPath, err := oi.getStaticOutroPath(voiceName, dayOfWeek)
	if err != nil {
		return "", err
	}
	
	// Return URL to the outro file
	filename := filepath.Base(outroPath)
	return fmt.Sprintf("%s/audio/outros/%s", baseURL, filename), nil
}

// ValidateOutros checks that all required outro files exist
func (oi *OutroIntegration) ValidateOutros() error {
	voices := []string{"Amelia", "Antoni", "Hope", "Rory", "Danielle", "Stuart"}
	types := []string{"joke", "wisdom", "teaser", "challenge", "funfact"}
	
	missingCount := 0
	for _, voice := range voices {
		for _, outroType := range types {
			pattern := filepath.Join("final_outros", fmt.Sprintf("outro_%s_*_%s.mp3", outroType, voice))
			matches, _ := filepath.Glob(pattern)
			if len(matches) == 0 {
				fmt.Printf("❌ Missing: %s outros for %s\n", outroType, voice)
				missingCount++
			} else {
				fmt.Printf("✅ Found %d %s outros for %s\n", len(matches), outroType, voice)
			}
		}
	}
	
	if missingCount > 0 {
		return fmt.Errorf("missing %d outro types", missingCount)
	}
	
	fmt.Println("✅ All outro files validated successfully!")
	return nil
}

// applyVolumeBoost applies a 2.2x volume boost to match the intro track
func (oi *OutroIntegration) applyVolumeBoost(audioData []byte) ([]byte, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Printf("[OUTRO] ffmpeg not found, returning original audio\n")
		return audioData, nil
	}

	// Create temp files
	tempDir := os.TempDir()
	inputFile := filepath.Join(tempDir, fmt.Sprintf("outro_input_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("outro_boosted_%d.mp3", time.Now().Unix()))

	// Write input data
	if err := os.WriteFile(inputFile, audioData, 0644); err != nil {
		return audioData, nil
	}
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// Apply volume boost to match intro (2.2x)
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-filter_complex",
		"[0:a]volume=2.2[boosted]",
		"-map", "[boosted]",
		"-c:a", "libmp3lame",
		"-b:a", "192k",
		"-ar", "44100",
		"-y",
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("[OUTRO] Volume boost failed: %v\n", err)
		return audioData, nil
	}

	// Read the boosted audio
	boostedData, err := os.ReadFile(outputFile)
	if err != nil {
		return audioData, nil
	}

	fmt.Printf("[OUTRO] Successfully applied 2.2x volume boost\n")
	return boostedData, nil
}

// mixOutroWithAmbience mixes the outro with ambient sounds at 15% volume
func (oi *OutroIntegration) mixOutroWithAmbience(outroData []byte, ambienceData []byte) ([]byte, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return oi.applyVolumeBoost(outroData)
	}

	// Create temp files
	tempDir := os.TempDir()
	outroFile := filepath.Join(tempDir, fmt.Sprintf("outro_voice_%d.mp3", time.Now().Unix()))
	ambienceFile := filepath.Join(tempDir, fmt.Sprintf("ambience_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("outro_mixed_%d.mp3", time.Now().Unix()))

	// Write files
	if err := os.WriteFile(outroFile, outroData, 0644); err != nil {
		return oi.applyVolumeBoost(outroData)
	}
	defer os.Remove(outroFile)

	if err := os.WriteFile(ambienceFile, ambienceData, 0644); err != nil {
		return oi.applyVolumeBoost(outroData)
	}
	defer os.Remove(ambienceFile)
	defer os.Remove(outputFile)

	// Mix with ambient sounds at 15% (matching intro) and apply 2.2x boost to voice
	cmd := exec.Command("ffmpeg",
		"-i", outroFile,
		"-stream_loop", "-1",
		"-i", ambienceFile,
		"-t", "30",
		"-filter_complex",
		"[1:a]volume=0.15,afade=t=in:st=0:d=1[ambience_quiet];"+
			"[0:a]volume=2.2,apad=whole_dur=30[voice_boosted];"+
			"[voice_boosted][ambience_quiet]amix=inputs=2:duration=longest[mixed];"+
			"[mixed]afade=t=out:st=28:d=2[out]",
		"-map", "[out]",
		"-c:a", "libmp3lame",
		"-b:a", "192k",
		"-y",
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("[OUTRO] Mixing failed: %v\nStderr: %s\n", err, stderr.String())
		return oi.applyVolumeBoost(outroData)
	}

	// Read the mixed audio
	mixedData, err := os.ReadFile(outputFile)
	if err != nil {
		return oi.applyVolumeBoost(outroData)
	}

	fmt.Printf("[OUTRO] Successfully mixed with ambient sounds and applied volume boost\n")
	return mixedData, nil
}