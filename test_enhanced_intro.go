package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	fmt.Println("🎵 Testing Enhanced Intro Generation")
	fmt.Println("=====================================")

	// Check for required environment variable
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		log.Fatal("Error: ELEVENLABS_API_KEY environment variable is not set")
	}

	// Check if sound_effects directory exists
	if _, err := os.Stat("sound_effects"); err != nil {
		log.Fatal("Error: sound_effects directory not found. Please ensure you're running from the project root.")
	}

	// Check for required sound files
	requiredFiles := []string{
		"sound_effects/ambience/forest-ambience.mp3",
		"sound_effects/ambience/jungle_sounds.mp3",
		"sound_effects/ambience/morning-birdsong.mp3",
		"sound_effects/chimes/sparkle_chime.mp3",
	}

	fmt.Println("\n✅ Checking required sound files:")
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); err != nil {
			log.Fatalf("   ❌ Missing: %s", file)
		}
		fmt.Printf("   ✓ Found: %s\n", file)
	}

	// Create the enhanced intro mixer
	mixer := services.NewEnhancedIntroMixer(elevenLabsKey)

	// Get the daily voice
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()
	
	fmt.Printf("\n🎤 Using voice: %s (ID: %s)\n", dailyVoice.Name, dailyVoice.ID)

	// Get available ambiences
	ambiences := mixer.GetAvailableAmbiences()
	fmt.Println("\n🌿 Available ambience tracks:")
	for _, amb := range ambiences {
		fmt.Printf("   • %s (%s)\n", amb.Name, amb.Path)
	}

	// Define the intro text
	introText := "Welcome, nature detectives! Time to discover an amazing bird from your neighborhood. Today's special bird is waiting to sing for you!"
	
	fmt.Printf("\n📝 Intro text: \"%s\"\n", introText)

	// Generate the enhanced intro
	fmt.Println("\n⏳ Generating enhanced intro (this may take a few seconds)...")
	
	startTime := time.Now()
	audioData, err := mixer.GenerateEnhancedIntroWithText(introText, dailyVoice.ID)
	if err != nil {
		log.Fatalf("Failed to generate enhanced intro: %v", err)
	}
	
	duration := time.Since(startTime)
	fmt.Printf("✅ Generated intro in %.2f seconds (size: %d bytes)\n", duration.Seconds(), len(audioData))

	// Get which ambience was selected
	selectedAmbience := mixer.GetSelectedAmbience()
	fmt.Printf("🎶 Selected ambience: %s\n", selectedAmbience)

	// Save the audio to a test file
	testFile := fmt.Sprintf("test_enhanced_intro_%s.mp3", time.Now().Format("20060102_150405"))
	if err := os.WriteFile(testFile, audioData, 0644); err != nil {
		log.Fatalf("Failed to save test file: %v", err)
	}
	fmt.Printf("\n💾 Saved to: %s\n", testFile)

	// Try to play the audio (macOS)
	fmt.Println("\n🔊 Attempting to play the intro...")
	
	// Check which audio player is available
	var playCmd *exec.Cmd
	if _, err := exec.LookPath("afplay"); err == nil {
		// macOS
		playCmd = exec.Command("afplay", testFile)
		fmt.Println("   Using afplay (macOS)")
	} else if _, err := exec.LookPath("play"); err == nil {
		// Linux with sox
		playCmd = exec.Command("play", testFile)
		fmt.Println("   Using play (sox)")
	} else if _, err := exec.LookPath("ffplay"); err == nil {
		// ffplay from ffmpeg
		playCmd = exec.Command("ffplay", "-nodisp", "-autoexit", testFile)
		fmt.Println("   Using ffplay")
	} else {
		fmt.Println("   ⚠️  No audio player found. Please play the file manually.")
		fmt.Printf("   File location: %s\n", testFile)
		return
	}

	if playCmd != nil {
		if err := playCmd.Run(); err != nil {
			fmt.Printf("   ⚠️  Failed to play audio: %v\n", err)
			fmt.Printf("   You can play the file manually: %s\n", testFile)
		} else {
			fmt.Println("   ✅ Playback complete!")
		}
	}

	// Test getting ambience for background (for next track)
	fmt.Println("\n🎵 Testing ambience retrieval for next track...")
	ambienceData, err := mixer.GetAmbienceForBackground()
	if err != nil {
		fmt.Printf("   ❌ Failed to get ambience: %v\n", err)
	} else {
		fmt.Printf("   ✅ Retrieved %s ambience for background use (%d bytes)\n", selectedAmbience, len(ambienceData))
	}

	fmt.Println("\n✨ Test complete!")
	fmt.Printf("📁 Generated file: %s\n", testFile)
	fmt.Println("\nYou can use this enhanced intro in your daily updates by ensuring the sound_effects")
	fmt.Println("directory is present. The system will automatically use enhanced intros when available.")
}