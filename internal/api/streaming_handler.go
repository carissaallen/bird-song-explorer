package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/gin-gonic/gin"
)

// StreamingSession stores location data for a play session
type StreamingSession struct {
	SessionID    string
	Location     *models.Location
	BirdName     string
	BirdAudioURL string
	VoiceID      string
	CreatedAt    time.Time
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
	
	// Bird will be set by the webhook handler when it creates the session
	// We don't fetch a daily bird here anymore since each play gets a fresh bird
	
	// Store session
	sessionStore[newSession.SessionID] = newSession
	
	// Periodic cleanup
	go cleanupSessions()
	
	return newSession
}

// StreamIntro handles streaming requests for track 1 (intro)
func (h *Handler) StreamIntro(c *gin.Context) {
	session := h.getOrCreateSession(c)
	
	// Get base URL
	baseURL := fmt.Sprintf("https://%s", c.Request.Host)
	if h.config.Environment == "development" {
		baseURL = fmt.Sprintf("http://%s", c.Request.Host)
	}
	
	// Get intro URL and voice ID
	introURL, voiceID := h.audioManager.GetRandomIntroURL(baseURL)
	session.VoiceID = voiceID // Store for other tracks
	
	// Log the streaming request
	locationInfo := "no location"
	if session.Location != nil {
		locationInfo = fmt.Sprintf("%s, %s", session.Location.City, session.Location.Country)
	}
	log.Printf("[STREAMING] Intro requested - Session: %s, Location: %s", 
		session.SessionID, locationInfo)
	
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
	// Get session by IP
	clientIP := c.ClientIP()
	sessionKey := fmt.Sprintf("ip_%s", clientIP)
	session, exists := sessionStore[sessionKey]
	
	if !exists || session.BirdName == "" {
		log.Printf("[STREAMING] No session or bird for announcement")
		// Return silence or default audio
		c.Status(http.StatusNotFound)
		return
	}
	
	log.Printf("[STREAMING] Bird announcement requested - Session: %s, Bird: %s", 
		sessionKey, session.BirdName)
	
	// Generate bird announcement audio
	audioData, err := h.audioManager.GenerateBirdAnnouncement(session.BirdName, session.VoiceID)
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
	// Get session by IP
	clientIP := c.ClientIP()
	sessionKey := fmt.Sprintf("ip_%s", clientIP)
	session, exists := sessionStore[sessionKey]
	
	if !exists || session.BirdAudioURL == "" {
		log.Printf("[STREAMING] No session or bird audio URL")
		c.Status(http.StatusNotFound)
		return
	}
	
	log.Printf("[STREAMING] Bird song requested - Session: %s, Bird: %s", 
		sessionKey, session.BirdName)
	
	// Redirect to the bird song URL
	c.Redirect(http.StatusFound, session.BirdAudioURL)
}

// StreamDescription handles streaming requests for track 4 (location-aware description)
func (h *Handler) StreamDescription(c *gin.Context) {
	// Get session by IP
	clientIP := c.ClientIP()
	sessionKey := fmt.Sprintf("ip_%s", clientIP)
	session, exists := sessionStore[sessionKey]
	
	if !exists || session.BirdName == "" {
		log.Printf("[STREAMING] No session for description")
		c.Status(http.StatusNotFound)
		return
	}
	
	// Determine location coordinates for fact generation
	factLat := float64(0)
	factLng := float64(0)
	locationInfo := "generic"
	
	if session.Location != nil {
		factLat = session.Location.Latitude
		factLng = session.Location.Longitude
		locationInfo = fmt.Sprintf("%s, %s", session.Location.City, session.Location.Country)
	}
	
	log.Printf("[STREAMING] Description requested - Session: %s, Location: %s, Coords: (%f, %f)", 
		sessionKey, locationInfo, factLat, factLng)
	
	// Get bird details
	bird := &models.Bird{
		CommonName: session.BirdName,
	}
	
	// Generate location-aware description
	audioData, err := h.audioManager.GenerateLocationAwareDescription(
		bird, 
		session.VoiceID,
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
	// Get session by IP
	clientIP := c.ClientIP()
	sessionKey := fmt.Sprintf("ip_%s", clientIP)
	session, exists := sessionStore[sessionKey]
	
	if !exists {
		log.Printf("[STREAMING] No session for outro")
		c.Status(http.StatusNotFound)
		return
	}
	
	log.Printf("[STREAMING] Outro requested - Session: %s", sessionKey)
	
	// Generate outro audio
	audioData, err := h.audioManager.GenerateOutro(session.BirdName, session.VoiceID)
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