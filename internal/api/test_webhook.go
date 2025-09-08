package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TestWebhookHandler simulates a webhook call for testing
func (h *Handler) TestWebhookHandler(c *gin.Context) {
	// Get optional parameters
	cardID := c.Query("cardId")
	if cardID == "" {
		cardID = h.config.YotoCardID
	}
	
	deviceID := c.Query("deviceId")
	eventType := c.Query("eventType")
	if eventType == "" {
		eventType = "card.played"
	}

	// Create webhook payload
	webhook := map[string]interface{}{
		"eventType": eventType,
		"cardId":    cardID,
		"deviceId":  deviceID,
		"userId":    "test-user",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Marshal to JSON
	webhookJSON, err := json.Marshal(webhook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook payload"})
		return
	}

	// Create a new request to the webhook endpoint
	req, err := http.NewRequest("POST", "/api/v1/yoto/webhook", bytes.NewBuffer(webhookJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers from original request
	req.Header = c.Request.Header.Clone()
	req.Header.Set("Content-Type", "application/json")
	
	// Create a response writer to capture the webhook response
	rw := &responseWriter{
		body:       &bytes.Buffer{},
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}

	// Process the webhook using V4 handler with regional checking
	// Create a proper context with the original request's headers for IP detection
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Real-IP", c.ClientIP())
	
	// Create a new gin context for the webhook handler
	webhookContext := &gin.Context{
		Request: req,
		Writer:  rw,
		Params:  c.Params,
		Keys:    make(map[string]interface{}),
	}
	
	h.HandleYotoWebhookV4(webhookContext)

	// Parse the response
	var webhookResponse map[string]interface{}
	if err := json.Unmarshal(rw.body.Bytes(), &webhookResponse); err != nil {
		webhookResponse = map[string]interface{}{
			"raw_response": rw.body.String(),
		}
	}

	// Return test results
	c.JSON(http.StatusOK, gin.H{
		"test_info": gin.H{
			"description": "Simulated webhook call",
			"webhook_payload": webhook,
			"client_ip": c.ClientIP(),
		},
		"webhook_response": webhookResponse,
		"status_code": rw.statusCode,
	})
}

// responseWriter implements http.ResponseWriter for capturing responses
type responseWriter struct {
	body       *bytes.Buffer
	header     http.Header
	statusCode int
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// Implement gin.ResponseWriter interface
func (rw *responseWriter) WriteString(s string) (int, error) {
	return rw.body.WriteString(s)
}

func (rw *responseWriter) Written() bool {
	return rw.body.Len() > 0
}

func (rw *responseWriter) WriteHeaderNow() {}

func (rw *responseWriter) Status() int {
	return rw.statusCode
}

func (rw *responseWriter) Size() int {
	return rw.body.Len()
}

func (rw *responseWriter) Pusher() http.Pusher {
	return nil
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	return nil
}

func (rw *responseWriter) Flush() {}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("hijack not supported")
}