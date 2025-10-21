package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/gin-gonic/gin"
)

type StreamingSession struct {
	SessionID      string
	Location       *models.Location
	BirdName       string
	ScientificName string
	BirdAudioURL   string
	VoiceID        string
	CreatedAt      time.Time
}

var sessionStore = make(map[string]*StreamingSession)

func cleanupSessions() {
	for id, session := range sessionStore {
		if time.Since(session.CreatedAt) > 15*time.Minute {
			delete(sessionStore, id)
		}
	}
}

// CreateSessionForBird creates a new session for a specific bird
// This ensures the icon and narration match when the card is played
func (h *Handler) CreateSessionForBird(cardID string, birdName string) string {
	sessionID := fmt.Sprintf("%s_%d", cardID, time.Now().Unix())

	session := &StreamingSession{
		SessionID: sessionID,
		BirdName:  birdName,
		CreatedAt: time.Now(),
	}

	sessionStore[sessionID] = session
	log.Printf("[SESSION] Created session %s for bird: %s", sessionID, birdName)

	go cleanupSessions()

	return sessionID
}

// getOrCreateSession gets an existing session or creates a new one
// Uses the session ID from query parameter to maintain state across tracks
func (h *Handler) getOrCreateSession(c *gin.Context, sessionID string) *StreamingSession {
	clientIP := c.ClientIP()

	if sessionID != "" {
		if existingSession, exists := sessionStore[sessionID]; exists {
			if time.Since(existingSession.CreatedAt) > 15*time.Minute {
				log.Printf("[STREAMING] Session %s expired (age: %v), creating new one", sessionID, time.Since(existingSession.CreatedAt))
				delete(sessionStore, sessionID)
			} else {
				log.Printf("[STREAMING] Using existing session %s for bird: %s (age: %v)", sessionID, existingSession.BirdName, time.Since(existingSession.CreatedAt))
				return existingSession
			}
		}
	}

	sessionKey := sessionID
	if sessionKey == "" {
		sessionKey = fmt.Sprintf("ip_%s_%d", clientIP, time.Now().Unix())
	}

	newSession := &StreamingSession{
		SessionID: sessionKey,
		CreatedAt: time.Now(),
	}

	location, err := h.locationService.GetLocationFromIP(clientIP)
	if err == nil && location != nil {
		newSession.Location = location
	}

	sessionStore[newSession.SessionID] = newSession
	go cleanupSessions()
	return newSession
}

func (h *Handler) StreamIntro(c *gin.Context) {
	birdName := c.Query("bird")
	sessionID := c.Query("session")
	session := h.getOrCreateSession(c, sessionID)

	if birdName != "" && session.BirdName != birdName {
		session.BirdName = birdName
	} else if session.BirdName == "" {
		// Fallback: Get the current cycling bird
		bird := h.availableBirds.GetCyclingBird()
		if bird != nil {
			session.BirdName = bird.CommonName
			session.ScientificName = bird.ScientificName
			log.Printf("[STREAMING] intro: Using cycling bird: %s", bird.CommonName)
		} else {
			log.Printf("[STREAMING] intro: No bird available")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Bird content not ready yet. Please try again in a few minutes.",
			})
			return
		}
	}

	birdDir := strings.ToLower(strings.ReplaceAll(session.BirdName, " ", "_"))
	gcsURL := fmt.Sprintf("https://storage.googleapis.com/bird-song-explorer-audio/birds/%s/narration/intro.mp3", birdDir)

	sessionStore[session.SessionID] = session
	c.Header("X-Session-ID", session.SessionID)
	c.Redirect(http.StatusFound, gcsURL)
}

func (h *Handler) StreamBirdAnnouncement(c *gin.Context) {
	sessionID := c.Query("session")
	session := h.getOrCreateSession(c, sessionID)

	birdName := session.BirdName
	if birdName == "" {
		// Fallback: Get the current cycling bird
		bird := h.availableBirds.GetCyclingBird()
		if bird != nil {
			birdName = bird.CommonName
			session.BirdName = birdName
			sessionStore[session.SessionID] = session
			log.Printf("[STREAMING] announcement: Using cycling bird: %s", birdName)
		} else {
			log.Printf("[STREAMING] announcement: No bird available")
			c.Status(http.StatusBadRequest)
			return
		}
	}

	birdDir := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	gcsURL := fmt.Sprintf("https://storage.googleapis.com/bird-song-explorer-audio/birds/%s/narration/announcement.mp3", birdDir)

	c.Redirect(http.StatusFound, gcsURL)
}

func (h *Handler) StreamDescription(c *gin.Context) {
	sessionID := c.Query("session")
	session := h.getOrCreateSession(c, sessionID)

	birdName := session.BirdName
	if birdName == "" {
		// Fallback: Get the current cycling bird
		bird := h.availableBirds.GetCyclingBird()
		if bird != nil {
			birdName = bird.CommonName
			session.BirdName = birdName
			sessionStore[session.SessionID] = session
			log.Printf("[STREAMING] description: Using cycling bird: %s", birdName)
		} else {
			log.Printf("[STREAMING] description: No bird available")
			c.Status(http.StatusBadRequest)
			return
		}
	}

	birdDir := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	gcsURL := fmt.Sprintf("https://storage.googleapis.com/bird-song-explorer-audio/birds/%s/narration/description.mp3", birdDir)

	c.Redirect(http.StatusFound, gcsURL)
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

func (h *Handler) StreamOutro(c *gin.Context) {
	sessionID := c.Query("session")
	session := h.getOrCreateSession(c, sessionID)

	birdName := session.BirdName
	if birdName == "" {
		// Fallback: Get the current cycling bird
		bird := h.availableBirds.GetCyclingBird()
		if bird != nil {
			birdName = bird.CommonName
			session.BirdName = birdName
			sessionStore[session.SessionID] = session
			log.Printf("[STREAMING] outro: Using cycling bird: %s", birdName)
		} else {
			log.Printf("[STREAMING] outro: No bird available")
			c.Status(http.StatusBadRequest)
			return
		}
	}

	birdDir := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	gcsURL := fmt.Sprintf("https://storage.googleapis.com/bird-song-explorer-audio/birds/%s/narration/outro.mp3", birdDir)

	c.Redirect(http.StatusFound, gcsURL)
}

// StreamOutro_OLD - Previous implementation with local files and ukulele mixing
func (h *Handler) StreamOutro_OLD(c *gin.Context) {
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
