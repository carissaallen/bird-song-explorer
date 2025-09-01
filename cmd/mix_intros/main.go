package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	var (
		processAll   = flag.Bool("all", false, "Process all intro files")
		inputFile    = flag.String("input", "", "Single intro file to process")
		outputDir    = flag.String("output", "final_intros/with_nature", "Output directory for mixed files")
		natureSound  = flag.String("nature", "", "Type of nature sound (forest, morning_birds, gentle_rain, etc.)")
		listSounds   = flag.Bool("list", false, "List available nature sound types")
	)

	flag.Parse()

	// List available nature sounds
	if *listSounds {
		fmt.Println("Available nature sound types:")
		fmt.Println("  - forest        : General forest ambience")
		fmt.Println("  - morning_birds : Dawn chorus")
		fmt.Println("  - gentle_rain   : Soft rain sounds")
		fmt.Println("  - wind_trees    : Wind through trees")
		fmt.Println("  - stream        : Babbling brook")
		fmt.Println("  - meadow        : Open field sounds")
		fmt.Println("  - night         : Evening crickets")
		fmt.Println("\nIf not specified, sound is selected based on time of day")
		return
	}

	mixer := services.NewIntroMixer()

	// Process all intros
	if *processAll {
		fmt.Println("🎵 Processing all intro files with nature sounds...")
		if err := mixer.PreprocessAllIntros(); err != nil {
			log.Fatalf("Failed to process intros: %v", err)
		}
		fmt.Println("✅ All intros processed successfully!")
		fmt.Printf("📁 Mixed files saved to: %s\n", *outputDir)
		return
	}

	// Process single file
	if *inputFile != "" {
		fmt.Printf("🎵 Processing single file: %s\n", *inputFile)
		
		// Read input file
		introData, err := os.ReadFile(*inputFile)
		if err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		}

		// Mix with nature sounds
		mixedData, err := mixer.MixIntroWithNatureSounds(introData, *natureSound)
		if err != nil {
			log.Fatalf("Failed to mix audio: %v", err)
		}

		// Create output directory
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		// Generate output filename
		outputFile := fmt.Sprintf("%s/mixed_%s", *outputDir, *inputFile)
		
		// Save mixed file
		if err := os.WriteFile(outputFile, mixedData, 0644); err != nil {
			log.Fatalf("Failed to save mixed file: %v", err)
		}

		fmt.Printf("✅ Mixed file saved to: %s\n", outputFile)
		return
	}

	// Show usage if no action specified
	fmt.Println("Bird Song Explorer - Intro Mixer")
	fmt.Println("Mix pre-recorded intros with nature sounds")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mix_intros -all                    Process all intro files")
	fmt.Println("  mix_intros -input file.mp3         Process single file")
	fmt.Println("  mix_intros -list                   List available nature sounds")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -nature string   Specify nature sound type")
	fmt.Println("  -output string   Output directory (default: final_intros/with_nature)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mix_intros -all")
	fmt.Println("  mix_intros -input intro_00_Antoni.mp3 -nature forest")
	fmt.Println("  mix_intros -all -nature morning_birds")
}