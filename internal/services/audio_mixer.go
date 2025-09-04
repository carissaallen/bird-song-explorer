package services

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// AudioMixer handles mixing audio with background music or nature sounds
type AudioMixer struct {
	musicPath string
}

// NewAudioMixer creates a new audio mixer
func NewAudioMixer() *AudioMixer {
	// Try different possible paths for the music directory
	possiblePaths := []string{
		"/root/assets/music", // Docker container path
		"./assets/music",     // Local development path
		"assets/music",       // Relative path
	}

	musicPath := "./assets/music" // Default
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			musicPath = path
			break
		}
	}

	return &AudioMixer{
		musicPath: musicPath,
	}
}

// MixOutroWithMusic mixes the outro voice with background music
func (am *AudioMixer) MixOutroWithMusic(voiceData []byte, musicType string) ([]byte, error) {

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Printf("[AUDIO_MIXER] ffmpeg not found in PATH, returning voice only\n")
		return voiceData, nil
	}

	// Create temp files for processing
	tempDir := os.TempDir()
	voiceFile := filepath.Join(tempDir, fmt.Sprintf("outro_voice_%d.mp3", time.Now().Unix()))
	musicFile := am.selectBackgroundMusic(musicType)
	outputFile := filepath.Join(tempDir, fmt.Sprintf("outro_mixed_%d.mp3", time.Now().Unix()))

	// Write voice data to temp file
	if err := os.WriteFile(voiceFile, voiceData, 0644); err != nil {
		fmt.Printf("[AUDIO_MIXER] Failed to write voice file: %v\n", err)
		return nil, fmt.Errorf("failed to write voice file: %w", err)
	}
	defer os.Remove(voiceFile)
	defer os.Remove(outputFile)

	// Check if music file exists
	if _, err := os.Stat(musicFile); os.IsNotExist(err) {
		fmt.Printf("[AUDIO_MIXER] Music file not found: %s, returning voice only\n", musicFile)
		// List contents of music directory for debugging
		if entries, err := os.ReadDir(am.musicPath); err == nil {
			fmt.Printf("[AUDIO_MIXER] Contents of %s:\n", am.musicPath)
			for _, entry := range entries {
				fmt.Printf("[AUDIO_MIXER]   - %s\n", entry.Name())
			}
		} else {
			fmt.Printf("[AUDIO_MIXER] Could not read music directory: %v\n", err)
		}
		return voiceData, nil // Return voice only if no music available
	}

	// Mix audio using ffmpeg with simpler, more compatible settings:
	// - Music at moderate volume
	// - Fade in/out for smooth transitions
	// - Mix both tracks together
	cmd := exec.Command("ffmpeg",
		"-i", voiceFile, // Input: voice
		"-i", musicFile, // Input: music
		"-filter_complex",
		"[1:a]volume=0.25,afade=t=in:st=0:d=1,afade=t=out:st=28:d=2[music];"+ // Music with fades
			"[0:a]apad=whole_dur=30[voice];"+ // Pad voice to 30 seconds
			"[voice][music]amix=inputs=2:duration=longest:dropout_transition=3[out]", // Simple mix
		"-map", "[out]",
		"-t", "30", // Total duration 30 seconds
		"-c:a", "libmp3lame", // MP3 codec
		"-b:a", "128k", // Bitrate
		"-y", // Overwrite output
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If ffmpeg fails, return voice only
		fmt.Printf("[AUDIO_MIXER] ffmpeg mixing failed: %v\nStderr: %s\n", err, stderr.String())
		return voiceData, nil
	}

	// Read the mixed audio
	mixedData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read mixed audio: %w", err)
	}

	return mixedData, nil
}

