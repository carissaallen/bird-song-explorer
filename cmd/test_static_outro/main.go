package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	// Set environment variable to use static outros
	os.Setenv("USE_STATIC_OUTROS", "true")

	// Test outro integration
	outroIntegration := services.NewOutroIntegration()

	// Test each voice with different days
	voices := []string{"Amelia", "Antoni", "Hope", "Rory", "Danielle", "Stuart"}
	days := []time.Weekday{
		time.Monday,    // joke
		time.Tuesday,   // teaser
		time.Wednesday, // wisdom
		time.Thursday,  // teaser
		time.Friday,    // joke
		time.Saturday,  // challenge
		time.Sunday,    // funfact
	}

	fmt.Println("Testing Static Outro Integration")
	fmt.Println("=================================")

	for _, voice := range voices {
		for _, day := range days {
			fmt.Printf("\nTesting %s on %s...\n", voice, day.String())

			// Generate outro (without bird song for this test)
			outroAudio, err := outroIntegration.GenerateOutroWithBirdSong(
				voice,
				day,
				nil, // No bird song for this test
				"",
			)

			if err != nil {
				log.Printf("❌ Error for %s on %s: %v", voice, day, err)
			} else {
				fmt.Printf("✅ Success: Generated %d bytes for %s on %s\n", 
					len(outroAudio), voice, day)
			}
		}
	}

	fmt.Println("\n✅ Static outro integration test complete!")
}