package api

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/gin-gonic/gin"
)

// StreamingSession stores location data for a play session
type StreamingSession struct {
	SessionID      string
	Location       *models.Location
	BirdName       string
	ScientificName string
	BirdAudioURL   string
	VoiceID        string
	CreatedAt      time.Time
}

// sessionStore is a simple in-memory store for sessions (could be Redis in production)
var sessionStore = make(map[string]*StreamingSession)

// cleanupSessions removes old sessions (older than 1 hour)
func cleanupSessions() {
	for id, session := range sessionStore {
		if time.Since(session.CreatedAt) > time.Hour {
			delete(sessionStore, id)
		}
	}
}

// getOrCreateSession gets an existing session or creates a new one
// Note: Caching removed - always creates fresh session for new bird selection each time
func (h *Handler) getOrCreateSession(c *gin.Context) *StreamingSession {
	// Always create a fresh session (no caching) - user gets new bird each time they play
	clientIP := c.ClientIP()
	sessionKey := fmt.Sprintf("ip_%s_%d", clientIP, time.Now().Unix())

	// Create new session
	newSession := &StreamingSession{
		SessionID: sessionKey,
		CreatedAt: time.Now(),
	}

	// Get location from IP
	location, err := h.locationService.GetLocationFromIP(clientIP)
	if err == nil && location != nil {
		newSession.Location = location
		log.Printf("[STREAMING] Created fresh session for IP %s: %s, %s",
			clientIP, location.City, location.Country)
	} else {
		log.Printf("[STREAMING] Created session for IP %s (location detection failed: %v)", clientIP, err)
	}

	// Store session temporarily for other tracks in this play session
	sessionStore[newSession.SessionID] = newSession
	go cleanupSessions()
	return newSession
}

// StreamIntro handles streaming requests for track 1 (intro)
func (h *Handler) StreamIntro(c *gin.Context) {
	birdName := c.Query("bird")
	sessionID := c.Query("session")
	session := h.getOrCreateSession(c)

	// Check if this is a new session or if we need a new bird
	needNewBird := false
	if session.BirdName == "" || session.BirdName != birdName {
		needNewBird = true
	}

	// If we need a new bird, select one and update the card
	if needNewBird {
		log.Printf("[STREAMING] New session detected, selecting fresh bird for IP %s", c.ClientIP())

		// Select a bird from the prerecorded birds only
		bird, err := h.selectPrerecordedBird(session.Location)
		if err != nil {
			log.Printf("[STREAMING] Failed to select prerecorded bird: %v", err)
			// Use a fallback bird
			bird = &models.Bird{
				CommonName:     "American Robin",
				ScientificName: "Turdus migratorius",
				Region:         "north_america",
			}
		}

		if bird != nil {
			session.BirdName = bird.CommonName
			session.ScientificName = bird.ScientificName
			session.BirdAudioURL = bird.AudioURL
			log.Printf("[STREAMING] Selected new bird: %s", bird.CommonName)

			// Try to update the card with new streaming URLs
			// Extract card ID from session parameter if available
			cardID := h.config.YotoCardID
			if sessionID != "" && strings.Contains(sessionID, "_") {
				parts := strings.Split(sessionID, "_")
				if len(parts) > 0 {
					cardID = parts[0]
				}
			}

			if cardID != "" {
				baseURL := h.config.BaseURL
				if baseURL == "" {
					baseURL = fmt.Sprintf("https://%s", c.Request.Host)
				}

				// Generate new session ID for the updated card
				newSessionID := fmt.Sprintf("%s_%d", cardID, time.Now().Unix())
				session.SessionID = newSessionID

				// Try to update the card (but don't fail if we can't)
				contentManager := h.yotoClient.NewContentManager()
				err := contentManager.UpdateCardWithStreamingTracks(cardID, bird.CommonName, baseURL, newSessionID)
				if err != nil {
					log.Printf("[STREAMING] Failed to update card (will continue with current bird): %v", err)
				} else {
					log.Printf("[STREAMING] Successfully updated card with new bird: %s", bird.CommonName)
				}
			}
		}
	} else if birdName != "" {
		session.BirdName = birdName
		log.Printf("[STREAMING] Session updated with bird: %s", birdName)
	}

	baseURL := h.config.BaseURL
	if baseURL == "" {
		// Fallback to request host if BASE_URL not configured
		baseURL = fmt.Sprintf("https://%s", c.Request.Host)
		if h.config.Environment == "development" {
			baseURL = fmt.Sprintf("http://%s", c.Request.Host)
		}
	}

	// Get intro URL and voice ID
	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)
	session.VoiceID = voiceID // Store for other tracks

	// Log the streaming request
	locationInfo := "no location"
	if session.Location != nil {
		locationInfo = fmt.Sprintf("%s, %s", session.Location.City, session.Location.Country)
	}
	log.Printf("[STREAMING] Intro requested - Session: %s, Bird: %s, Location: %s",
		session.SessionID, session.BirdName, locationInfo)

	// Store session for subsequent tracks
	sessionStore[session.SessionID] = session

	// Add session ID to response header so Yoto can pass it to next tracks
	c.Header("X-Session-ID", session.SessionID)

	// Extract the intro file path from the URL to normalize it
	// The URL format is: baseURL/audio/intros/filename
	if strings.Contains(introURL, "/audio/intros/") {
		parts := strings.Split(introURL, "/audio/intros/")
		if len(parts) > 1 {
			introFile := parts[1]
			introPath := fmt.Sprintf("assets/final_intros/%s", introFile)

			// Check if the file exists
			if _, err := os.Stat(introPath); err == nil {
				// Normalize and serve the intro
				h.normalizeAndServeAudio(c, introPath, "intro")
				return
			}
		}
	}

	// Fallback: If it's a local file, serve it directly
	if strings.Contains(introURL, c.Request.Host) {
		// Extract the path and redirect to it
		parts := strings.Split(introURL, c.Request.Host)
		if len(parts) > 1 {
			c.Redirect(http.StatusFound, parts[1])
			return
		}
	}

	// Otherwise redirect to the intro URL
	c.Redirect(http.StatusFound, introURL)
}