// selectBackgroundMusic selects appropriate background music
func (am *AudioMixer) selectBackgroundMusic(musicType string) string {
	// Define available music tracks
	// These would be pre-loaded royalty-free children's music
	musicTracks := map[string][]string{
		"cheerful": {
			"outro_music_happy.mp3",
			"outro_music_ukulele.mp3",
			"outro_music_xylophone.mp3",
		},
		"calm": {
			"outro_music_lullaby.mp3",
			"outro_music_gentle.mp3",
		},
		"seasonal_summer": {
			"outro_music_summer.mp3",
		},
		"seasonal_winter": {
			"outro_music_winter.mp3",
		},
		"seasonal_spring": {
			"outro_music_spring.mp3",
		},
		"seasonal_autumn": {
			"outro_music_autumn.mp3",
		},
	}

	// Get season-appropriate music
	season := getCurrentSeason()
	seasonalKey := "seasonal_" + season

	// Check if we have seasonal music
	if tracks, exists := musicTracks[seasonalKey]; exists && len(tracks) > 0 {
		rand.Seed(time.Now().UnixNano())
		selected := tracks[rand.Intn(len(tracks))]
		return filepath.Join(am.musicPath, selected)
	}

	// Fall back to cheerful music
	if tracks, exists := musicTracks["cheerful"]; exists && len(tracks) > 0 {
		rand.Seed(time.Now().UnixNano())
		selected := tracks[rand.Intn(len(tracks))]
		return filepath.Join(am.musicPath, selected)
	}

	// Default fallback
	return filepath.Join(am.musicPath, "outro_music_default.mp3")
}

// DownloadAndCacheMusic downloads a music file from URL and caches it
func (am *AudioMixer) DownloadAndCacheMusic(url string, filename string) error {
	musicDir := am.musicPath

	// Create music directory if it doesn't exist
	if err := os.MkdirAll(musicDir, 0755); err != nil {
		return fmt.Errorf("failed to create music directory: %w", err)
	}

	outputPath := filepath.Join(musicDir, filename)

	// Check if already cached
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("[AUDIO_MIXER] Music already cached: %s\n", filename)
		return nil
	}

	// Download the file (placeholder - would need actual implementation)
	fmt.Printf("[AUDIO_MIXER] Would download music from %s to %s\n", url, outputPath)

	return nil
}

// MixOutroWithNatureSounds mixes the outro voice with bird song as background
func (am *AudioMixer) MixOutroWithNatureSounds(voiceData []byte, birdSongData []byte) ([]byte, error) {
	fmt.Printf("[AUDIO_MIXER] Starting nature sounds mixing process\n")

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Printf("[AUDIO_MIXER] ffmpeg not found in PATH, returning voice only\n")
		return voiceData, nil
	}

	// Create temp files for processing
	tempDir := os.TempDir()
	voiceFile := filepath.Join(tempDir, fmt.Sprintf("outro_voice_%d.mp3", time.Now().Unix()))
	birdFile := filepath.Join(tempDir, fmt.Sprintf("bird_song_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("outro_mixed_%d.mp3", time.Now().Unix()))

	fmt.Printf("[AUDIO_MIXER] Voice file: %s\n", voiceFile)
	fmt.Printf("[AUDIO_MIXER] Bird song file: %s\n", birdFile)
	fmt.Printf("[AUDIO_MIXER] Output file: %s\n", outputFile)

	// Write voice data to temp file
	if err := os.WriteFile(voiceFile, voiceData, 0644); err != nil {
		fmt.Printf("[AUDIO_MIXER] Failed to write voice file: %v\n", err)
		return nil, fmt.Errorf("failed to write voice file: %w", err)
	}
	defer os.Remove(voiceFile)

	// Write bird song data to temp file
	if err := os.WriteFile(birdFile, birdSongData, 0644); err != nil {
		fmt.Printf("[AUDIO_MIXER] Failed to write bird song file: %v\n", err)
		os.Remove(voiceFile)
		return nil, fmt.Errorf("failed to write bird song file: %w", err)
	}
	defer os.Remove(birdFile)
	defer os.Remove(outputFile)

	fmt.Printf("[AUDIO_MIXER] Files written, proceeding with mixing\n")

	// Mix audio using ffmpeg with volume normalization:
	// - Bird song at 15% volume while voice is playing
	// - Apply loudnorm for consistent levels with other tracks
	// - Loop bird song if it's shorter than the outro
	// - Fade in/out for smooth transitions
	cmd := exec.Command("ffmpeg",
		"-i", voiceFile, // Input: voice
		"-stream_loop", "-1", // Loop the bird song
		"-i", birdFile, // Input: bird song
		"-t", "30", // Limit to 30 seconds total
		"-filter_complex",
		"[1:a]volume=0.15,afade=t=in:st=0:d=1[bird_quiet];"+ // Bird song at low volume (matching intro)
			"[0:a]volume=2.2,apad=whole_dur=30[voice_boosted];"+ // Boost voice volume to match intro track
			"[voice_boosted][bird_quiet]amix=inputs=2:duration=longest[mixed];"+ // Mix them
			"[mixed]afade=t=out:st=28:d=2[out]", // Fade out at end - no loudnorm to preserve dynamics
		"-map", "[out]",
		"-c:a", "libmp3lame", // MP3 codec
		"-b:a", "192k", // Higher bitrate for better quality
		"-y", // Overwrite output
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If ffmpeg fails, return voice only
		fmt.Printf("[AUDIO_MIXER] ffmpeg mixing failed: %v\nStderr: %s\n", err, stderr.String())
		return voiceData, nil
	}

	// Read the mixed audio
	mixedData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read mixed audio: %w", err)
	}

	fmt.Printf("[AUDIO_MIXER] Successfully mixed outro with bird song (size: %d bytes)\n", len(mixedData))
	return mixedData, nil
}

