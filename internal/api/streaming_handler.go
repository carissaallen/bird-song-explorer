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
		// Get the daily bird from cache (set by scheduler at 12:00 UTC)
		localDate := time.Now().UTC().Format("2006-01-02")
		cachedBirdName, exists := h.updateCache.GetDailyGlobalBird(localDate)
		if exists && cachedBirdName != "" {
			session.BirdName = cachedBirdName
			log.Printf("[STREAMING] intro: Using cached daily bird: %s", cachedBirdName)
		} else {
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
		// First try to get the daily bird from cache (set by scheduler at 12:00 UTC)
		localDate := time.Now().UTC().Format("2006-01-02")
		cachedBirdName, exists := h.updateCache.GetDailyGlobalBird(localDate)
		if exists && cachedBirdName != "" {
			birdName = cachedBirdName
			session.BirdName = birdName
			sessionStore[session.SessionID] = session
			log.Printf("[STREAMING] announcement: Using cached daily bird: %s", birdName)
		} else {
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
		// First try to get the daily bird from cache (set by scheduler at 12:00 UTC)
		localDate := time.Now().UTC().Format("2006-01-02")
		cachedBirdName, exists := h.updateCache.GetDailyGlobalBird(localDate)
		if exists && cachedBirdName != "" {
			birdName = cachedBirdName
			session.BirdName = birdName
			sessionStore[session.SessionID] = session
			log.Printf("[STREAMING] description: Using cached daily bird: %s", birdName)
		} else {
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
	}

	birdDir := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	gcsURL := fmt.Sprintf("https://storage.googleapis.com/bird-song-explorer-audio/birds/%s/narration/description.mp3", birdDir)

	c.Redirect(http.StatusFound, gcsURL)
}

func (h *Handler) StreamOutro(c *gin.Context) {
	sessionID := c.Query("session")
	session := h.getOrCreateSession(c, sessionID)

	birdName := session.BirdName
	if birdName == "" {
		// First try to get the daily bird from cache (set by scheduler at 12:00 UTC)
		localDate := time.Now().UTC().Format("2006-01-02")
		cachedBirdName, exists := h.updateCache.GetDailyGlobalBird(localDate)
		if exists && cachedBirdName != "" {
			birdName = cachedBirdName
			session.BirdName = birdName
			sessionStore[session.SessionID] = session
			log.Printf("[STREAMING] outro: Using cached daily bird: %s", birdName)
		} else {
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
	}

	birdDir := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	gcsURL := fmt.Sprintf("https://storage.googleapis.com/bird-song-explorer-audio/birds/%s/narration/outro.mp3", birdDir)

	c.Redirect(http.StatusFound, gcsURL)
}
