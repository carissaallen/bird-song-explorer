package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
)

func main() {
	fmt.Println("🎵 Testing TTS Speed Comparison (0.95 vs 0.90)")
	fmt.Println("==============================================")

	// Check for required environment variable
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		log.Fatal("Error: ELEVENLABS_API_KEY environment variable is not set")
	}

	// Get the daily voice
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()
	fmt.Printf("\n🎤 Using voice: %s (ID: %s)\n", dailyVoice.Name, dailyVoice.ID)

	// Test bird name
	birdName := "American Robin"
	fmt.Printf("🐦 Test bird: %s\n", birdName)

	// Test text with pauses
	announcementText := fmt.Sprintf("Today's bird is the %s!\n\n...Listen carefully to its unique song.", birdName)
	
	fmt.Println("\n━━━ Speed Comparison Test ━━━")
	fmt.Printf("📝 Text: \"%s\"\n", announcementText)

	// Generate at 0.95 speed (old)
	fmt.Println("\n⏳ Generating at 0.95 speed (old)...")
	audio095, err := generateTTSWithSpeed(announcementText, dailyVoice.ID, elevenLabsKey, 0.95)
	if err != nil {
		log.Printf("Failed to generate at 0.95: %v", err)
	} else {
		file095 := "test_speed_095.mp3"
		os.WriteFile(file095, audio095, 0644)
		fmt.Printf("💾 Saved 0.95 speed: %s (%d bytes)\n", file095, len(audio095))
	}

	// Generate at 0.90 speed (new)
	fmt.Println("\n⏳ Generating at 0.90 speed (new)...")
	audio090, err := generateTTSWithSpeed(announcementText, dailyVoice.ID, elevenLabsKey, 0.90)
	if err != nil {
		log.Printf("Failed to generate at 0.90: %v", err)
	} else {
		file090 := "test_speed_090.mp3"
		os.WriteFile(file090, audio090, 0644)
		fmt.Printf("💾 Saved 0.90 speed: %s (%d bytes)\n", file090, len(audio090))
	}

	// Generate at 0.85 speed (even slower for comparison)
	fmt.Println("\n⏳ Generating at 0.85 speed (even slower)...")
	audio085, err := generateTTSWithSpeed(announcementText, dailyVoice.ID, elevenLabsKey, 0.85)
	if err != nil {
		log.Printf("Failed to generate at 0.85: %v", err)
	} else {
		file085 := "test_speed_085.mp3"
		os.WriteFile(file085, audio085, 0644)
		fmt.Printf("💾 Saved 0.85 speed: %s (%d bytes)\n", file085, len(audio085))
	}

	// Play comparison
	fmt.Println("\n🔊 Playing speed comparison...")
	
	if _, err := exec.LookPath("afplay"); err == nil {
		fmt.Println("\n1️⃣  Playing at 0.95 speed (old - slightly fast)...")
		exec.Command("afplay", "test_speed_095.mp3").Run()
		
		time.Sleep(1 * time.Second)
		
		fmt.Println("2️⃣  Playing at 0.90 speed (new - better pace)...")
		exec.Command("afplay", "test_speed_090.mp3").Run()
		
		time.Sleep(1 * time.Second)
		
		fmt.Println("3️⃣  Playing at 0.85 speed (even slower)...")
		exec.Command("afplay", "test_speed_085.mp3").Run()
	}

	// Summary
	fmt.Println("\n✨ Speed Comparison Summary")
	fmt.Println("===========================")
	fmt.Println("Speed Settings:")
	fmt.Println("  • 0.95 = Original (slightly too fast)")
	fmt.Println("  • 0.90 = New default (better for kids)")
	fmt.Println("  • 0.85 = Even slower (if needed)")
	fmt.Println("")
	fmt.Println("Combined with pauses:")
	fmt.Println("  • Line breaks (\\n\\n) for sentence pauses")
	fmt.Println("  • Ellipsis (...) for dramatic effect")
	fmt.Println("  • Speed 0.90 for comfortable listening")
	fmt.Println("")
	fmt.Println("Files generated:")
	fmt.Println("  • test_speed_095.mp3")
	fmt.Println("  • test_speed_090.mp3")
	fmt.Println("  • test_speed_085.mp3")
	fmt.Println("")
	fmt.Println("🎧 The 0.90 speed should:")
	fmt.Println("  • Feel more relaxed and comfortable")
	fmt.Println("  • Give children more time to process")
	fmt.Println("  • Work well with the added pauses")
}

// generateTTSWithSpeed generates TTS at a specific speed
func generateTTSWithSpeed(text string, voiceID string, apiKey string, speed float64) ([]byte, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.5,
			"speed":            speed,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ElevenLabs API error: %d - %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}