package main

import (
	"fmt"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func compareFactGeneration(birdName, scientificName, family, region string) {
	fmt.Printf("\n################################################\n")
	fmt.Printf("# %s (%s)\n", birdName, scientificName)
	fmt.Printf("################################################\n")
	
	bird := &models.Bird{
		CommonName:     birdName,
		ScientificName: scientificName,
		Family:         family,
		Region:         region,
		Latitude:       40.7128,
		Longitude:      -74.0060,
	}
	
	// 1. Show current/old fact generation
	fmt.Println("\n=== CURRENT IMPLEMENTATION (generateBirdFacts) ===")
	
	// Note: This would normally be private, but for demo we'd need to make it public
	// or copy the logic here
	fmt.Println("Current facts are generic:")
	fmt.Printf("1. The %s's scientific name is %s.\n", bird.CommonName, bird.ScientificName)
	if family != "" {
		fmt.Printf("2. It belongs to the %s family.\n", family)
	}
	fmt.Printf("3. This bird can be found in %s.\n", region)
	fmt.Println("4. Listen carefully to hear its distinctive call!")
	fmt.Println("5. Birds use songs to communicate with each other.")
	
	fmt.Println("\nCurrent description for ElevenLabs (from generateBirdDescription):")
	fmt.Println("Did you know? [generic facts]. Isn't that amazing? Nature is full of wonderful surprises!")
	
	// 2. Show improved fact generation
	fmt.Println("\n=== IMPROVED IMPLEMENTATION (GenerateExplorersGuideScript) ===")
	generator := services.NewImprovedFactGenerator()
	script := generator.GenerateExplorersGuideScript(bird)
	
	fmt.Println("Full Explorer's Guide Script:")
	fmt.Println("---")
	
	// Format the script nicely
	sentences := strings.Split(script, ". ")
	for i, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			fmt.Printf("%d. %s.\n", i+1, sentence)
		}
	}
	
	// Show statistics
	fmt.Println("\n--- Statistics Comparison ---")
	oldText := "Did you know? The American Robin's scientific name is Turdus migratorius. It belongs to the Turdidae family. This bird can be found in North America. Listen carefully to hear its distinctive call! Birds use songs to communicate with each other. Isn't that amazing? Nature is full of wonderful surprises!"
	
	fmt.Printf("Old approach: ~%d words, ~%d characters\n", 
		len(strings.Fields(oldText)), len(oldText))
	fmt.Printf("New approach: ~%d words, ~%d characters\n", 
		len(strings.Fields(script)), len(script))
	fmt.Printf("Content improvement: %.1fx more detailed\n", 
		float64(len(script))/float64(len(oldText)))
}

func main() {
	fmt.Println("========================================================")
	fmt.Println("Bird Song Explorer - Fact Generation Comparison")
	fmt.Println("========================================================")
	fmt.Println()
	fmt.Println("This test compares the CURRENT vs IMPROVED fact generation")
	fmt.Println("for the Bird Explorer's Guide (Track 3) WITHOUT using ElevenLabs API.")
	fmt.Println()
	fmt.Println("Key Improvements:")
	fmt.Println("✓ Scientific name and family moved to beginning")
	fmt.Println("✓ Real Wikipedia content (physical description, habitat, diet)")
	fmt.Println("✓ iNaturalist conservation status and observations")
	fmt.Println("✓ More engaging and educational content")
	fmt.Println("✓ Better structure and flow")
	
	// Test with a few birds
	birds := []struct {
		common     string
		scientific string
		family     string
		region     string
	}{
		{"American Robin", "Turdus migratorius", "Turdidae", "North America"},
		{"Northern Cardinal", "Cardinalis cardinalis", "Cardinalidae", "Eastern United States"},
		{"Blue Jay", "Cyanocitta cristata", "Corvidae", "Eastern North America"},
	}
	
	for _, bird := range birds {
		compareFactGeneration(bird.common, bird.scientific, bird.family, bird.region)
	}
	
	fmt.Println("\n========================================================")
	fmt.Println("IMPLEMENTATION GUIDE")
	fmt.Println("========================================================")
	fmt.Println()
	fmt.Println("To integrate the improved facts into your application:")
	fmt.Println()
	fmt.Println("1. In bird_selector.go, update enrichWithWikipedia() to use:")
	fmt.Println("   generator := NewImprovedFactGenerator()")
	fmt.Println("   bird.Description = generator.GenerateExplorersGuideScript(bird)")
	fmt.Println()
	fmt.Println("2. In content_update.go, the generateBirdDescription() method")
	fmt.Println("   will receive the full script instead of simple facts")
	fmt.Println()
	fmt.Println("3. The script already includes transitions and is ready for")
	fmt.Println("   ElevenLabs TTS conversion when you're ready")
	fmt.Println()
	fmt.Println("4. You can test the content generation WITHOUT using")
	fmt.Println("   ElevenLabs credits by running this test script")
}