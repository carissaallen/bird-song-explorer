package services

import (
	"strings"
	"testing"
	"time"

	"github.com/callen/bird-song-explorer/internal/config"
)

func TestGetRandomIntroURL_ConsistencyWithinDay(t *testing.T) {
	am := NewAudioManager()
	baseURL := "https://example.com"

	// Get the daily voice to know what to expect
	voiceManager := config.NewVoiceManager()
	dailyVoice := voiceManager.GetDailyVoice()

	// Call the method multiple times
	firstURL, firstVoiceID := am.GetRandomIntroURL(baseURL)
	secondURL, secondVoiceID := am.GetRandomIntroURL(baseURL)
	thirdURL, thirdVoiceID := am.GetRandomIntroURL(baseURL)

	// All calls should return the same intro URL and voice ID
	if firstURL != secondURL || secondURL != thirdURL {
		t.Errorf("GetRandomIntroURL not consistent within the same day. Got %s, %s, %s",
			firstURL, secondURL, thirdURL)
	}

	// All calls should return the same voice ID
	if firstVoiceID != secondVoiceID || secondVoiceID != thirdVoiceID {
		t.Errorf("Voice ID not consistent within the same day. Got %s, %s, %s",
			firstVoiceID, secondVoiceID, thirdVoiceID)
	}

	// The voice ID should match the daily voice
	if firstVoiceID != dailyVoice.ID {
		t.Errorf("Voice ID doesn't match daily voice. Got %s, expected %s",
			firstVoiceID, dailyVoice.ID)
	}

	// The intro should contain the daily voice name
	if !strings.Contains(firstURL, dailyVoice.Name) {
		// Check if it's the fallback case (Antoni)
		if !strings.Contains(firstURL, "Antoni") {
			t.Errorf("Intro URL doesn't contain the expected voice name. Got %s, expected voice: %s",
				firstURL, dailyVoice.Name)
		}
	}

	// Verify the URL format
	if !strings.HasPrefix(firstURL, baseURL+"/audio/intros/intro_") {
		t.Errorf("Intro URL has unexpected format: %s", firstURL)
	}
}

func TestVoiceConsistencyAcrossServices(t *testing.T) {
	// This test verifies that all services use the same voice for a given day

	// Get the voice from VoiceManager
	voiceManager := config.NewVoiceManager()
	dailyVoice1 := voiceManager.GetDailyVoice()

	// Get it again to ensure consistency
	dailyVoice2 := voiceManager.GetDailyVoice()

	if dailyVoice1.ID != dailyVoice2.ID {
		t.Errorf("VoiceManager returning different voices on same day: %s vs %s",
			dailyVoice1.ID, dailyVoice2.ID)
	}

	// Test that AudioManager uses the same voice
	am := NewAudioManager()
	introURL, voiceID := am.GetRandomIntroURL("https://example.com")

	// The voice ID should match exactly
	if voiceID != dailyVoice1.ID {
		t.Errorf("AudioManager returning different voice ID than VoiceManager. Got: %s, Expected: %s",
			voiceID, dailyVoice1.ID)
	}

	// The intro should match the daily voice (or fallback to Antoni)
	if !strings.Contains(introURL, dailyVoice1.Name) && !strings.Contains(introURL, "Antoni") {
		t.Errorf("AudioManager using different voice than VoiceManager. URL: %s, Expected voice: %s",
			introURL, dailyVoice1.Name)
	}
}

func TestDeterministicIntroSelection(t *testing.T) {
	am := NewAudioManager()
	baseURL := "https://example.com"

	// Get today's selection
	todayIntro, todayVoiceID := am.GetRandomIntroURL(baseURL)

	// Create a fixed day seed for testing
	now := time.Now()
	daySeed := now.Year()*10000 + int(now.Month())*100 + now.Day()

	// Multiple calls on the same day should return the same intro and voice ID
	for i := 0; i < 10; i++ {
		intro, voiceID := am.GetRandomIntroURL(baseURL)
		if intro != todayIntro {
			t.Errorf("Intro selection not deterministic. Expected %s, got %s on iteration %d",
				todayIntro, intro, i)
		}
		if voiceID != todayVoiceID {
			t.Errorf("Voice ID not deterministic. Expected %s, got %s on iteration %d",
				todayVoiceID, voiceID, i)
		}
	}

	t.Logf("Day seed: %d, Selected intro: %s, Voice ID: %s", daySeed, todayIntro, todayVoiceID)
}
