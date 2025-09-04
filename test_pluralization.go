package main

import (
	"fmt"
	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	// Create a test bird
	bird := &models.Bird{
		CommonName:     "Common Merganser",
		ScientificName: "Mergus merganser",
	}

	// Create fact generator
	fg := services.NewImprovedFactGeneratorV4("")

	// Test location context with 1 sighting
	locationContext := services.LocationContext{
		CityName:  "Bend",
		StateName: "Oregon",
		RecentSightings: []services.RecentSighting{
			{
				LocationName: "Drake Park",
				Date:         "2025-09-01",
				Count:        1,
				DaysAgo:      1,
			},
		},
		Distance: 1.0,
	}

	// Generate the script
	_ = locationContext // Mark as used
	script := fg.GenerateExplorersGuideScriptWithLocation(bird, 44.0582, -121.3153)

	// Check for pluralization issues
	fmt.Println("Testing pluralization with 1 sighting:")
	fmt.Println("========================================")
	
	// Extract relevant portions
	if contains(script, "1 times") {
		fmt.Println("❌ FOUND BUG: '1 times' instead of '1 time'")
	} else if contains(script, "1 time") {
		fmt.Println("✅ CORRECT: '1 time' (singular)")
	}

	if contains(script, "1 miles") {
		fmt.Println("❌ FOUND BUG: '1 miles' instead of '1 mile'")
	} else if contains(script, "1 mile") || contains(script, "less than 1.0 mile") {
		fmt.Println("✅ CORRECT: '1 mile' (singular)")
	}

	if contains(script, "1 sightings") {
		fmt.Println("❌ FOUND BUG: '1 sightings' instead of '1 sighting'")  
	} else if contains(script, "1 sighting") || contains(script, "1 recent sighting") {
		fmt.Println("✅ CORRECT: '1 sighting' (singular)")
	}

	// Print relevant excerpts
	fmt.Println("\nScript excerpt (first 500 chars):")
	if len(script) > 500 {
		fmt.Println(script[:500] + "...")
	} else {
		fmt.Println(script)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}