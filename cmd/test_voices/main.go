package main

import (
	"fmt"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	// Test VoiceManager
	fmt.Println("Testing Voice Manager:")
	fmt.Println("======================")
	
	voiceManager := config.NewVoiceManager()
	
	// Show all available voices
	fmt.Println("\nAll Available Voices:")
	for i, voice := range voiceManager.GetAvailableVoices() {
		fmt.Printf("%d. %s (%s) - ID: %s\n", i+1, voice.Name, voice.Region, voice.ID)
	}
	
	// Test daily voice selection
	fmt.Println("\nDaily Voice Selection (deterministic by date):")
	dailyVoice := voiceManager.GetDailyVoice()
	fmt.Printf("Today's voice: %s (%s)\n", dailyVoice.Name, dailyVoice.Region)
	
	// Simulate different days to show rotation
	fmt.Println("\nVoice Rotation Pattern (next 10 days):")
	now := time.Now()
	for i := 0; i < 10; i++ {
		future := now.AddDate(0, 0, i)
		seed := future.Year()*10000 + int(future.Month())*100 + future.Day()
		voiceIndex := seed % len(voiceManager.GetAvailableVoices())
		voice := voiceManager.GetAvailableVoices()[voiceIndex]
		fmt.Printf("  %s: %s (%s)\n", future.Format("2006-01-02"), voice.Name, voice.Region)
	}
	
	// Test NarrationManager
	fmt.Println("\nTesting Narration Manager:")
	fmt.Println("==========================")
	
	narrationManager := services.NewNarrationManager("test-key")
	
	// Test daily voice selection
	voiceConfig := narrationManager.SelectDailyVoice()
	fmt.Printf("Selected Voice: %s (ID: %s)\n", voiceConfig.Name, voiceConfig.VoiceID)
	
	// Test random voice selection
	fmt.Println("\nRandom Voice Selection (5 samples):")
	for i := 0; i < 5; i++ {
		randomVoice := narrationManager.GetRandomVoice()
		fmt.Printf("  Random %d: %s\n", i+1, randomVoice.Name)
	}
	
	// Test AudioManager intro selection
	fmt.Println("\nTesting Audio Manager:")
	fmt.Println("======================")
	
	audioManager := services.NewAudioManager()
	baseURL := "https://example.com"
	
	introURL, voiceID := audioManager.GetRandomIntroURL(baseURL)
	fmt.Printf("Selected Intro: %s\n", introURL)
	fmt.Printf("Voice ID: %s\n", voiceID)
	
	// Check if intro matches daily voice
	if voiceID == dailyVoice.ID {
		fmt.Println("✓ Audio Manager correctly uses daily voice")
	} else {
		fmt.Println("✗ Voice mismatch between managers!")
	}
}