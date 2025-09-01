package main

import (
	"fmt"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	fmt.Println("BIRD EXPLORER'S GUIDE - VERSION COMPARISON")
	fmt.Println("===========================================")
	fmt.Println()
	
	birds := []struct {
		common     string
		scientific string
		family     string
		region     string
	}{
		{"American Robin", "Turdus migratorius", "Turdidae", "North America"},
		{"Northern Cardinal", "Cardinalis cardinalis", "Cardinalidae", "Eastern United States"},
		{"Ruby-throated Hummingbird", "Archilochus colubris", "Trochilidae", "Eastern North America"},
		{"Great Horned Owl", "Bubo virginianus", "Strigidae", "Americas"},
		{"Pileated Woodpecker", "Dryocopus pileatus", "Picidae", "North America"},
	}
	
	for _, birdData := range birds {
		bird := &models.Bird{
			CommonName:     birdData.common,
			ScientificName: birdData.scientific,
			Family:         birdData.family,
			Region:         birdData.region,
			Latitude:       40.7128,  // NYC for testing
			Longitude:      -74.0060,
		}
		
		fmt.Printf("\n============================================\n")
		fmt.Printf("BIRD: %s\n", bird.CommonName)
		fmt.Printf("============================================\n")
		
		// Generate with original version
		fmt.Println("\n--- ORIGINAL VERSION (V1) ---")
		generatorV1 := services.NewImprovedFactGenerator()
		scriptV1 := generatorV1.GenerateExplorersGuideScript(bird)
		fmt.Println(scriptV1)
		fmt.Printf("\nLength: %d characters, ~%d words\n", len(scriptV1), len(strings.Fields(scriptV1)))
		
		// Generate with improved version
		fmt.Println("\n--- IMPROVED VERSION (V2) ---")
		generatorV2 := services.NewImprovedFactGeneratorV2()
		scriptV2 := generatorV2.GenerateExplorersGuideScript(bird)
		fmt.Println(scriptV2)
		fmt.Printf("\nLength: %d characters, ~%d words\n", len(scriptV2), len(strings.Fields(scriptV2)))
		fmt.Printf("Estimated speech time: ~%d seconds\n", generatorV2.EstimateReadingTime(scriptV2))
		
		// Analysis
		fmt.Println("\n--- CONTENT ANALYSIS ---")
		analyzeContent(scriptV2)
		
		fmt.Println("\nPress Enter for next bird...")
		fmt.Scanln()
	}
	
	fmt.Println("\n===========================================")
	fmt.Println("SUMMARY OF IMPROVEMENTS")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println("Version 2 Enhancements:")
	fmt.Println("✓ Uses both English and Simple Wikipedia for richer content")
	fmt.Println("✓ Adds vocalization descriptions (songs and calls)")
	fmt.Println("✓ Includes nesting and baby bird information")
	fmt.Println("✓ Features amazing abilities and record-breaking facts")
	fmt.Println("✓ Size comparisons to familiar objects")
	fmt.Println("✓ Action words make descriptions more vivid")
	fmt.Println("✓ Seasonal watching tips")
	fmt.Println("✓ More engaging transitions between sections")
	fmt.Println("✓ Longer, more comprehensive content (60-90 seconds vs 30-40)")
}

func analyzeContent(text string) {
	lower := strings.ToLower(text)
	
	features := map[string]bool{
		"Scientific name":     strings.Contains(lower, "scientific name"),
		"Family mentioned":    strings.Contains(lower, "family"),
		"Physical description": strings.Contains(lower, "color") || strings.Contains(lower, "size") || strings.Contains(lower, "wing"),
		"Vocalizations":       strings.Contains(lower, "song") || strings.Contains(lower, "call") || strings.Contains(lower, "sound"),
		"Diet information":    strings.Contains(lower, "eat") || strings.Contains(lower, "food") || strings.Contains(lower, "feed"),
		"Habitat described":   strings.Contains(lower, "live") || strings.Contains(lower, "habitat") || strings.Contains(lower, "found"),
		"Nesting/babies":      strings.Contains(lower, "nest") || strings.Contains(lower, "egg") || strings.Contains(lower, "baby"),
		"Amazing abilities":   strings.Contains(lower, "incredible") || strings.Contains(lower, "amazing") || strings.Contains(lower, "special"),
		"Conservation":        strings.Contains(lower, "protect") || strings.Contains(lower, "help") || strings.Contains(lower, "conservation"),
		"Seasonal info":       strings.Contains(lower, "spring") || strings.Contains(lower, "summer") || strings.Contains(lower, "winter"),
		"Size comparisons":    strings.Contains(lower, "as long as") || strings.Contains(lower, "size of") || strings.Contains(lower, "smaller than"),
		"Action words":        strings.Contains(lower, "zoom") || strings.Contains(lower, "swoop") || strings.Contains(lower, "hop") || strings.Contains(lower, "gobble"),
	}
	
	fmt.Println("Content includes:")
	for feature, present := range features {
		if present {
			fmt.Printf("  ✓ %s\n", feature)
		} else {
			fmt.Printf("  ✗ %s\n", feature)
		}
	}
}