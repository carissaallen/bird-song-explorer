package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
func (h *Handler) getOrCreateSession(c *gin.Context) *StreamingSession {
	// Use IP address as session key for consistency across tracks
	clientIP := c.ClientIP()
	sessionKey := fmt.Sprintf("ip_%s", clientIP)

	// Check if we have a recent session for this IP
	if session, exists := sessionStore[sessionKey]; exists {
		// Check if session is not too old (10 minutes for a play session)
		if time.Since(session.CreatedAt) < 10*time.Minute {
			log.Printf("[STREAMING] Using existing session for IP %s", clientIP)
			return session
		}
	}

	// Create new session
	newSession := &StreamingSession{
		SessionID: sessionKey,
		CreatedAt: time.Now(),
	}

	// Get location from IP
	location, err := h.locationService.GetLocationFromIP(clientIP)
	if err == nil && location != nil {
		newSession.Location = location
		log.Printf("[STREAMING] Detected location from IP %s: %s, %s",
			clientIP, location.City, location.Country)
	} else {
		log.Printf("[STREAMING] Failed to detect location from IP %s: %v", clientIP, err)
	}

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

		// Select a new bird based on location
		var bird *models.Bird
		var err error
		if session.Location != nil {
			bird, err = h.birdSelector.SelectBirdOfDay(session.Location)
			if err != nil {
				log.Printf("[STREAMING] Failed to select regional bird: %v", err)
			}
		}

		// Fall back to global bird if needed
		if bird == nil {
			globalLocation := &models.Location{
				Latitude:  40.7128,
				Longitude: -74.0060,
				City:      "Global",
				Region:    "Global",
				Country:   "Global",
			}
			bird, err = h.birdSelector.SelectBirdOfDay(globalLocation)
			if err != nil {
				log.Printf("[STREAMING] Failed to select global bird: %v", err)
				// Use the bird from the URL as last resort
				if birdName != "" {
					session.BirdName = birdName
				}
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

	// If it's a local file, serve it directly
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

	_, voiceID := h.audioManager.GetRandomIntroURL("")

	log.Printf("[STREAMING] Bird announcement requested - Bird: %s, Voice: %s",
		birdName, voiceID)

	audioData, err := h.audioManager.GenerateBirdAnnouncement(birdName, voiceID)
	if err != nil {
		log.Printf("[STREAMING] Failed to generate announcement: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Set appropriate headers for streaming
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Length", fmt.Sprintf("%d", len(audioData)))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	// Stream the audio
	c.Data(http.StatusOK, "audio/mpeg", audioData)
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

	log.Printf("[STREAMING] Bird song requested - Bird: %s, Session: %s", birdName, sessionID)

	if sessionID != "" {
		if session, exists := sessionStore[sessionID]; exists && session.BirdAudioURL != "" {
			log.Printf("[STREAMING] Using audio URL from session %s for %s: %s", sessionID, birdName, session.BirdAudioURL)
			h.proxyAudioContent(c, session.BirdAudioURL)
			return
		}
	}

	// Fall back to IP-based session lookup
	clientIP := c.ClientIP()
	sessionKey := fmt.Sprintf("ip_%s", clientIP)
	if session, exists := sessionStore[sessionKey]; exists && session.BirdAudioURL != "" {
		log.Printf("[STREAMING] Using audio URL from IP session for %s: %s", birdName, session.BirdAudioURL)
		h.proxyAudioContent(c, session.BirdAudioURL)
		return
	}

	// Then check the cached bird audio URL from daily update
	localDate := time.Now().Format("2006-01-02")
	cachedBirdName, audioURL, hasAudio := h.updateCache.GetDailyGlobalBirdWithAudio(localDate)

	if hasAudio && cachedBirdName == birdName && audioURL != "" {
		// Use cached audio URL - proxy the content instead of redirecting
		log.Printf("[STREAMING] Proxying cached audio for %s from %s", birdName, audioURL)
		h.proxyAudioContent(c, audioURL)
		return
	}

	// If no session exists, try to fetch the bird's audio directly
	log.Printf("[STREAMING] No session found, fetching audio for %s", birdName)
	bird, err := h.birdSelector.GetBirdByName(birdName)
	if err == nil && bird != nil && bird.AudioURL != "" {
		log.Printf("[STREAMING] Found audio for %s: %s", birdName, bird.AudioURL)
		// Store in session for future requests
		if sessionID != "" {
			sessionStore[sessionID] = &StreamingSession{
				SessionID:      sessionID,
				BirdName:       birdName,
				ScientificName: bird.ScientificName,
				BirdAudioURL:   bird.AudioURL,
				CreatedAt:      time.Now(),
			}
		}
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

// StreamDescription handles streaming requests for track 4 (location-aware description)
func (h *Handler) StreamDescription(c *gin.Context) {
	// Get bird name from query parameter
	birdName := c.Query("bird")
	if birdName == "" {
		log.Printf("[STREAMING] No bird name provided for description")
		c.Status(http.StatusBadRequest)
		return
	}

	// Try to detect location from client IP for location-aware facts
	clientIP := c.ClientIP()
	factLat := float64(0)
	factLng := float64(0)
	locationInfo := "generic"

	log.Printf("[STREAMING] Using generic facts (Yoto server IP: %s)", clientIP)

	log.Printf("[STREAMING] Description requested - Bird: %s, Location: %s, Coords: (%f, %f)",
		birdName, locationInfo, factLat, factLng)

	_, voiceID := h.audioManager.GetRandomIntroURL("")

	bird := &models.Bird{
		CommonName: birdName,
	}

	// Generate location-aware description
	audioData, err := h.audioManager.GenerateLocationAwareDescription(
		bird,
		voiceID,
		factLat,
		factLng,
	)

	if err != nil {
		log.Printf("[STREAMING] Failed to generate description: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Set appropriate headers for streaming
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Length", fmt.Sprintf("%d", len(audioData)))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	// Stream the audio
	c.Data(http.StatusOK, "audio/mpeg", audioData)
}

// StreamOutro handles streaming requests for track 5 (outro)
func (h *Handler) StreamOutro(c *gin.Context) {
	birdName := c.Query("bird")
	if birdName == "" {
		log.Printf("[STREAMING] No bird name provided for outro")
		c.Status(http.StatusBadRequest)
		return
	}

	_, voiceID := h.audioManager.GetRandomIntroURL("")

	log.Printf("[STREAMING] Outro requested - Bird: %s, Voice: %s", birdName, voiceID)

	audioData, err := h.audioManager.GenerateOutro(birdName, voiceID)
	if err != nil {
		log.Printf("[STREAMING] Failed to generate outro: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Set appropriate headers for streaming
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Length", fmt.Sprintf("%d", len(audioData)))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	// Stream the audio
	c.Data(http.StatusOK, "audio/mpeg", audioData)
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
