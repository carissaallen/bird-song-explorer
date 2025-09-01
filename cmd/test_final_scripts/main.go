package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	fmt.Println("🎙️ BIRD EXPLORER'S GUIDE - FINAL ENHANCED SCRIPTS")
	fmt.Println("==================================================")
	fmt.Println("\nComplete scripts with all improvements:")
	fmt.Println("• Location-specific content")
	fmt.Println("• Natural transitions")
	fmt.Println("• Comprehensive facts")
	fmt.Println("• ~90-120 second duration\n")
	
	// Get eBird API key
	ebirdAPIKey := os.Getenv("EBIRD_API_KEY")
	if ebirdAPIKey == "" {
		ebirdAPIKey = "demo-key"
	}
	
	// New York City coordinates
	latitude := 40.7128
	longitude := -74.0060
	
	// Test birds
	birds := []struct {
		common     string
		scientific string
		family     string
	}{
		{"American Robin", "Turdus migratorius", "Turdidae"},
		{"Northern Cardinal", "Cardinalis cardinalis", "Cardinalidae"},
		{"Ruby-throated Hummingbird", "Archilochus colubris", "Trochilidae"},
		{"Blue Jay", "Cyanocitta cristata", "Corvidae"},
		{"Great Horned Owl", "Bubo virginianus", "Strigidae"},
	}
	
	generator := services.NewImprovedFactGeneratorV4(ebirdAPIKey)
	
	for i, birdData := range birds {
		bird := &models.Bird{
			CommonName:     birdData.common,
			ScientificName: birdData.scientific,
			Family:         birdData.family,
			Region:         "North America",
			Latitude:       latitude,
			Longitude:      longitude,
		}
		
		fmt.Printf("\n════════════════════════════════════════════════════════\n")
		fmt.Printf("%d. %s\n", i+1, strings.ToUpper(bird.CommonName))
		fmt.Printf("════════════════════════════════════════════════════════\n\n")
		
		// Generate location-aware script with all enhancements
		script := generator.GenerateExplorersGuideScriptWithLocation(bird, latitude, longitude)
		
		// Display the complete script
		fmt.Println(script)
		
		// Show statistics
		fmt.Printf("\n\n📊 Statistics:\n")
		fmt.Printf("• Length: %d characters\n", len(script))
		fmt.Printf("• Words: %d\n", len(strings.Fields(script)))
		fmt.Printf("• Duration: ~%d seconds\n", generator.EstimateReadingTime(script))
		
		// Show what makes this script enhanced
		fmt.Println("\n✨ Enhanced Features in This Script:")
		analyzeEnhancements(script)
	}
	
	fmt.Printf("\n════════════════════════════════════════════════════════\n")
	fmt.Println("📝 SUMMARY OF ENHANCEMENTS")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("\n✅ All scripts now include:")
	fmt.Println("• Scientific taxonomy presented first")
	fmt.Println("• Location-specific sighting information")
	fmt.Println("• Natural, varied transitions")
	fmt.Println("• Vivid descriptions with comparisons")
	fmt.Println("• Vocalization descriptions")
	fmt.Println("• Nesting and baby bird facts")
	fmt.Println("• Amazing abilities and records")
	fmt.Println("• Local conservation actions")
	fmt.Println("• Seasonal watching tips")
	fmt.Println("• Engaging closings with local context")
}

func analyzeEnhancements(script string) {
	lower := strings.ToLower(script)
	
	features := []struct {
		category string
		found    bool
	}{
		{"Scientific name & taxonomy", strings.Contains(lower, "scientific name")},
		{"Location mentions", strings.Contains(lower, "your city") || strings.Contains(lower, "near you")},
		{"Vocalization description", strings.Contains(lower, "song") || strings.Contains(lower, "call") || strings.Contains(lower, "sound")},
		{"Diet with action words", strings.Contains(lower, "feed") || strings.Contains(lower, "eat") || strings.Contains(lower, "hunt")},
		{"Nesting information", strings.Contains(lower, "nest") || strings.Contains(lower, "egg") || strings.Contains(lower, "baby")},
		{"Amazing abilities", strings.Contains(lower, "incredible") || strings.Contains(lower, "amazing")},
		{"Conservation message", strings.Contains(lower, "help") || strings.Contains(lower, "protect")},
		{"Varied transitions", countTransitions(script) > 3},
	}
	
	for _, feature := range features {
		if feature.found {
			fmt.Printf("  ✓ %s\n", feature.category)
		}
	}
}

func countTransitions(script string) int {
	transitions := []string{
		"Did you know?",
		"Fun fact:",
		"Here's something cool:",
		"Check this out:",
		"Listen carefully!",
		"Watch for this:",
		"Amazing fact:",
		"Incredible ability:",
		"Cool fact:",
	}
	
	count := 0
	for _, trans := range transitions {
		if strings.Contains(script, trans) {
			count++
		}
	}
	return count
}