// MixOutroWithAmbienceAndJingle mixes the outro voice with ambience and adds a ukulele jingle
func (am *AudioMixer) MixOutroWithAmbienceAndJingle(voiceData []byte, ambienceData []byte, ambienceName string) ([]byte, error) {
	fmt.Printf("[AUDIO_MIXER] Starting outro mixing with %s ambience and ukulele jingle\n", ambienceName)

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Printf("[AUDIO_MIXER] ffmpeg not found in PATH, returning voice only\n")
		return voiceData, nil
	}

	// Create temp files for processing
	tempDir := os.TempDir()
	voiceFile := filepath.Join(tempDir, fmt.Sprintf("outro_voice_%d.mp3", time.Now().Unix()))
	ambienceFile := filepath.Join(tempDir, fmt.Sprintf("outro_ambience_%d.mp3", time.Now().Unix()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("outro_mixed_%d.mp3", time.Now().Unix()))

	// Path to ukulele jingle
	ukuleleFile := "sound_effects/chimes/ukulele_short.mp3"

	fmt.Printf("[AUDIO_MIXER] Voice file: %s\n", voiceFile)
	fmt.Printf("[AUDIO_MIXER] Ambience file: %s\n", ambienceFile)
	fmt.Printf("[AUDIO_MIXER] Ukulele file: %s\n", ukuleleFile)
	fmt.Printf("[AUDIO_MIXER] Output file: %s\n", outputFile)

	// Write voice data to temp file
	if err := os.WriteFile(voiceFile, voiceData, 0644); err != nil {
		fmt.Printf("[AUDIO_MIXER] Failed to write voice file: %v\n", err)
		return nil, fmt.Errorf("failed to write voice file: %w", err)
	}
	defer os.Remove(voiceFile)

	// Write ambience data to temp file
	if err := os.WriteFile(ambienceFile, ambienceData, 0644); err != nil {
		fmt.Printf("[AUDIO_MIXER] Failed to write ambience file: %v\n", err)
		os.Remove(voiceFile)
		return nil, fmt.Errorf("failed to write ambience file: %w", err)
	}
	defer os.Remove(ambienceFile)
	defer os.Remove(outputFile)

	// Check if ukulele file exists
	if _, err := os.Stat(ukuleleFile); err != nil {
		fmt.Printf("[AUDIO_MIXER] Ukulele file not found: %s, mixing without jingle\n", ukuleleFile)
		// Fall back to mixing without jingle
		return am.mixOutroWithAmbienceOnly(voiceFile, ambienceFile, outputFile)
	}

	fmt.Printf("[AUDIO_MIXER] Files written, proceeding with mixing\n")

	// Mix audio using ffmpeg with volume normalization:
	// - Ambience at 15% volume while voice is playing
	// - Fade out ambience 1-2 seconds faster after voice ends
	// - Crossfade to ukulele jingle at the end
	// - Apply loudnorm for consistent levels
	cmd := exec.Command("ffmpeg",
		"-i", voiceFile, // Input 0: voice
		"-i", ambienceFile, // Input 1: ambience
		"-i", ukuleleFile, // Input 2: ukulele jingle
		"-filter_complex",
		// Ambience: low volume during voice, fade out faster (1 second instead of 2)
		"[1:a]volume=0.15,afade=t=in:st=0:d=1[ambience_low];"+
			// Get voice duration for timing
			"[0:a]apad=whole_dur=28[voice_padded];"+
			// Mix voice with ambience
			"[voice_padded][ambience_low]amix=inputs=2:duration=first:dropout_transition=1[voice_with_ambience];"+
			// Prepare ukulele with delay (starts sooner after voice)
			"[2:a]adelay=26500|26500,volume=0.8[ukulele_delayed];"+
			// Fade out voice+ambience mix faster before ukulele
			"[voice_with_ambience]afade=t=out:st=25.5:d=1[mix_fadeout];"+
			// Overlay ukulele at the end
			"[mix_fadeout][ukulele_delayed]amix=inputs=2:duration=longest[mixed];"+
			// Final fade out - no loudnorm to preserve dynamics
			"[mixed]afade=t=out:st=29.5:d=1[out]",
		"-map", "[out]",
		"-t", "30", // Total duration reduced by 2 seconds
		"-c:a", "libmp3lame", // MP3 codec
		"-b:a", "192k", // Higher bitrate for better quality
		"-y", // Overwrite output
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If ffmpeg fails, try simpler mixing
		fmt.Printf("[AUDIO_MIXER] Complex mixing failed: %v\nStderr: %s\n", err, stderr.String())
		fmt.Printf("[AUDIO_MIXER] Falling back to simpler mixing\n")
		return am.mixOutroWithAmbienceOnly(voiceFile, ambienceFile, outputFile)
	}

	// Read the mixed audio
	mixedData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read mixed audio: %w", err)
	}

	fmt.Printf("[AUDIO_MIXER] Successfully mixed outro with %s ambience and ukulele jingle (size: %d bytes)\n", ambienceName, len(mixedData))
	return mixedData, nil
}

