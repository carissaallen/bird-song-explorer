package config

import (
	"os"
	"strconv"
)

// AudioConfig contains audio-related configuration settings
type AudioConfig struct {
	UseNatureSounds    bool   // Enable nature sounds in intros
	NatureSoundVolume  float64 // Volume level for nature sounds (0.0 to 1.0)
	IntroDelaySeconds  float64 // Delay before voice starts in intro
	DefaultNatureSound string  // Default nature sound type
}

// GetAudioConfig returns the audio configuration from environment variables
func GetAudioConfig() *AudioConfig {
	config := &AudioConfig{
		UseNatureSounds:    true,  // Default to enabled
		NatureSoundVolume:  0.1,   // Default to 10% volume
		IntroDelaySeconds:  2.5,   // Default 2.5 second delay
		DefaultNatureSound: "",    // Empty means time-based selection
	}

	// Check environment variables
	if val := os.Getenv("USE_NATURE_SOUNDS"); val != "" {
		if use, err := strconv.ParseBool(val); err == nil {
			config.UseNatureSounds = use
		}
	}

	if val := os.Getenv("NATURE_SOUND_VOLUME"); val != "" {
		if volume, err := strconv.ParseFloat(val, 64); err == nil {
			if volume >= 0.0 && volume <= 1.0 {
				config.NatureSoundVolume = volume
			}
		}
	}

	if val := os.Getenv("INTRO_DELAY_SECONDS"); val != "" {
		if delay, err := strconv.ParseFloat(val, 64); err == nil {
			if delay >= 0.0 && delay <= 10.0 {
				config.IntroDelaySeconds = delay
			}
		}
	}

	if val := os.Getenv("DEFAULT_NATURE_SOUND"); val != "" {
		config.DefaultNatureSound = val
	}

	return config
}

// GetNatureSoundEnabled returns whether nature sounds are enabled
func GetNatureSoundEnabled() bool {
	return GetAudioConfig().UseNatureSounds
}