package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	var (
		introFile   = flag.String("intro", "", "Path to intro MP3 file (optional, will use sample if not provided)")
		soundType   = flag.String("type", "forest", "Type of nature sound: forest, morning_birds, gentle_rain, wind_trees, stream, meadow, night")
		outputFile  = flag.String("output", "test_intro_with_nature.mp3", "Output filename")
		fetchOnly   = flag.Bool("fetch-only", false, "Only fetch nature sound without mixing")
	)

	flag.Parse()

	fmt.Println("🌿 Bird Song Explorer - Nature Sound Mixer Test")
	fmt.Println("================================================")

	// If fetch-only mode, just download the nature sound
	if *fetchOnly {
		fmt.Printf("🔍 Fetching nature sound type: %s\n", *soundType)
		fetcher := services.NewNatureSoundFetcher()
		
		soundData, err := fetcher.GetNatureSoundByType(*soundType)
		if err != nil {
			log.Fatalf("❌ Failed to fetch nature sound: %v", err)
		}

		outputPath := fmt.Sprintf("nature_sound_%s.mp3", *soundType)
		if err := os.WriteFile(outputPath, soundData, 0644); err != nil {
			log.Fatalf("❌ Failed to save nature sound: %v", err)
		}

		fmt.Printf("✅ Nature sound saved to: %s\n", outputPath)
		fmt.Printf("📊 File size: %.2f MB\n", float64(len(soundData))/1024/1024)
		return
	}

	// Get intro audio data
	var introData []byte
	var err error

	if *introFile != "" {
		// Use provided intro file
		fmt.Printf("📁 Reading intro file: %s\n", *introFile)
		introData, err = os.ReadFile(*introFile)
		if err != nil {
			log.Fatalf("❌ Failed to read intro file: %v", err)
		}
	} else {
		// Look for a sample intro file
		possibleIntros := []string{
			"final_intros/intro_00_Antoni.mp3",
			"final_intros/intro_00_Amelia.mp3",
			"final_intros/intro_00_Charlotte.mp3",
			"test_intro_elevenlabs.mp3",
		}

		for _, path := range possibleIntros {
			if data, err := os.ReadFile(path); err == nil {
				fmt.Printf("📁 Using sample intro: %s\n", path)
				introData = data
				break
			}
		}

		if introData == nil {
			log.Fatalf("❌ No intro file found. Please provide an intro file with -intro flag or place a file at final_intros/intro_00_Antoni.mp3")
		}
	}

	// Mix with nature sounds
	fmt.Printf("🎵 Mixing intro with %s sounds...\n", *soundType)
	mixer := services.NewIntroMixer()
	
	mixedData, err := mixer.MixIntroWithNatureSounds(introData, *soundType)
	if err != nil {
		log.Fatalf("❌ Failed to mix audio: %v", err)
	}

	// Save the mixed audio
	if err := os.WriteFile(*outputFile, mixedData, 0644); err != nil {
		log.Fatalf("❌ Failed to save mixed audio: %v", err)
	}

	fmt.Println("\n✅ Success! Mixed audio created")
	fmt.Println("=====================================")
	fmt.Printf("📁 Output file: %s\n", *outputFile)
	fmt.Printf("📊 File size: %.2f MB\n", float64(len(mixedData))/1024/1024)
	fmt.Printf("🎧 Nature sound type: %s\n", *soundType)
	fmt.Println("\n🎵 Audio Structure (auto-adjusted to intro length):")
	fmt.Println("  0-2s   : Nature sounds fade in (25% volume)")
	fmt.Println("  2s     : Voice intro begins")
	fmt.Println("  2-6s   : Voice with nature background (10% volume)")
	fmt.Println("  6-7s   : Fade out after voice ends")
	
	// Get absolute path for easy access
	absPath, _ := filepath.Abs(*outputFile)
	fmt.Printf("\n🎧 Play with: open '%s'\n", absPath)
	fmt.Println("   or: afplay " + *outputFile + " (on macOS)")
}