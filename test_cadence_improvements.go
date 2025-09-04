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
	fmt.Println("🎵 Testing Improved Speech Cadence with Pauses")
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

	// TRACK 2: Test announcement with pauses
	fmt.Println("\n━━━ TRACK 2: Bird Announcement (with pauses) ━━━")
	
	// Old version (for comparison)
	oldAnnouncementText := fmt.Sprintf("Today's bird is the %s! Listen carefully to its unique song.", birdName)
	fmt.Printf("❌ OLD (run-on): \"%s\"\n", oldAnnouncementText)
	
	// New version with pauses
	newAnnouncementText := fmt.Sprintf("Today's bird is the %s!\n\n...Listen carefully to its unique song.", birdName)
	fmt.Printf("✅ NEW (paused): \"%s\"\n", newAnnouncementText)
	
	// Generate both versions for comparison
	fmt.Println("\n⏳ Generating announcement without pauses...")
	oldAnnouncement, err := generateTTS(oldAnnouncementText, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate old announcement: %v", err)
	} else {
		oldFile := "test_announcement_old_runon.mp3"
		os.WriteFile(oldFile, oldAnnouncement, 0644)
		fmt.Printf("💾 Saved old version: %s\n", oldFile)
	}
	
	fmt.Println("⏳ Generating announcement with pauses...")
	newAnnouncement, err := generateTTS(newAnnouncementText, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate new announcement: %v", err)
	} else {
		newFile := "test_announcement_new_paused.mp3"
		os.WriteFile(newFile, newAnnouncement, 0644)
		fmt.Printf("💾 Saved new version: %s\n", newFile)
	}

	// TRACK 4: Test description with pauses
	fmt.Println("\n━━━ TRACK 4: Bird Description (with pauses) ━━━")
	
	description := "The American Robin is a migratory songbird that's common across North America"
	
	// Old version
	oldDescriptionText := fmt.Sprintf("Did you know? %s Isn't that amazing? Nature is full of wonderful surprises!", description)
	fmt.Printf("❌ OLD (run-on): \"%s\"\n", oldDescriptionText)
	
	// New version with pauses
	newDescriptionText := fmt.Sprintf("Did you know?\n\n%s\n\n...Isn't that amazing?\n\nNature is full of wonderful surprises!", description)
	fmt.Printf("✅ NEW (paused): \"%s\"\n", newDescriptionText)
	
	fmt.Println("\n⏳ Generating description without pauses...")
	oldDescription, err := generateTTS(oldDescriptionText, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate old description: %v", err)
	} else {
		oldFile := "test_description_old_runon.mp3"
		os.WriteFile(oldFile, oldDescription, 0644)
		fmt.Printf("💾 Saved old version: %s\n", oldFile)
	}
	
	fmt.Println("⏳ Generating description with pauses...")
	newDescription, err := generateTTS(newDescriptionText, dailyVoice.ID, elevenLabsKey)
	if err != nil {
		log.Printf("Failed to generate new description: %v", err)
	} else {
		newFile := "test_description_new_paused.mp3"
		os.WriteFile(newFile, newDescription, 0644)
		fmt.Printf("💾 Saved new version: %s\n", newFile)
	}

	// Play comparison
	fmt.Println("\n🔊 Playing comparisons...")
	fmt.Println("\n1️⃣  ANNOUNCEMENT COMPARISON:")
	
	if _, err := exec.LookPath("afplay"); err == nil {
		fmt.Println("   Playing OLD (run-on) version...")
		exec.Command("afplay", "test_announcement_old_runon.mp3").Run()
		
		time.Sleep(1 * time.Second)
		
		fmt.Println("   Playing NEW (with pauses) version...")
		exec.Command("afplay", "test_announcement_new_paused.mp3").Run()
	}
	
	fmt.Println("\n2️⃣  DESCRIPTION COMPARISON:")
	if _, err := exec.LookPath("afplay"); err == nil {
		fmt.Println("   Playing OLD (run-on) version...")
		exec.Command("afplay", "test_description_old_runon.mp3").Run()
		
		time.Sleep(1 * time.Second)
		
		fmt.Println("   Playing NEW (with pauses) version...")
		exec.Command("afplay", "test_description_new_paused.mp3").Run()
	}

	// Summary
	fmt.Println("\n✨ Pause Technique Summary")
	fmt.Println("==========================")
	fmt.Println("ElevenLabs pause techniques used:")
	fmt.Println("  • Line breaks (\\n\\n) = Natural pause between sentences")
	fmt.Println("  • Ellipsis (...) = Brief dramatic pause")
	fmt.Println("  • Combination creates more natural speech rhythm")
	fmt.Println("")
	fmt.Println("Files generated for comparison:")
	fmt.Println("  • test_announcement_old_runon.mp3 (no pauses)")
	fmt.Println("  • test_announcement_new_paused.mp3 (with pauses)")
	fmt.Println("  • test_description_old_runon.mp3 (no pauses)")
	fmt.Println("  • test_description_new_paused.mp3 (with pauses)")
	fmt.Println("")
	fmt.Println("🎧 Listen for:")
	fmt.Println("  • More natural breathing spaces")
	fmt.Println("  • Better emphasis on key information")
	fmt.Println("  • Easier for children to process")
	fmt.Println("")
	fmt.Println("📝 For future intro/outro re-recording:")
	fmt.Println("  • Apply same pause techniques")
	fmt.Println("  • Test with actual voice before bulk generation")
	fmt.Println("  • Consider saving pause-formatted scripts")
}

// generateTTS is a simple TTS generation function for testing
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