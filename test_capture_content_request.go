package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// LoggingTransport wraps http.RoundTripper to log requests
type LoggingTransport struct {
	Transport http.RoundTripper
	LogFile   *os.File
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only log requests to Yoto API
	if req.URL.Host == "api.yotoplay.com" && req.URL.Path == "/content" {
		// Read and store the request body
		var bodyBytes []byte
		if req.Body != nil {
			bodyBytes, _ = io.ReadAll(req.Body)
			// Restore the body for the actual request
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Log the request details
		logEntry := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"method":    req.Method,
			"url":       req.URL.String(),
			"headers":   req.Header,
		}

		// Parse and include the JSON body if present
		if len(bodyBytes) > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
				logEntry["body"] = jsonBody
			} else {
				logEntry["raw_body"] = string(bodyBytes)
			}
		}

		// Write to log file
		encoder := json.NewEncoder(t.LogFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(logEntry); err != nil {
			log.Printf("Failed to log request: %v", err)
		}
		
		// Also write a separator for readability
		fmt.Fprintln(t.LogFile, "\n===========================================\n")
	}

	// Proceed with the actual request
	if t.Transport == nil {
		t.Transport = http.DefaultTransport
	}
	return t.Transport.RoundTrip(req)
}

func main() {
	// Create log file
	timestamp := time.Now().Format("20060102_150405")
	logFileName := fmt.Sprintf("yoto_content_request_%s.txt", timestamp)
	logFile, err := os.Create(logFileName)
	if err != nil {
		log.Fatal("Failed to create log file:", err)
	}
	defer logFile.Close()

	fmt.Printf("Logging Yoto API requests to: %s\n", logFileName)
	fmt.Fprintln(logFile, "Yoto Content API Request Log")
	fmt.Fprintf(logFile, "Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintln(logFile, "===========================================\n")

	// Create HTTP client with logging transport
	client := &http.Client{
		Transport: &LoggingTransport{
			LogFile: logFile,
		},
	}

	// Make request to trigger daily update
	schedulerToken := os.Getenv("SCHEDULER_TOKEN")
	if schedulerToken == "" {
		fmt.Println("Warning: SCHEDULER_TOKEN not set, request may be unauthorized")
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/daily-update", nil)
	if err != nil {
		log.Fatal("Failed to create request:", err)
	}

	if schedulerToken != "" {
		req.Header.Set("X-Scheduler-Token", schedulerToken)
	}

	fmt.Println("Triggering daily update...")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed to make request:", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response status: %d\n", resp.StatusCode)
	fmt.Printf("Response body: %s\n", string(body))

	fmt.Printf("\nYoto API requests have been logged to: %s\n", logFileName)
	fmt.Println("Use 'cat", logFileName, "' to view the captured requests")
}