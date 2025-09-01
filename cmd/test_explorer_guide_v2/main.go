package main

import (
	"fmt"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	fmt.Println("🦅 BIRD EXPLORER'S GUIDE - ENHANCED VERSION")
	fmt.Println("============================================")
	fmt.Println()
	
	// Test with diverse bird species
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
		{"Blue Jay", "Cyanocitta cristata", "Corvidae", "Eastern North America"},
		{"Bald Eagle", "Haliaeetus leucocephalus", "Accipitridae", "North America"},
	}
	
	generator := services.NewImprovedFactGeneratorV2()
	
	for _, birdData := range birds {
		bird := &models.Bird{
			CommonName:     birdData.common,
			ScientificName: birdData.scientific,
			Family:         birdData.family,
			Region:         birdData.region,
			Latitude:       40.7128,  // NYC coordinates for observation data
			Longitude:      -74.0060,
		}
		
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("🦜 %s\n", strings.ToUpper(bird.CommonName))
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		
		script := generator.GenerateExplorersGuideScript(bird)
		
		// Format the script with line breaks for readability
		formatted := strings.ReplaceAll(script, ". ", ".\n\n")
		formatted = strings.ReplaceAll(formatted, "! ", "!\n\n")
		formatted = strings.ReplaceAll(formatted, "? ", "?\n\n")
		
		fmt.Println(formatted)
		
		fmt.Printf("\n📊 Statistics:\n")
		fmt.Printf("   • Length: %d characters\n", len(script))
		fmt.Printf("   • Words: ~%d\n", len(strings.Fields(script)))
		fmt.Printf("   • Estimated speech: ~%d seconds\n", generator.EstimateReadingTime(script))
		
		// Analyze content sections
		analyzeContentSections(script)
	}
	
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Println("✨ Enhanced Features Demonstrated:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("• Scientific taxonomy presented first")
	fmt.Println("• Vivid physical descriptions with comparisons")
	fmt.Println("• Detailed vocalization descriptions")
	fmt.Println("• Seasonal and habitat information")
	fmt.Println("• Nesting and baby bird facts")
	fmt.Println("• Amazing abilities and records")
	fmt.Println("• Conservation awareness")
	fmt.Println("• Engaging transitions and action words")
}

func analyzeContentSections(script string) {
	lower := strings.ToLower(script)
	
	sections := []struct {
		name string
		keywords []string
	}{
		{"Taxonomy", []string{"scientific name", "family"}},
		{"Physical", []string{"color", "size", "wing", "tail", "crest", "stripe"}},
		{"Vocalizations", []string{"song", "call", "sound", "sing", "whistle", "chirp"}},
		{"Habitat", []string{"live", "found", "habitat", "forest", "garden"}},
		{"Diet", []string{"eat", "food", "feed", "seed", "insect", "nectar"}},
		{"Nesting", []string{"nest", "egg", "baby", "chick", "hatch"}},
		{"Abilities", []string{"can", "able", "incredible", "amazing", "fastest", "largest"}},
		{"Conservation", []string{"help", "protect", "scientist", "watch"}},
	}
	
	fmt.Println("\n   📋 Content includes:")
	for _, section := range sections {
		found := false
		for _, keyword := range section.keywords {
			if strings.Contains(lower, keyword) {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("      ✓ %s\n", section.name)
		}
	}
}