package main

import (
	"fmt"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func testBird(birdName, scientificName, family, region string) {
	fmt.Printf("\n============================================\n")
	fmt.Printf("Testing: %s\n", birdName)
	fmt.Printf("============================================\n")
	
	// Create a test bird
	bird := &models.Bird{
		CommonName:     birdName,
		ScientificName: scientificName,
		Family:         family,
		Region:         region,
		// These would normally come from location service
		Latitude:  40.7128,  // Example: NYC
		Longitude: -74.0060,
	}
	
	// Generate the improved Explorer's Guide script
	generator := services.NewImprovedFactGenerator()
	script := generator.GenerateExplorersGuideScript(bird)
	
	fmt.Println("\n--- Generated Explorer's Guide Script ---")
	fmt.Println(script)
	
	// Show statistics
	wordCount := len(strings.Fields(script))
	charCount := len(script)
	sentences := strings.Count(script, ".") + strings.Count(script, "!") + strings.Count(script, "?")
	
	fmt.Printf("\n--- Statistics ---\n")
	fmt.Printf("Characters: %d\n", charCount)
	fmt.Printf("Words: ~%d\n", wordCount)
	fmt.Printf("Sentences: ~%d\n", sentences)
	fmt.Printf("Estimated speech duration: ~%d seconds\n", wordCount*60/150) // ~150 words per minute
}

func main() {
	fmt.Println("Bird Song Explorer - Improved Explorer's Guide Test")
	fmt.Println("====================================================")
	fmt.Println("This test demonstrates the improved fact generation for Track 3")
	fmt.Println("(Bird Explorer's Guide) with better organization and content.")
	fmt.Println()
	fmt.Println("Key improvements:")
	fmt.Println("1. Scientific name and family moved to the beginning")
	fmt.Println("2. More detailed information from Wikipedia and iNaturalist")
	fmt.Println("3. Better organization: Scientific intro → Physical → Habitat → Diet → Conservation → Fun facts")
	fmt.Println("4. Natural transitions between sections")
	fmt.Println("5. Kid-friendly language throughout")
	
	// Test with several birds
	birds := []struct {
		common     string
		scientific string
		family     string
		region     string
	}{
		{"American Robin", "Turdus migratorius", "Turdidae", "North America"},
		{"Northern Cardinal", "Cardinalis cardinalis", "Cardinalidae", "Eastern United States"},
		{"Blue Jay", "Cyanocitta cristata", "Corvidae", "Eastern North America"},
		{"Red-tailed Hawk", "Buteo jamaicensis", "Accipitridae", "North America"},
		{"Ruby-throated Hummingbird", "Archilochus colubris", "Trochilidae", "Eastern North America"},
	}
	
	for _, bird := range birds {
		testBird(bird.common, bird.scientific, bird.family, bird.region)
		fmt.Println("\n(Press Enter to continue...)")
		fmt.Scanln()
	}
	
	fmt.Println("\n====================================================")
	fmt.Println("Summary")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println("The improved Explorer's Guide provides:")
	fmt.Println()
	fmt.Println("1. **Better Structure**: Scientific name first, then organized sections")
	fmt.Println("2. **Richer Content**: Pulls from multiple Wikipedia and iNaturalist sources")
	fmt.Println("3. **Kid-Friendly**: Language adapted for young listeners")
	fmt.Println("4. **Educational Value**: Includes conservation awareness and citizen science")
	fmt.Println("5. **Natural Flow**: Smooth transitions between topics")
	fmt.Println()
	fmt.Println("This content can be generated WITHOUT calling ElevenLabs API")
	fmt.Println("until you're ready to create the final audio track.")
}