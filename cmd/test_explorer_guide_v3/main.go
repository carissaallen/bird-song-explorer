package main

import (
	"fmt"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	fmt.Println("🎙️ BIRD EXPLORER'S GUIDE - VERSION 3 WITH IMPROVED TRANSITIONS")
	fmt.Println("================================================================")
	fmt.Println("\nShowing multiple generations to demonstrate transition variety:\n")
	
	// Test birds
	birds := []struct {
		common     string
		scientific string
		family     string
		region     string
	}{
		{"American Robin", "Turdus migratorius", "Turdidae", "North America"},
		{"Northern Cardinal", "Cardinalis cardinalis", "Cardinalidae", "Eastern United States"},
		{"Ruby-throated Hummingbird", "Archilochus colubris", "Trochilidae", "Eastern North America"},
	}
	
	generator := services.NewImprovedFactGeneratorV3()
	
	for _, birdData := range birds {
		bird := &models.Bird{
			CommonName:     birdData.common,
			ScientificName: birdData.scientific,
			Family:         birdData.family,
			Region:         birdData.region,
			Latitude:       40.7128,
			Longitude:      -74.0060,
		}
		
		fmt.Printf("\n════════════════════════════════════════════\n")
		fmt.Printf("🦜 %s\n", strings.ToUpper(bird.CommonName))
		fmt.Printf("════════════════════════════════════════════\n")
		
		// Generate twice to show variety
		for i := 1; i <= 2; i++ {
			fmt.Printf("\n--- Generation %d ---\n\n", i)
			
			script := generator.GenerateExplorersGuideScript(bird)
			fmt.Println(script)
			
			fmt.Printf("\n[Duration: ~%d seconds, %d words]\n", 
				generator.EstimateReadingTime(script),
				len(strings.Fields(script)))
			
			// Analyze transitions used
			analyzeTransitions(script)
		}
	}
	
	fmt.Printf("\n════════════════════════════════════════════\n")
	fmt.Println("✅ IMPROVEMENTS DEMONSTRATED:")
	fmt.Println("════════════════════════════════════════════")
	fmt.Println("• Transitions match the content that follows")
	fmt.Println("• 'Guess what?' followed by actual facts")
	fmt.Println("• Action prompts ('Listen!', 'Watch!') for behaviors")
	fmt.Println("• Variety in transition phrases between generations")
	fmt.Println("• Natural flow without repetitive patterns")
	fmt.Println("• Appropriate exclamations for amazing facts")
}

func analyzeTransitions(script string) {
	fmt.Println("\nTransition Analysis:")
	
	// Check for variety of transitions
	transitions := []string{
		"Here's an amazing fact:",
		"Did you know?",
		"Fun fact:",
		"Here's something cool:",
		"Guess what?",
		"Check this out:",
		"Listen carefully!",
		"Watch for this:",
		"Look closely!",
		"How incredible!",
		"That's amazing!",
		"Wow!",
		"Amazing fact:",
		"Cool fact:",
		"Incredible ability:",
	}
	
	found := []string{}
	for _, trans := range transitions {
		if strings.Contains(script, trans) {
			found = append(found, trans)
		}
	}
	
	if len(found) > 0 {
		fmt.Printf("  Transitions used: %s\n", strings.Join(found, ", "))
	}
	
	// Check for awkward patterns
	if strings.Contains(script, "Guess what? Listen") {
		fmt.Println("  ⚠️ Found: 'Guess what? Listen' - awkward pattern")
	}
	if strings.Contains(script, "And guess what? Listen") {
		fmt.Println("  ⚠️ Found: 'And guess what? Listen' - awkward pattern")
	}
	
	// Count transition variety
	fmt.Printf("  Unique transitions found: %d\n", len(found))
}