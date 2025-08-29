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

	// Mix audio using ffmpeg:
	// - Bird song at 15% volume while voice is playing
	// - Bird song at 40% volume after voice ends
	// - Loop bird song if it's shorter than the outro
	// - Fade in/out for smooth transitions
	cmd := exec.Command("ffmpeg",
		"-i", voiceFile, // Input: voice
		"-stream_loop", "-1", // Loop the bird song
		"-i", birdFile, // Input: bird song
		"-t", "30", // Limit to 30 seconds total
		"-filter_complex",
		"[1:a]volume=0.15,afade=t=in:st=0:d=1[bird_quiet];"+ // Bird song at low volume
			"[0:a]apad=whole_dur=30[voice];"+ // Pad voice to 30 seconds
			"[voice][bird_quiet]amix=inputs=2:duration=longest[mixed];"+ // Mix them
			"[mixed]afade=t=out:st=28:d=2[out]", // Fade out at end
		"-map", "[out]",
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

	fmt.Printf("[AUDIO_MIXER] Successfully mixed outro with bird song (size: %d bytes)\n", len(mixedData))
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
