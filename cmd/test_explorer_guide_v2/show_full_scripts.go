package main

import (
	"fmt"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	fmt.Println("🎙️ BIRD EXPLORER'S GUIDE - FULL SCRIPTS AS READ")
	fmt.Println("================================================")
	fmt.Println("\nThese are the complete scripts as they will be spoken:\n")
	
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
	}
	
	generator := services.NewImprovedFactGeneratorV2()
	
	for i, birdData := range birds {
		bird := &models.Bird{
			CommonName:     birdData.common,
			ScientificName: birdData.scientific,
			Family:         birdData.family,
			Region:         birdData.region,
			Latitude:       40.7128,  // NYC coordinates for observation data
			Longitude:      -74.0060,
		}
		
		fmt.Printf("\n────────────────────────────────────────────\n")
		fmt.Printf("%d. %s\n", i+1, strings.ToUpper(bird.CommonName))
		fmt.Printf("────────────────────────────────────────────\n\n")
		
		script := generator.GenerateExplorersGuideScript(bird)
		
		// Show the complete, unbroken script
		fmt.Println(script)
		
		fmt.Printf("\n\n[Duration: ~%d seconds]\n", generator.EstimateReadingTime(script))
	}
	
	fmt.Printf("\n================================================\n")
	fmt.Println("Note: These scripts flow naturally when read aloud,")
	fmt.Println("with engaging transitions between topics.")
}