// StreamBirdAnnouncement handles streaming requests for track 2 (bird announcement)
func (h *Handler) StreamBirdAnnouncement(c *gin.Context) {
	birdName := c.Query("bird")
	if birdName == "" {
		log.Printf("[STREAMING] No bird name provided for announcement")
		c.Status(http.StatusBadRequest)
		return
	}

	// Try to find the prerecorded bird announcement file
	// The bird should have been selected from available prerecorded birds
	prerecordedPath := h.findPrerecordedBirdFile(birdName, "bird-announcement.mp3")

	if prerecordedPath != "" {
		log.Printf("[STREAMING] Using prerecorded announcement for %s: %s", birdName, prerecordedPath)

		// Construct the full file path
		fullPath := fmt.Sprintf("prerecorded_tts/audio/%s", prerecordedPath)

		// Check if the file exists
		if _, err := os.Stat(fullPath); err == nil {
			// Normalize and serve the announcement
			h.normalizeAndServeAudio(c, fullPath, "announcement")
			return
		}

		// Fallback to redirect if we can't process the file
		baseURL := h.config.BaseURL
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://%s", c.Request.Host)
			if h.config.Environment == "development" {
				baseURL = fmt.Sprintf("http://%s", c.Request.Host)
			}
		}

		// Redirect to the static file URL
		redirectURL := fmt.Sprintf("%s/audio/prerecorded/%s", baseURL, prerecordedPath)
		log.Printf("[STREAMING] Redirecting to prerecorded announcement: %s", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// Fallback: If no prerecorded file found, log error and return not found
	log.Printf("[STREAMING] ERROR: No prerecorded announcement found for bird: %s", birdName)
	c.Status(http.StatusNotFound)
}

// StreamBirdSong handles streaming requests for track 3 (bird song)
func (h *Handler) StreamBirdSong(c *gin.Context) {
	birdName := c.Query("bird")
	if birdName == "" {
		log.Printf("[STREAMING] No bird name provided for bird song")
		c.Status(http.StatusBadRequest)
		return
	}

	sessionID := c.Query("session")

	log.Printf("[STREAMING] Bird song requested - Bird: %s, Session: %s (fetching fresh from Xeno-Canto)", birdName, sessionID)

	// Always fetch fresh bird song from Xeno-Canto (no caching)
	bird, err := h.birdSelector.GetBirdByName(birdName)
	if err == nil && bird != nil && bird.AudioURL != "" {
		log.Printf("[STREAMING] Found audio for %s: %s", birdName, bird.AudioURL)
		h.proxyAudioContent(c, bird.AudioURL)
		return
	}

	// If we don't have the audio URL, we can't fetch it with just the common name
	// This should rarely happen as the webhook should have set the audio URL
	log.Printf("[STREAMING] No audio URL available for %s", birdName)

	// Try a fallback approach - use pre-known URL for common birds
	if fallbackURL := getFallbackBirdAudioURL(birdName); fallbackURL != "" {
		log.Printf("[STREAMING] Using fallback audio for %s", birdName)
		h.proxyAudioContent(c, fallbackURL)
		return
	}

	log.Printf("[STREAMING] No audio found for %s", birdName)
	c.Status(http.StatusNotFound)
}

// StreamDescription handles streaming requests for track 4 (bird guide)
func (h *Handler) StreamDescription(c *gin.Context) {
	// Get bird name from query parameter
	birdName := c.Query("bird")
	if birdName == "" {
		log.Printf("[STREAMING] No bird name provided for description")
		c.Status(http.StatusBadRequest)
		return
	}

	// Try to find the prerecorded bird guide file
	prerecordedPath := h.findPrerecordedBirdFile(birdName, "bird-guide.mp3")

	if prerecordedPath != "" {
		log.Printf("[STREAMING] Using prerecorded guide for %s: %s", birdName, prerecordedPath)

		// Construct the full file path
		fullPath := fmt.Sprintf("prerecorded_tts/audio/%s", prerecordedPath)

		// Check if the file exists
		if _, err := os.Stat(fullPath); err == nil {
			// Normalize and serve the guide
			h.normalizeAndServeAudio(c, fullPath, "guide")
			return
		}

		// Fallback to redirect if we can't process the file
		baseURL := h.config.BaseURL
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://%s", c.Request.Host)
			if h.config.Environment == "development" {
				baseURL = fmt.Sprintf("http://%s", c.Request.Host)
			}
		}

		// Redirect to the static file URL
		redirectURL := fmt.Sprintf("%s/audio/prerecorded/%s", baseURL, prerecordedPath)
		log.Printf("[STREAMING] Redirecting to prerecorded guide: %s", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// Fallback: If no prerecorded file found, log error and return not found
	log.Printf("[STREAMING] ERROR: No prerecorded guide found for bird: %s", birdName)
	c.Status(http.StatusNotFound)
}

// normalizeAndServeAudio normalizes an audio file to -16 LUFS and serves it
func (h *Handler) normalizeAndServeAudio(c *gin.Context, audioPath string, trackName string) {
	// Create a temporary file for the normalized audio
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s_normalized_*.mp3", trackName))
	if err != nil {
		log.Printf("[STREAMING] Failed to create temp file for %s: %v", trackName, err)
		// Fallback to serving the original file
		audioData, err := os.ReadFile(audioPath)
		if err != nil {
			log.Printf("[STREAMING] Failed to read %s file: %v", trackName, err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Header("Content-Type", "audio/mpeg")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, "audio/mpeg", audioData)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Use ffmpeg to normalize the audio to -16 LUFS for consistent volume
	cmd := exec.Command("ffmpeg",
		"-i", audioPath,
		"-af", "loudnorm=I=-16:TP=-1.5:LRA=11",
		"-codec:a", "libmp3lame",
		"-b:a", "128k",
		"-f", "mp3",
		"-y",
		tempFile.Name(),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[STREAMING] FFmpeg normalization error for %s: %v, output: %s", trackName, err, string(output))
		// Fallback to serving the original file
		audioData, err := os.ReadFile(audioPath)
		if err != nil {
			log.Printf("[STREAMING] Failed to read %s file: %v", trackName, err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Header("Content-Type", "audio/mpeg")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, "audio/mpeg", audioData)
		return
	}

	// Read the normalized audio file
	normalizedData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		log.Printf("[STREAMING] Failed to read normalized %s file: %v", trackName, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Serve the normalized audio
	log.Printf("[STREAMING] Serving normalized %s (-16 LUFS)", trackName)
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "audio/mpeg", normalizedData)
}

// StreamOutro handles streaming requests for track 5 (outro)
func (h *Handler) StreamOutro(c *gin.Context) {
	birdName := c.Query("bird")
	if birdName == "" {
		log.Printf("[STREAMING] No bird name provided for outro")
		c.Status(http.StatusBadRequest)
		return
	}

	log.Printf("[STREAMING] Outro requested - Bird: %s", birdName)

	// Check if we have pre-recorded outros - using only Amelia for consistency
	// We'll select from different categories for variety
	outroFiles := []string{
		"assets/final_outros/outro_joke_00_Amelia.mp3",
		"assets/final_outros/outro_joke_01_Amelia.mp3",
		"assets/final_outros/outro_wisdom_00_Amelia.mp3",
		"assets/final_outros/outro_wisdom_01_Amelia.mp3",
		"assets/final_outros/outro_teaser_00_Amelia.mp3",
		"assets/final_outros/outro_funfact_00_Amelia.mp3",
	}

	// Select an outro deterministically based on the day
	now := time.Now()
	daySeed := now.Year()*10000 + int(now.Month())*100 + now.Day()
	outroIndex := daySeed % len(outroFiles)
	selectedOutro := outroFiles[outroIndex]

	// Path to ukulele signoff
	ukulelePath := "assets/sound_effects/chimes/ukulele_short.mp3"

	// Combine the outro narration with ukulele signoff using ffmpeg
	// Create a temporary file for the combined audio
	tempFile, err := os.CreateTemp("", "outro_combined_*.mp3")
	if err != nil {
		log.Printf("[STREAMING] Failed to create temp file: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Use ffmpeg to concatenate the outro with the ukulele, normalizing volume
	// The filter complex:
	// 1. Normalizes the outro to -16 LUFS
	// 2. Normalizes the ukulele to -20 LUFS (slightly quieter for better blend)
	// 3. Concatenates them together
	cmd := exec.Command("ffmpeg",
		"-i", selectedOutro,
		"-i", ukulelePath,
		"-filter_complex",
		"[0:a]loudnorm=I=-16:TP=-1.5:LRA=11[a1];[1:a]loudnorm=I=-20:TP=-1.5:LRA=11[a2];[a1][a2]concat=n=2:v=0:a=1[out]",
		"-map", "[out]",
		"-codec:a", "libmp3lame",
		"-b:a", "128k",
		"-f", "mp3",
		"-y",
		tempFile.Name(),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[STREAMING] FFmpeg error: %v, output: %s", err, string(output))
		// Fallback to just the outro without ukulele
		outroData, err := os.ReadFile(selectedOutro)
		if err != nil {
			log.Printf("[STREAMING] Failed to read outro file: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Header("Content-Type", "audio/mpeg")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, "audio/mpeg", outroData)
		return
	}

	// Read the combined audio file
	combinedData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		log.Printf("[STREAMING] Failed to read combined file: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Serve the combined audio
	log.Printf("[STREAMING] Serving combined outro with ukulele signoff")
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "audio/mpeg", combinedData)
}

// proxyAudioContent fetches and streams audio from external URL
func (h *Handler) proxyAudioContent(c *gin.Context, audioURL string) {
	req, err := http.NewRequest("GET", audioURL, nil)
	if err != nil {
		log.Printf("[STREAMING] Failed to create request for %s: %v", audioURL, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Add user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BirdSongExplorer/1.0)")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[STREAMING] Failed to fetch audio from %s: %v", audioURL, err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[STREAMING] External audio returned status %d for %s", resp.StatusCode, audioURL)
		c.Status(resp.StatusCode)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// Default to audio/mpeg if not specified
		contentType = "audio/mpeg"
	}
	c.Header("Content-Type", contentType)
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Header("Content-Length", contentLength)
	}
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	log.Printf("[STREAMING] Proxying audio - Content-Type: %s, URL: %s", contentType, audioURL)

	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		log.Printf("[STREAMING] Error streaming audio: %v", err)
	}
}

// getFallbackBirdAudioURL returns a known audio URL for common birds
func getFallbackBirdAudioURL(birdName string) string {
	// Map of common birds to known working audio URLs
	fallbackURLs := map[string]string{
		"American Robin":    "https://www.bird-sounds.net/audio/american-robin.mp3",
		"Northern Cardinal": "https://www.bird-sounds.net/audio/northern-cardinal.mp3",
		"Blue Jay":          "https://www.bird-sounds.net/audio/blue-jay.mp3",
		// For now, return empty for birds without fallback
	}

	if url, exists := fallbackURLs[birdName]; exists {
		return url
	}
	return ""
}

// findPrerecordedBirdFile finds a prerecorded bird file for the given bird and file type
func (h *Handler) findPrerecordedBirdFile(birdName string, fileName string) string {
	// Convert bird name to directory format (lowercase with hyphens)
	birdDir := strings.ToLower(strings.ReplaceAll(birdName, " ", "-"))

	// Check different regions for this bird
	regions := []string{"north_america", "europe", "asia", "australia", "south_america", "central_america"}

	for _, region := range regions {
		filePath := fmt.Sprintf("%s/%s/%s", region, birdDir, fileName)
		fullPath := fmt.Sprintf("prerecorded_tts/audio/%s", filePath)

		// Check if file exists
		if _, err := os.Stat(fullPath); err == nil {
			log.Printf("[STREAMING] Found prerecorded file: %s", fullPath)
			return filePath
		}
	}

	log.Printf("[STREAMING] No prerecorded file found for %s/%s", birdDir, fileName)
	return ""
}

// selectPrerecordedBird selects a bird from available prerecorded birds based on location
func (h *Handler) selectPrerecordedBird(location *models.Location) (*models.Bird, error) {
	// Map of available prerecorded birds with their regions
	prerecordedBirds := map[string][]string{
		"north_america": {
			"American Robin",
			"Northern Cardinal",
			"Blue Jay",
			"Mourning Dove",
			"Black-capped Chickadee",
			"House Sparrow",
		},
		"europe": {
			"European Robin",
			"Great Tit",
			"Common Chaffinch",
			"Eurasian Blue Tit",
			"House Sparrow",
		},
		"asia": {
			"House Sparrow",
			"Great Tit",
			"Oriental Magpie-robin",
			"Japanese White-eye",
		},
		"australia": {
			"Australian Magpie",
			"Rainbow Lorikeet",
			"Kookaburra",
		},
		"south_america": {
			"Great Kiskadee",
			"Rufous Hornero",
		},
		"central_america": {
			"Great Kiskadee",
		},
	}

	// Determine the best region based on location
	var selectedRegion string
	var availableBirds []string

	if location != nil {
		// Map location to region
		if location.Country == "United States" || location.Country == "Canada" || location.Country == "Mexico" {
			selectedRegion = "north_america"
		} else if location.Country == "Australia" || location.Country == "New Zealand" {
			selectedRegion = "australia"
		} else if location.Country == "Brazil" || location.Country == "Argentina" || location.Country == "Chile" {
			selectedRegion = "south_america"
		} else if location.Region == "Europe" {
			selectedRegion = "europe"
		} else if location.Region == "Asia" {
			selectedRegion = "asia"
		} else {
			// Default to North America for unknown locations
			selectedRegion = "north_america"
		}
	} else {
		// Default region if no location
		selectedRegion = "north_america"
	}

	availableBirds = prerecordedBirds[selectedRegion]

	// Select a random bird from the region (new bird per card insertion)
	rand.Seed(time.Now().UnixNano())
	birdIndex := rand.Intn(len(availableBirds))
	selectedBird := availableBirds[birdIndex]

	// Get the bird's scientific name
	scientificNames := map[string]string{
		"American Robin":         "Turdus migratorius",
		"Northern Cardinal":      "Cardinalis cardinalis",
		"Blue Jay":               "Cyanocitta cristata",
		"Mourning Dove":          "Zenaida macroura",
		"Black-capped Chickadee": "Poecile atricapillus",
		"European Robin":         "Erithacus rubecula",
		"Great Tit":              "Parus major",
		"Common Chaffinch":       "Fringilla coelebs",
		"Eurasian Blue Tit":      "Cyanistes caeruleus",
		"House Sparrow":          "Passer domesticus",
		"Oriental Magpie-robin":  "Copsychus saularis",
		"Japanese White-eye":     "Zosterops japonicus",
		"Australian Magpie":      "Gymnorhina tibicen",
		"Rainbow Lorikeet":       "Trichoglossus moluccanus",
		"Kookaburra":             "Dacelo novaeguineae",
		"Great Kiskadee":         "Pitangus sulphuratus",
		"Rufous Hornero":         "Furnarius rufus",
	}

	scientificName := scientificNames[selectedBird]

	// Try to get the bird's audio URL from xeno-canto
	audioURL := ""
	if h.birdSelector != nil {
		if bird, err := h.birdSelector.GetBirdByName(selectedBird); err == nil && bird != nil {
			audioURL = bird.AudioURL
		}
	}

	log.Printf("[STREAMING] Randomly selected bird: %s (index %d of %d) from region %s", selectedBird, birdIndex, len(availableBirds), selectedRegion)

	return &models.Bird{
		CommonName:     selectedBird,
		ScientificName: scientificName,
		Region:         selectedRegion,
		AudioURL:       audioURL,
	}, nil
}
