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
	fmt.Println("🎵 Testing Pause Length Comparison")
	fmt.Println("===================================")

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

	fmt.Println("\n━━━ Pause Length Comparison ━━━")
	
	// Version 1: Original single pause (2 line breaks)
	text1 := fmt.Sprintf("Today's bird is the %s!\n\n...Listen carefully to its unique song.", birdName)
	fmt.Printf("\n1️⃣  ORIGINAL (2 line breaks):\n\"%s\"\n", text1)
	
	// Version 2: New extended pause (3 line breaks)
	text2 := fmt.Sprintf("Today's bird is the %s!\n\n\n...Listen carefully to its unique song.", birdName)
	fmt.Printf("\n2️⃣  EXTENDED (3 line breaks):\n\"%s\"\n", text2)
	
	// Version 3: Even longer pause (4 line breaks) for comparison
	text3 := fmt.Sprintf("Today's bird is the %s!\n\n\n\n...Listen carefully to its unique song.", birdName)
	fmt.Printf("\n3️⃣  EXTRA LONG (4 line breaks):\n\"%s\"\n", text3)

	// Generate all three versions
	fmt.Println("\n⏳ Generating announcements with different pause lengths...")
	
	// Version 1
	audio1, err := generateTTS(text1, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate version 1: %v", err)
	} else {
		file1 := "test_pause_2breaks.mp3"
		os.WriteFile(file1, audio1, 0644)
		fmt.Printf("💾 Saved 2 breaks version: %s\n", file1)
	}
	
	// Version 2
	audio2, err := generateTTS(text2, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate version 2: %v", err)
	} else {
		file2 := "test_pause_3breaks.mp3"
		os.WriteFile(file2, audio2, 0644)
		fmt.Printf("💾 Saved 3 breaks version: %s\n", file2)
	}
	
	// Version 3
	audio3, err := generateTTS(text3, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate version 3: %v", err)
	} else {
		file3 := "test_pause_4breaks.mp3"
		os.WriteFile(file3, audio3, 0644)
		fmt.Printf("💾 Saved 4 breaks version: %s\n", file3)
	}

	// Play comparison
	fmt.Println("\n🔊 Playing pause comparison...")
	
	if _, err := exec.LookPath("afplay"); err == nil {
		fmt.Println("\n   Playing 2 breaks (original)...")
		exec.Command("afplay", "test_pause_2breaks.mp3").Run()
		
		time.Sleep(500 * time.Millisecond)
		
		fmt.Println("   Playing 3 breaks (new)...")
		exec.Command("afplay", "test_pause_3breaks.mp3").Run()
		
		time.Sleep(500 * time.Millisecond)
		
		fmt.Println("   Playing 4 breaks (extra long)...")
		exec.Command("afplay", "test_pause_4breaks.mp3").Run()
	}

	// Summary
	fmt.Println("\n✨ Pause Comparison Summary")
	fmt.Println("============================")
	fmt.Println("Pause techniques tested:")
	fmt.Println("  • 2 line breaks (\\n\\n) = Original pause")
	fmt.Println("  • 3 line breaks (\\n\\n\\n) = Extended pause (NEW)")
	fmt.Println("  • 4 line breaks (\\n\\n\\n\\n) = Extra long pause")
	fmt.Println("")
	fmt.Println("Files generated:")
	fmt.Println("  • test_pause_2breaks.mp3 (original)")
	fmt.Println("  • test_pause_3breaks.mp3 (new recommendation)")
	fmt.Println("  • test_pause_4breaks.mp3 (if more pause needed)")
	fmt.Println("")
	fmt.Println("🎧 Listen for:")
	fmt.Println("  • Clear separation between bird name and instruction")
	fmt.Println("  • Time for children to process the bird name")
	fmt.Println("  • Natural conversation flow")
	fmt.Println("")
	fmt.Println("💡 Recommendation:")
	fmt.Println("  The 3-break pause gives good emphasis on the bird name")
	fmt.Println("  before moving to the listening instruction.")
}

// generateTTS generates TTS with current settings (0.90 speed)
func generateTTS(text string, voiceID string, apiKey string) ([]byte, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.5,
			"speed":            0.90, // New slower speed
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