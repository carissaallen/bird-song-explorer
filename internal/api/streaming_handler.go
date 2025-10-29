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

// getDailyBirdWithFallback gets the bird from cache with timezone awareness
// If cache fails, falls back to GetCyclingBird() and updates the card
func (h *Handler) getDailyBirdWithFallback(c *gin.Context, context string) (string, error) {
	now := time.Now().UTC()
	var lookupDate string

	// Timezone-aware cache lookup
	// Before 12:00 UTC: use yesterday's bird (matches yesterday's card icon)
	// After 12:00 UTC: use today's bird (matches today's card icon)
	if now.Hour() < 12 {
		yesterday := now.AddDate(0, 0, -1)
		lookupDate = yesterday.Format("2006-01-02")
		log.Printf("[STREAMING] %s: Before daily update (hour=%d), looking up yesterday: %s", context, now.Hour(), lookupDate)
	} else {
		lookupDate = now.Format("2006-01-02")
		log.Printf("[STREAMING] %s: After daily update (hour=%d), looking up today: %s", context, now.Hour(), lookupDate)
	}

	// Try primary lookup
	cachedBirdName, exists := h.updateCache.GetDailyGlobalBird(lookupDate)
	if exists && cachedBirdName != "" {
		log.Printf("[STREAMING] %s: ✅ Using cached daily bird: %s (date: %s)", context, cachedBirdName, lookupDate)
		return cachedBirdName, nil
	}

	// Try yesterday as backup (in case cache failed)
	yesterday := now.AddDate(0, 0, -1)
	cachedBirdName, exists = h.updateCache.GetDailyGlobalBird(yesterday.Format("2006-01-02"))
	if exists && cachedBirdName != "" {
		log.Printf("[STREAMING] %s: ⚠️  Primary cache miss, using yesterday's bird: %s", context, cachedBirdName)
		return cachedBirdName, nil
	}

	// Fallback: Get cycling bird AND update card with new icon
	log.Printf("[STREAMING] %s: ❌ Cache failed, falling back to GetCyclingBird()", context)
	bird := h.availableBirds.GetCyclingBird()
	if bird == nil {
		return "", fmt.Errorf("no bird available")
	}

	// Update the card with the fallback bird's icon
	cardID := h.config.YotoCardID
	if cardID != "" {
		baseURL := fmt.Sprintf("https://%s", c.Request.Host)
		sessionID := fmt.Sprintf("%s_%d", cardID, now.Unix())

		log.Printf("[STREAMING] %s: 🔄 Updating card with fallback bird: %s", context, bird.CommonName)
		contentManager := h.yotoClient.NewContentManager()
		err := contentManager.UpdateCardWithStreamingTracks(cardID, bird.CommonName, baseURL, sessionID)
		if err != nil {
			log.Printf("[STREAMING] %s: ⚠️  Failed to update card: %v", context, err)
		} else {
			log.Printf("[STREAMING] %s: ✅ Card updated with fresh icon for: %s", context, bird.CommonName)
		}
	}

	return bird.CommonName, nil
}

func (h *Handler) StreamIntro(c *gin.Context) {
	birdName := c.Query("bird")
	sessionID := c.Query("session")
	session := h.getOrCreateSession(c, sessionID)

	if birdName != "" && session.BirdName != birdName {
		session.BirdName = birdName
	} else if session.BirdName == "" {
		// Get bird with timezone-aware caching and fallback
		selectedBird, err := h.getDailyBirdWithFallback(c, "intro")
		if err != nil {
			log.Printf("[STREAMING] intro: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Bird content not ready yet. Please try again in a few minutes.",
			})
			return
		}
		session.BirdName = selectedBird
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
		// Get bird with timezone-aware caching and fallback
		selectedBird, err := h.getDailyBirdWithFallback(c, "announcement")
		if err != nil {
			log.Printf("[STREAMING] announcement: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}
		birdName = selectedBird
		session.BirdName = birdName
		sessionStore[session.SessionID] = session
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
		// Get bird with timezone-aware caching and fallback
		selectedBird, err := h.getDailyBirdWithFallback(c, "description")
		if err != nil {
			log.Printf("[STREAMING] description: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}
		birdName = selectedBird
		session.BirdName = birdName
		sessionStore[session.SessionID] = session
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
		// Get bird with timezone-aware caching and fallback
		selectedBird, err := h.getDailyBirdWithFallback(c, "outro")
		if err != nil {
			log.Printf("[STREAMING] outro: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}
		birdName = selectedBird
		session.BirdName = birdName
		sessionStore[session.SessionID] = session
	}

	birdDir := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	gcsURL := fmt.Sprintf("https://storage.googleapis.com/bird-song-explorer-audio/birds/%s/narration/outro.mp3", birdDir)

	c.Redirect(http.StatusFound, gcsURL)
}
