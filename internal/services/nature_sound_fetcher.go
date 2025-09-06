package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NatureSoundFetcher fetches ambient nature sounds from Xeno-canto
type NatureSoundFetcher struct {
	cacheDir string
	client   *http.Client
}

// NewNatureSoundFetcher creates a new nature sound fetcher
func NewNatureSoundFetcher() *NatureSoundFetcher {
	return &NatureSoundFetcher{
		cacheDir: "audio_cache/nature_sounds",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// XenoCantoNatureResponse represents the response for nature/ambient sounds
type XenoCantoNatureResponse struct {
	Recordings []XenoCantoNatureRecording `json:"recordings"`
}

// XenoCantoNatureRecording represents a nature sound recording
type XenoCantoNatureRecording struct {
	ID       string      `json:"id"`
	Gen      string      `json:"gen"`       // Genus
	Sp       string      `json:"sp"`        // Species
	En       string      `json:"en"`        // English name
	Rec      string      `json:"rec"`       // Recordist
	Type     string      `json:"type"`      // Type of recording
	File     string      `json:"file"`      // Audio file URL
	FileName string      `json:"file-name"` // Original filename
	Length   string      `json:"length"`    // Duration
	Time     string      `json:"time"`      // Time of day recorded
	Also     interface{} `json:"also"`      // Other species in recording (can be string or array)
	Rmk      string      `json:"rmk"`       // Remarks
	Q        string      `json:"q"`         // Quality rating
}

// GetNatureSoundByType fetches nature sounds based on type
func (nsf *NatureSoundFetcher) GetNatureSoundByType(soundType string) ([]byte, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(nsf.cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Check cache first
	cacheFile := filepath.Join(nsf.cacheDir, fmt.Sprintf("%s.mp3", soundType))
	if data, err := nsf.checkCache(cacheFile); err == nil {
		fmt.Printf("[NATURE_FETCHER] Using cached nature sound: %s\n", soundType)
		return data, nil
	}

	// Map sound types to search queries
	queries := nsf.getSoundTypeQuery(soundType)

	fmt.Printf("[NATURE_FETCHER] Fetching nature sounds for type: %s\n", soundType)

	// Try each query until we find suitable recordings
	for _, query := range queries {
		recordings, err := nsf.searchXenoCanto(query)
		if err != nil {
			fmt.Printf("[NATURE_FETCHER] Error searching for %s: %v\n", query, err)
			continue
		}

		if len(recordings) > 0 {
			// Select a high-quality recording
			selected := nsf.selectBestRecording(recordings)
			if selected != nil {
				// Download the audio
				audioData, err := nsf.downloadAudio(selected.File)
				if err != nil {
					fmt.Printf("[NATURE_FETCHER] Error downloading audio: %v\n", err)
					continue
				}

				// Cache the audio
				os.WriteFile(cacheFile, audioData, 0644)

				fmt.Printf("[NATURE_FETCHER] Successfully fetched nature sound: %s (%s)\n",
					soundType, selected.En)
				return audioData, nil
			}
		}
	}

	return nil, fmt.Errorf("no suitable nature sounds found for type: %s", soundType)
}

// getSoundTypeQuery maps sound types to Xeno-canto search queries
func (nsf *NatureSoundFetcher) getSoundTypeQuery(soundType string) []string {
	switch soundType {
	case "forest":
		// Forest ambience - look for dawn chorus or forest recordings
		return []string{
			"type:dawn chorus",
			"type:soundscape forest",
			"rmk:ambient forest",
		}
	case "morning_birds":
		// Dawn chorus
		return []string{
			"type:dawn chorus",
			"time:05-08",
			"rmk:morning chorus",
		}
	case "gentle_rain":
		// Rain sounds - look for recordings with rain in remarks
		return []string{
			"rmk:rain",
			"rmk:light rain",
			"rmk:drizzle",
		}
	case "wind_trees":
		// Wind sounds
		return []string{
			"rmk:wind",
			"rmk:windy",
			"rmk:breeze",
		}
	case "stream":
		// Water sounds
		return []string{
			"rmk:stream",
			"rmk:creek",
			"rmk:water",
			"rmk:river",
		}
	case "meadow":
		// Open field sounds - insects and distant birds
		return []string{
			"type:soundscape meadow",
			"rmk:grassland",
			"rmk:field",
			"rmk:meadow",
		}
	case "night":
		// Night sounds - owls, crickets
		return []string{
			"type:nocturnal",
			"time:20-04",
			"rmk:night",
			"gen:Strix", // Owls
		}
	default:
		// Default to general soundscapes
		return []string{
			"type:soundscape",
			"type:dawn chorus",
		}
	}
}

// searchXenoCanto searches the Xeno-canto API
func (nsf *NatureSoundFetcher) searchXenoCanto(query string) ([]XenoCantoNatureRecording, error) {
	// Build the API URL
	baseURL := "https://www.xeno-canto.org/api/2/recordings"
	params := url.Values{}
	params.Set("query", query)

	// Add quality filter - only high quality recordings
	if !strings.Contains(query, "q:") {
		params.Set("query", query+" q:A")
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make the request
	resp, err := nsf.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query Xeno-canto: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var response XenoCantoNatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Recordings, nil
}

// selectBestRecording selects the best recording from the results
func (nsf *NatureSoundFetcher) selectBestRecording(recordings []XenoCantoNatureRecording) *XenoCantoNatureRecording {
	if len(recordings) == 0 {
		return nil
	}

	// Filter for high quality recordings (A or B quality)
	var highQuality []XenoCantoNatureRecording
	for _, rec := range recordings {
		if rec.Q == "A" || rec.Q == "B" {
			// Also check if it's not too short (at least 20 seconds)
			if nsf.isDurationSufficient(rec.Length) {
				highQuality = append(highQuality, rec)
			}
		}
	}

	if len(highQuality) == 0 {
		// Fall back to any recording if no high quality ones
		highQuality = recordings
	}

	// Select randomly from high quality recordings
	rand.Seed(time.Now().UnixNano())
	selected := highQuality[rand.Intn(len(highQuality))]

	return &selected
}

// isDurationSufficient checks if the recording is long enough
func (nsf *NatureSoundFetcher) isDurationSufficient(duration string) bool {
	// Duration is in format "0:30" or "1:45"
	parts := strings.Split(duration, ":")
	if len(parts) != 2 {
		return true // Can't parse, assume it's okay
	}

	// Convert to seconds
	minutes := 0
	seconds := 0
	fmt.Sscanf(parts[0], "%d", &minutes)
	fmt.Sscanf(parts[1], "%d", &seconds)

	totalSeconds := minutes*60 + seconds
	return totalSeconds >= 20 // At least 20 seconds
}

// downloadAudio downloads the audio file
func (nsf *NatureSoundFetcher) downloadAudio(audioURL string) ([]byte, error) {
	// Xeno-canto URLs need to be modified
	// They provide URLs like "//www.xeno-canto.org/sounds/..."
	if strings.HasPrefix(audioURL, "//") {
		audioURL = "https:" + audioURL
	}

	resp, err := nsf.client.Get(audioURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return data, nil
}

// checkCache checks if a cached version exists and is recent
func (nsf *NatureSoundFetcher) checkCache(cacheFile string) ([]byte, error) {
	info, err := os.Stat(cacheFile)
	if err != nil {
		return nil, err
	}

	// Use cache if less than 7 days old
	if time.Since(info.ModTime()) > 7*24*time.Hour {
		return nil, fmt.Errorf("cache expired")
	}

	return os.ReadFile(cacheFile)
}

// GetAmbientSoundscape fetches a general ambient soundscape
func (nsf *NatureSoundFetcher) GetAmbientSoundscape() ([]byte, error) {
	// Based on time of day, select appropriate soundscape
	hour := time.Now().Hour()

	var soundType string
	switch {
	case hour >= 5 && hour < 9:
		soundType = "morning_birds"
	case hour >= 9 && hour < 17:
		soundType = "forest"
	case hour >= 17 && hour < 20:
		soundType = "meadow"
	default:
		soundType = "night"
	}

	return nsf.GetNatureSoundByType(soundType)
}
