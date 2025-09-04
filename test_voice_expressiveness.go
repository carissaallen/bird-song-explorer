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

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		log.Fatal("ELEVENLABS_API_KEY not set")
	}

	// Get today's voice
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()
	
	fmt.Printf("🎤 Testing voice expressiveness with: %s (%s)\n", dailyVoice.Name, dailyVoice.Region)
	fmt.Printf("📅 Date: %s\n", time.Now().Format("2006-01-02"))
	fmt.Println("=" + string(make([]byte, 60)))

	// Test text - same for all tests
	testText := "Hello little explorers! Today we have a very special bird for you. It's the amazing House Finch! Listen carefully to its beautiful song!"

	// Test configurations
	tests := []struct {
		name      string
		stability float64
		simBoost  float64
		speed     float64
		useBoost  bool
	}{
		{
			name:      "Original (monotone)",
			stability: 0.75,
			simBoost:  0.75,
			speed:     0.90,
			useBoost:  false,
		},
		{
			name:      "Low stability for emotion",
			stability: 0.35,
			simBoost:  0.85,
			speed:     0.92,
			useBoost:  true,
		},
		{
			name:      "Very low stability",
			stability: 0.2,
			simBoost:  0.9,
			speed:     0.92,
			useBoost:  true,
		},
		{
			name:      "Medium stability balanced",
			stability: 0.5,
			simBoost:  0.85,
			speed:     0.92,
			useBoost:  true,
		},
	}

	for i, test := range tests {
		fmt.Printf("\n🧪 Test %d: %s\n", i+1, test.name)
		fmt.Printf("   stability: %.2f, similarity: %.2f, speed: %.2f, speaker_boost: %v\n", 
			test.stability, test.simBoost, test.speed, test.useBoost)
		
		audio, err := generateTestAudio(testText, dailyVoice.ID, elevenLabsKey, 
			test.stability, test.simBoost, test.speed, test.useBoost)
		
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}
		
		filename := fmt.Sprintf("test_voice_%d_%s.mp3", i+1, test.name)
		err = os.WriteFile(filename, audio, 0644)
		if err != nil {
			fmt.Printf("   ❌ Failed to save: %v\n", err)
			continue
		}
		
		fmt.Printf("   ✅ Saved: %s (%d bytes)\n", filename, len(audio))
	}
	
	fmt.Println("\n🎧 Test files generated! Listen to compare:")
	fmt.Println("   1. test_voice_1_Original (monotone).mp3")
	fmt.Println("   2. test_voice_2_Low stability for emotion.mp3")
	fmt.Println("   3. test_voice_3_Very low stability.mp3")
	fmt.Println("   4. test_voice_4_Medium stability balanced.mp3")
	fmt.Println("\n💡 Lower stability = more emotional range")
	fmt.Println("💡 Higher similarity_boost = closer to original voice")
	fmt.Println("💡 use_speaker_boost = enhanced clarity")
}

func generateTestAudio(text, voiceID, apiKey string, stability, simBoost, speed float64, useBoost bool) ([]byte, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	voiceSettings := map[string]interface{}{
		"stability":        stability,
		"similarity_boost": simBoost,
		"speed":           speed,
	}
	
	if useBoost {
		voiceSettings["use_speaker_boost"] = true
	}

	requestBody := map[string]interface{}{
		"text":           text,
		"model_id":       "eleven_monolingual_v1",
		"voice_settings": voiceSettings,
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
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}