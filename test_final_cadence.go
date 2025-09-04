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
	fmt.Println("🎵 Final Cadence Test - With Proper Pauses")
	fmt.Println("==========================================")

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

	fmt.Println("\n━━━ TRACK 2: Bird Announcement ━━━")
	
	// New improved announcement with spaced periods
	announcementText := fmt.Sprintf("Today's bird is the %s! . . . Listen carefully to its unique song.", birdName)
	fmt.Printf("📝 Text: \"%s\"\n", announcementText)
	
	fmt.Println("⏳ Generating announcement with proper pauses...")
	announcementAudio, err := generateTTS(announcementText, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate announcement: %v", err)
	} else {
		file1 := "test_final_announcement.mp3"
		os.WriteFile(file1, announcementAudio, 0644)
		fmt.Printf("💾 Saved: %s\n", file1)
	}

	fmt.Println("\n━━━ TRACK 4: Bird Description ━━━")
	
	// New improved description with spaced periods and em dash
	description := "The American Robin is a migratory songbird that's common across North America"
	descriptionText := fmt.Sprintf("Did you know? . . . %s . . . Isn't that amazing? — Nature is full of wonderful surprises!", description)
	fmt.Printf("📝 Text: \"%s\"\n", descriptionText)
	
	fmt.Println("⏳ Generating description with proper pauses...")
	descriptionAudio, err := generateTTS(descriptionText, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate description: %v", err)
	} else {
		file2 := "test_final_description.mp3"
		os.WriteFile(file2, descriptionAudio, 0644)
		fmt.Printf("💾 Saved: %s\n", file2)
	}

	// Play the results
	fmt.Println("\n🔊 Playing final versions...")
	
	if _, err := exec.LookPath("afplay"); err == nil {
		fmt.Println("\n   Playing Track 2 (Announcement)...")
		exec.Command("afplay", "test_final_announcement.mp3").Run()
		
		time.Sleep(1 * time.Second)
		
		fmt.Println("   Playing Track 4 (Description)...")
		exec.Command("afplay", "test_final_description.mp3").Run()
	}

	// Summary
	fmt.Println("\n✨ Final Implementation Summary")
	fmt.Println("================================")
	fmt.Println("Improvements Applied:")
	fmt.Println("  ✅ Speed: 0.90 (slower for children)")
	fmt.Println("  ✅ Pauses: Spaced periods (. . .) for major pauses")
	fmt.Println("  ✅ Pauses: Em dash (—) for minor pauses")
	fmt.Println("")
	fmt.Println("Track 2 (Announcement):")
	fmt.Println("  • Clear pause after bird name")
	fmt.Println("  • Using: \"[bird]! . . . Listen carefully...\"")
	fmt.Println("")
	fmt.Println("Track 4 (Description):")
	fmt.Println("  • Pauses between questions and facts")
	fmt.Println("  • Using: \"Did you know? . . . [fact] . . . Isn't that amazing? — Nature...\"")
	fmt.Println("")
	fmt.Println("Files generated:")
	fmt.Println("  • test_final_announcement.mp3")
	fmt.Println("  • test_final_description.mp3")
	fmt.Println("")
	fmt.Println("🎧 These should now have:")
	fmt.Println("  • Natural, comfortable pacing")
	fmt.Println("  • Clear separation between thoughts")
	fmt.Println("  • Time for children to process information")
}

// generateTTS generates TTS with final settings
func generateTTS(text string, voiceID string, apiKey string) ([]byte, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	requestBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.5,
			"speed":            0.90, // Slower speed for kids
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