// mixOutroWithAmbienceOnly is a fallback for when ukulele is not available
func (am *AudioMixer) mixOutroWithAmbienceOnly(voiceFile, ambienceFile, outputFile string) ([]byte, error) {
	cmd := exec.Command("ffmpeg",
		"-i", voiceFile, // Input: voice
		"-i", ambienceFile, // Input: ambience
		"-filter_complex",
		"[1:a]volume=0.15,afade=t=in:st=0:d=1[ambience_low];"+
			"[0:a]apad=whole_dur=30[voice_padded];"+
			"[voice_padded][ambience_low]amix=inputs=2:duration=longest[mixed];"+
			"[mixed]afade=t=out:st=28:d=2[out]", // No loudnorm to preserve dynamics
		"-map", "[out]",
		"-t", "30",
		"-c:a", "libmp3lame",
		"-b:a", "192k",
		"-y",
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("[AUDIO_MIXER] Fallback mixing failed: %v\n", err)
		return nil, fmt.Errorf("fallback mixing failed: %w", err)
	}

	mixedData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read mixed audio: %w", err)
	}

	return mixedData, nil
}

// GenerateSimpleOutroMusic generates a simple outro tune using sox or ffmpeg
func GenerateSimpleOutroMusic() ([]byte, error) {
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("outro_music_%d.mp3", time.Now().Unix()))
	defer os.Remove(tempFile)

	// Generate a simple 25-second tune using sox (if available) or ffmpeg
	// This creates a simple, cheerful melody using sine waves
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "sine=frequency=523:duration=0.25,sine=frequency=587:duration=0.25,sine=frequency=659:duration=0.25,sine=frequency=523:duration=0.25",
		"-filter_complex",
		"[0:a]volume=0.3,aecho=0.8:0.9:40:0.4[out]", // Add echo for musicality
		"-t", "25", // 25 seconds duration
		"-c:a", "libmp3lame",
		"-b:a", "64k", // Lower bitrate for background music
		"-y",
		tempFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("[AUDIO_MIXER] Failed to generate simple music: %v\n", err)
		return nil, fmt.Errorf("failed to generate music: %w", err)
	}

	musicData, err := os.ReadFile(tempFile)
	if err != nil {
		return nil, err
	}

	return musicData, nil
}
