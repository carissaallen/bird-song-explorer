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
	fmt.Println("🎵 Testing Ambience Continuity Between Tracks")
	fmt.Println("==============================================")

	// Check for required environment variable
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		log.Fatal("Error: ELEVENLABS_API_KEY environment variable is not set")
	}

	// Check if sound_effects directory exists
	if _, err := os.Stat("sound_effects"); err != nil {
		log.Fatal("Error: sound_effects directory not found")
	}

	// Get the daily voice
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()
	fmt.Printf("\n🎤 Using voice: %s (ID: %s)\n", dailyVoice.Name, dailyVoice.ID)

	// Test bird name
	birdName := "American Robin"
	fmt.Printf("🐦 Test bird: %s\n", birdName)

	// TRACK 1: Generate enhanced intro
	fmt.Println("\n━━━ TRACK 1: Introduction ━━━")
	introMixer := services.NewEnhancedIntroMixer(elevenLabsKey)
	
	introText := "Welcome, nature detectives! Time to discover an amazing bird from your neighborhood."
	fmt.Printf("📝 Intro text: \"%s\"\n", introText)
	
	fmt.Println("⏳ Generating enhanced intro...")
	introData, err := introMixer.GenerateEnhancedIntroWithText(introText, dailyVoice.ID)
	if err != nil {
		log.Fatalf("Failed to generate intro: %v", err)
	}
	
	selectedAmbience := introMixer.GetSelectedAmbience()
	fmt.Printf("✅ Generated intro with %s ambience (size: %d bytes)\n", selectedAmbience, len(introData))
	
	// Save Track 1
	track1File := fmt.Sprintf("test_track1_intro_%s.mp3", time.Now().Format("20060102_150405"))
	if err := os.WriteFile(track1File, introData, 0644); err != nil {
		log.Fatalf("Failed to save track 1: %v", err)
	}
	fmt.Printf("💾 Saved Track 1: %s\n", track1File)

	// TRACK 2: Generate announcement with continuing ambience
	fmt.Println("\n━━━ TRACK 2: Today's Bird ━━━")
	
	// Get the ambience data for continuation
	ambienceData, err := introMixer.GetAmbienceForBackground()
	if err != nil {
		log.Fatalf("Failed to get ambience data: %v", err)
	}
	fmt.Printf("🎶 Retrieved %s ambience for Track 2 (%d bytes)\n", selectedAmbience, len(ambienceData))
	
	announcementMixer := services.NewEnhancedBirdAnnouncement(elevenLabsKey)
	
	fmt.Println("⏳ Generating announcement with continuing ambience...")
	announcementData, err := announcementMixer.GenerateAnnouncementWithAmbience(
		birdName, 
		dailyVoice.ID,
		selectedAmbience,
	)
	if err != nil {
		log.Fatalf("Failed to generate announcement: %v", err)
	}
	
	fmt.Printf("✅ Generated announcement with fading ambience (size: %d bytes)\n", len(announcementData))
	
	// Save Track 2
	track2File := fmt.Sprintf("test_track2_announcement_%s.mp3", time.Now().Format("20060102_150405"))
	if err := os.WriteFile(track2File, announcementData, 0644); err != nil {
		log.Fatalf("Failed to save track 2: %v", err)
	}
	fmt.Printf("💾 Saved Track 2: %s\n", track2File)

	// TRACK 3: Bird song (no ambience, just the actual bird recording)
	fmt.Println("\n━━━ TRACK 3: Bird Song ━━━")
	fmt.Println("ℹ️  Track 3 would be the actual bird song recording")
	fmt.Println("   No ambience mixing - listeners hear the pure bird song")

	// Combine tracks for continuous playback test
	fmt.Println("\n━━━ Creating Combined Test File ━━━")
	combinedFile := fmt.Sprintf("test_combined_tracks_%s.mp3", time.Now().Format("20060102_150405"))
	
	// Use ffmpeg to concatenate the tracks
	concatCmd := exec.Command("ffmpeg",
		"-i", track1File,
		"-i", track2File,
		"-filter_complex", "[0:a][1:a]concat=n=2:v=0:a=1[out]",
		"-map", "[out]",
		"-c:a", "libmp3lame",
		"-b:a", "192k",
		"-ar", "44100",
		"-y",
		combinedFile,
	)
	
	if err := concatCmd.Run(); err != nil {
		fmt.Printf("⚠️  Failed to create combined file: %v\n", err)
		fmt.Println("   You can play the tracks individually to test continuity")
	} else {
		fmt.Printf("✅ Created combined file: %s\n", combinedFile)
	}

	// Try to play the combined file or individual tracks
	fmt.Println("\n🔊 Playing the tracks...")
	
	var playCmd *exec.Cmd
	fileToPlay := combinedFile
	
	// Check if combined file exists, otherwise play tracks separately
	if _, err := os.Stat(combinedFile); err != nil {
		fileToPlay = track1File
	}
	
	if _, err := exec.LookPath("afplay"); err == nil {
		// macOS
		playCmd = exec.Command("afplay", fileToPlay)
		fmt.Println("   Using afplay (macOS)")
	} else if _, err := exec.LookPath("play"); err == nil {
		// Linux with sox
		playCmd = exec.Command("play", fileToPlay)
		fmt.Println("   Using play (sox)")
	} else if _, err := exec.LookPath("ffplay"); err == nil {
		// ffplay from ffmpeg
		playCmd = exec.Command("ffplay", "-nodisp", "-autoexit", fileToPlay)
		fmt.Println("   Using ffplay")
	}

	if playCmd != nil {
		fmt.Printf("   Playing: %s\n", fileToPlay)
		if err := playCmd.Run(); err != nil {
			fmt.Printf("   ⚠️  Playback failed: %v\n", err)
		}
		
		// If playing individual tracks, play track 2 next
		if fileToPlay == track1File {
			fmt.Println("\n   Playing Track 2...")
			playCmd = exec.Command("afplay", track2File)
			playCmd.Run()
		}
	} else {
		fmt.Println("   ⚠️  No audio player found")
	}

	// Summary
	fmt.Println("\n✨ Test Summary")
	fmt.Println("================")
	fmt.Printf("Track 1 (Intro):        %s\n", track1File)
	fmt.Printf("  - Ambience: %s (fades in, plays softly behind voice)\n", selectedAmbience)
	fmt.Printf("Track 2 (Announcement): %s\n", track2File)
	fmt.Printf("  - Ambience: %s (continues from Track 1, fades out)\n", selectedAmbience)
	fmt.Println("Track 3 (Bird Song):    [Would be actual bird recording]")
	fmt.Println("  - No ambience (pure bird song)")
	
	if _, err := os.Stat(combinedFile); err == nil {
		fmt.Printf("\nCombined file: %s\n", combinedFile)
		fmt.Println("Listen for the smooth ambience transition between Track 1 and Track 2!")
	}
	
	fmt.Println("\n🎧 The ambience should:")
	fmt.Println("  1. Fade in during Track 1 (Introduction)")
	fmt.Println("  2. Continue softly behind the voice in Track 1")
	fmt.Println("  3. Continue seamlessly into Track 2 (Today's Bird)")
	fmt.Println("  4. Gradually fade out during Track 2")
	fmt.Println("  5. Be completely silent for Track 3 (Bird Song)")
}