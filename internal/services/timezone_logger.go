package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// TimezoneLogger logs timezone usage for monitoring and debugging
type TimezoneLogger struct {
	logDir string
}

// NewTimezoneLogger creates a new timezone logger
func NewTimezoneLogger() *TimezoneLogger {
	logDir := "logs/timezone"
	os.MkdirAll(logDir, 0755)
	return &TimezoneLogger{
		logDir: logDir,
	}
}

// TimezoneLogEntry represents a single timezone log entry
type TimezoneLogEntry struct {
	Timestamp       time.Time              `json:"timestamp"`
	DeviceID        string                 `json:"device_id"`
	CardID          string                 `json:"card_id"`
	Timezone        string                 `json:"timezone"`
	LocalTime       string                 `json:"local_time"`
	LocalHour       int                    `json:"local_hour"`
	NatureSound     string                 `json:"nature_sound"`
	TimePeriod      string                 `json:"time_period"`
	Location        string                 `json:"location"`
	ServerTime      time.Time              `json:"server_time"`
	TimeDifference  float64                `json:"time_difference_hours"`
}

// LogTimezoneUsage logs timezone information when a card is played
func (tl *TimezoneLogger) LogTimezoneUsage(deviceID, cardID, timezone, location string) {
	timeHelper := NewUserTimeHelper()
	userTime := timeHelper.GetUserLocalTime(timezone)
	serverTime := time.Now()
	
	// Calculate time difference
	timeDiff := userTime.Sub(serverTime).Hours()
	
	entry := TimezoneLogEntry{
		Timestamp:      serverTime,
		DeviceID:       deviceID,
		CardID:         cardID,
		Timezone:       timezone,
		LocalTime:      userTime.Format("15:04:05"),
		LocalHour:      userTime.Hour(),
		NatureSound:    timeHelper.GetNatureSoundForUserTime(timezone),
		TimePeriod:     getTimePeriod(userTime.Hour()),
		Location:       location,
		ServerTime:     serverTime,
		TimeDifference: timeDiff,
	}

	// Log to console
	log.Printf("[TIMEZONE_LOG] Device: %s | Timezone: %s | Local: %s (Hour: %d) | Nature: %s | Location: %s | Diff: %.1fh",
		deviceID, timezone, entry.LocalTime, entry.LocalHour, entry.NatureSound, location, timeDiff)

	// Save to daily log file
	tl.saveToFile(entry)
	
	// Generate usage statistics
	tl.generateStats(entry)
}

// saveToFile saves the log entry to a daily JSON file
func (tl *TimezoneLogger) saveToFile(entry TimezoneLogEntry) {
	dateStr := time.Now().Format("2006-01-02")
	filename := filepath.Join(tl.logDir, fmt.Sprintf("timezone_%s.jsonl", dateStr))
	
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[TIMEZONE_LOG] Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("[TIMEZONE_LOG] Failed to marshal log entry: %v", err)
		return
	}

	file.Write(data)
	file.Write([]byte("\n"))
}

// generateStats generates and logs usage statistics
func (tl *TimezoneLogger) generateStats(entry TimezoneLogEntry) {
	// Read today's logs
	dateStr := time.Now().Format("2006-01-02")
	filename := filepath.Join(tl.logDir, fmt.Sprintf("timezone_%s.jsonl", dateStr))
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	// Count timezone usage
	timezoneCount := make(map[string]int)
	natureSoundCount := make(map[string]int)
	hourDistribution := make(map[int]int)
	
	lines := string(data)
	for _, line := range splitLines(lines) {
		if line == "" {
			continue
		}
		
		var logEntry TimezoneLogEntry
		if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
			timezoneCount[logEntry.Timezone]++
			natureSoundCount[logEntry.NatureSound]++
			hourDistribution[logEntry.LocalHour]++
		}
	}

	// Log statistics
	log.Printf("[TIMEZONE_STATS] Today's Usage:")
	log.Printf("  Top Timezones:")
	for tz, count := range timezoneCount {
		if count > 0 {
			log.Printf("    %s: %d plays", tz, count)
		}
	}
	log.Printf("  Nature Sounds Used:")
	for sound, count := range natureSoundCount {
		if count > 0 {
			log.Printf("    %s: %d times", sound, count)
		}
	}
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// GetTimezoneStats returns statistics for a given date
func (tl *TimezoneLogger) GetTimezoneStats(date string) map[string]interface{} {
	filename := filepath.Join(tl.logDir, fmt.Sprintf("timezone_%s.jsonl", date))
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return map[string]interface{}{
			"error": "No data for date",
			"date":  date,
		}
	}

	timezoneCount := make(map[string]int)
	natureSoundCount := make(map[string]int)
	hourDistribution := make(map[int]int)
	totalPlays := 0
	
	lines := string(data)
	for _, line := range splitLines(lines) {
		if line == "" {
			continue
		}
		
		var entry TimezoneLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			timezoneCount[entry.Timezone]++
			natureSoundCount[entry.NatureSound]++
			hourDistribution[entry.LocalHour]++
			totalPlays++
		}
	}

	return map[string]interface{}{
		"date":             date,
		"total_plays":      totalPlays,
		"timezones":        timezoneCount,
		"nature_sounds":    natureSoundCount,
		"hour_distribution": hourDistribution,
	}
}