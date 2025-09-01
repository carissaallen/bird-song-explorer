package services

import (
	"fmt"
	"os"
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

// GenerateOutroWithBirdSong creates the complete outro track
func (oi *OutroIntegration) GenerateOutroWithBirdSong(
	voiceName string,
	dayOfWeek time.Weekday,
	birdSongData []byte,
	baseURL string,
) ([]byte, error) {
	
	if !oi.useStatic {
		// Fall back to dynamic TTS generation (old method)
		return oi.generateDynamicOutro(voiceName, dayOfWeek, birdSongData)
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
	
	// Mix with bird song if available
	if birdSongData != nil && len(birdSongData) > 0 {
		fmt.Printf("[OUTRO] Mixing with bird song (%d bytes)\n", len(birdSongData))
		mixedAudio, err := oi.audioMixer.MixOutroWithNatureSounds(outroData, birdSongData)
		if err != nil {
			fmt.Printf("[OUTRO] Mixing failed, using outro only: %v\n", err)
			return outroData, nil
		}
		fmt.Printf("[OUTRO] Successfully mixed outro with bird song\n")
		return mixedAudio, nil
	}
	
	// Return outro without bird song if none available
	return outroData, nil
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
func (oi *OutroIntegration) generateDynamicOutro(voiceName string, dayOfWeek time.Weekday, birdSongData []byte) ([]byte, error) {
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