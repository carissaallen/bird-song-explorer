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
	fmt.Println("🎵 Testing Different Pause Methods in ElevenLabs TTS")
	fmt.Println("====================================================")

	// Check for required environment variable
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		log.Fatal("Error: ELEVENLABS_API_KEY environment variable is not set")
	}

	// Get the daily voice
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()
	fmt.Printf("\n🎤 Using voice: %s (ID: %s)\n", dailyVoice.Name, dailyVoice.ID)

	birdName := "American Robin"
	fmt.Printf("🐦 Test bird: %s\n", birdName)

	fmt.Println("\n━━━ Testing Different Pause Methods ━━━")
	
	// Test 1: No pause at all
	text1 := fmt.Sprintf("Today's bird is the %s! Listen carefully to its unique song.", birdName)
	fmt.Printf("\n1️⃣  NO PAUSE:\n\"%s\"\n", text1)
	
	// Test 2: Line breaks (current method)
	text2 := fmt.Sprintf("Today's bird is the %s!\n\n\nListen carefully to its unique song.", birdName)
	fmt.Printf("\n2️⃣  LINE BREAKS (\\n\\n\\n):\n\"%s\"\n", text2)
	
	// Test 3: Multiple periods
	text3 := fmt.Sprintf("Today's bird is the %s!... Listen carefully to its unique song.", birdName)
	fmt.Printf("\n3️⃣  ELLIPSIS (...):\n\"%s\"\n", text3)
	
	// Test 4: Multiple periods with spaces
	text4 := fmt.Sprintf("Today's bird is the %s! . . . Listen carefully to its unique song.", birdName)
	fmt.Printf("\n4️⃣  SPACED PERIODS (. . .):\n\"%s\"\n", text4)
	
	// Test 5: Comma pause
	text5 := fmt.Sprintf("Today's bird is the %s!, , , Listen carefully to its unique song.", birdName)
	fmt.Printf("\n5️⃣  COMMAS (, , ,):\n\"%s\"\n", text5)
	
	// Test 6: Em dash
	text6 := fmt.Sprintf("Today's bird is the %s! — Listen carefully to its unique song.", birdName)
	fmt.Printf("\n6️⃣  EM DASH (—):\n\"%s\"\n", text6)

	// Test 7: Combination
	text7 := fmt.Sprintf("Today's bird is the %s! ... ... Listen carefully to its unique song.", birdName)
	fmt.Printf("\n7️⃣  DOUBLE ELLIPSIS (... ...):\n\"%s\"\n", text7)

	// Generate all versions
	fmt.Println("\n⏳ Generating all pause method tests...")
	
	texts := []string{text1, text2, text3, text4, text5, text6, text7}
	names := []string{"no_pause", "line_breaks", "ellipsis", "spaced_periods", "commas", "em_dash", "double_ellipsis"}
	
	for i, text := range texts {
		audio, err := generateTTS(text, dailyVoice.ID, elevenLabsKey)
		if err != nil {
			log.Printf("Failed to generate %s: %v", names[i], err)
		} else {
			filename := fmt.Sprintf("test_pause_%s.mp3", names[i])
			os.WriteFile(filename, audio, 0644)
			fmt.Printf("💾 Saved: %s\n", filename)
		}
	}

	// Play comparison
	fmt.Println("\n🔊 Playing all versions for comparison...")
	
	if _, err := exec.LookPath("afplay"); err == nil {
		for i, name := range names {
			fmt.Printf("\n   %d. Playing %s...\n", i+1, name)
			exec.Command("afplay", fmt.Sprintf("test_pause_%s.mp3", name)).Run()
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Summary
	fmt.Println("\n✨ Pause Method Test Results")
	fmt.Println("=============================")
	fmt.Println("Files generated:")
	for _, name := range names {
		fmt.Printf("  • test_pause_%s.mp3\n", name)
	}
	fmt.Println("")
	fmt.Println("🎧 Listen for actual pause differences!")
	fmt.Println("")
	fmt.Println("💡 ElevenLabs pause support:")
	fmt.Println("  • Line breaks (\\n) might NOT create pauses")
	fmt.Println("  • Punctuation like ... or — might work better")
	fmt.Println("  • Multiple periods or commas might help")
	fmt.Println("  • We need to find what actually works!")
}

// generateTTS generates TTS with current settings
func generateTTS(text string, voiceID string, apiKey string) ([]byte, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.5,
			"speed":            0.90,
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