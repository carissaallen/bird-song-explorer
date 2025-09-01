package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	var (
		timezone = flag.String("tz", "America/New_York", "Timezone to test (e.g., America/New_York, Europe/London)")
		showAll  = flag.Bool("all", false, "Show all timezones")
		testMix  = flag.Bool("mix", false, "Test audio mixing for this timezone")
	)

	flag.Parse()

	fmt.Println("🌍 Bird Song Explorer - User Timezone Test")
	fmt.Println("===========================================")

	if *showAll {
		testAllTimezones()
		return
	}

	// Test specific timezone
	testTimezone(*timezone)

	if *testMix {
		fmt.Println("\n🎵 Testing Audio Mixing")
		testAudioMixing(*timezone)
	}
}

func testTimezone(timezone string) {
	timeHelper := services.NewUserTimeHelper()
	
	// Get user's local time
	userTime := timeHelper.GetUserLocalTime(timezone)
	serverTime := time.Now()
	
	// Get time context
	context := timeHelper.GetUserTimeContext(timezone)
	
	fmt.Printf("\n📍 Timezone: %s\n", timezone)
	fmt.Printf("🕐 Server Time: %s\n", serverTime.Format("15:04:05 MST"))
	fmt.Printf("👤 User Time:   %s\n", userTime.Format("15:04:05 MST"))
	fmt.Printf("📊 Hour:        %d\n", context["hour"])
	fmt.Printf("☀️ Time Period: %s\n", context["time_period"])
	fmt.Printf("👋 Greeting:    %s\n", context["greeting"])
	fmt.Printf("🌳 Is Daytime:  %v\n", context["is_daytime"])
	fmt.Printf("🎵 Nature Sound: %s\n", context["nature_sound"])
	
	// Calculate time difference
	timeDiff := userTime.Sub(serverTime).Hours()
	fmt.Printf("⏰ Time Difference: %.1f hours\n", timeDiff)
	
	// Show what sound would be selected at different times
	fmt.Println("\n🎼 Nature Sounds Throughout the Day:")
	testHours := []int{0, 6, 9, 12, 15, 18, 21}
	for _, hour := range testHours {
		// Get nature sound for this hour
		sound := getNatureSoundForHour(hour)
		fmt.Printf("  %02d:00 - %s\n", hour, sound)
	}
}

func getNatureSoundForHour(hour int) string {
	switch {
	case hour >= 5 && hour < 9:
		return "🐦 morning_birds (dawn chorus)"
	case hour >= 9 && hour < 12:
		return "🌲 forest (active forest)"
	case hour >= 12 && hour < 17:
		return "🌾 meadow (open field)"
	case hour >= 17 && hour < 20:
		return "🌧️ gentle_rain (calming)"
	case hour >= 20 && hour < 22:
		return "💧 stream (peaceful water)"
	default:
		return "🦉 night (crickets & owls)"
	}
}

func testAllTimezones() {
	timezones := []string{
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"America/Anchorage",
		"Pacific/Honolulu",
		"America/Toronto",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Australia/Sydney",
		"Asia/Tokyo",
		"Asia/Singapore",
	}

	fmt.Println("\n🌐 Testing All Timezones")
	fmt.Println("========================")
	
	timeHelper := services.NewUserTimeHelper()
	serverTime := time.Now()
	
	fmt.Printf("Server Time: %s\n\n", serverTime.Format("15:04:05 MST"))
	
	fmt.Printf("%-25s | %-15s | %-5s | %-15s | %-15s\n", 
		"Timezone", "Local Time", "Hour", "Nature Sound", "Time Period")
	fmt.Println(string(make([]byte, 90, 90)))
	
	for _, tz := range timezones {
		userTime := timeHelper.GetUserLocalTime(tz)
		context := timeHelper.GetUserTimeContext(tz)
		
		fmt.Printf("%-25s | %-15s | %-5v | %-15s | %-15s\n",
			tz,
			userTime.Format("15:04 MST"),
			context["hour"],
			context["nature_sound"],
			context["time_period"],
		)
	}
}

func testAudioMixing(timezone string) {
	fmt.Printf("\nTesting audio mixing for timezone: %s\n", timezone)
	
	// Initialize services
	timeHelper := services.NewUserTimeHelper()
	
	// Get nature sound for this timezone
	natureSoundType := timeHelper.GetNatureSoundForUserTime(timezone)
	userHour := timeHelper.GetUserLocalHour(timezone)
	
	fmt.Printf("User's local hour: %d\n", userHour)
	fmt.Printf("Selected nature sound: %s\n", natureSoundType)
	
	// Simulate mixing (without actual audio files)
	fmt.Println("\nSimulated mixing process:")
	fmt.Printf("1. Fetch nature sound: %s\n", natureSoundType)
	fmt.Printf("2. Load intro audio\n")
	fmt.Printf("3. Mix with ffmpeg:\n")
	fmt.Printf("   - 0-2.5s: Nature sound fade in (30%% volume)\n")
	fmt.Printf("   - 2.5s: Voice intro begins\n")
	fmt.Printf("   - 2.5-30s: Voice + nature background (10%% volume)\n")
	fmt.Printf("   - 28-30s: Fade out\n")
	
	// Log the timezone usage
	logger := services.NewTimezoneLogger()
	logger.LogTimezoneUsage("test-device", "test-card", timezone, "Test Location")
	
	fmt.Println("\n✅ Timezone usage logged successfully")
}