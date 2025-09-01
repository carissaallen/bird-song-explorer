package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	fmt.Println("📍 BIRD EXPLORER'S GUIDE - LOCATION-AWARE VERSION")
	fmt.Println("==================================================")
	fmt.Println("\nDemonstrating location-specific content with eBird sightings:\n")
	
	// Get eBird API key from environment
	ebirdAPIKey := os.Getenv("EBIRD_API_KEY")
	if ebirdAPIKey == "" {
		fmt.Println("⚠️ EBIRD_API_KEY not set - using simulated data")
		ebirdAPIKey = "demo-key"
	}
	
	// Test locations (real coordinates)
	locations := []struct {
		name      string
		city      string
		state     string
		latitude  float64
		longitude float64
	}{
		{"New York City", "New York City", "New York", 40.7128, -74.0060},
		{"San Francisco", "San Francisco", "California", 37.7749, -122.4194},
		{"Austin", "Austin", "Texas", 30.2672, -97.7431},
	}
	
	// Test birds
	birds := []struct {
		common     string
		scientific string
		family     string
	}{
		{"American Robin", "Turdus migratorius", "Turdidae"},
		{"Northern Cardinal", "Cardinalis cardinalis", "Cardinalidae"},
		{"Ruby-throated Hummingbird", "Archilochus colubris", "Trochilidae"},
	}
	
	generator := services.NewImprovedFactGeneratorV4(ebirdAPIKey)
	
	// Test one bird in different locations
	testBird := birds[0]
	fmt.Printf("Testing %s in different locations:\n", testBird.common)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	for _, loc := range locations {
		bird := &models.Bird{
			CommonName:     testBird.common,
			ScientificName: testBird.scientific,
			Family:         testBird.family,
			Region:         loc.state,
			Latitude:       loc.latitude,
			Longitude:      loc.longitude,
		}
		
		fmt.Printf("\n📍 Location: %s, %s\n", loc.city, loc.state)
		fmt.Printf("   Coordinates: %.4f, %.4f\n", loc.latitude, loc.longitude)
		fmt.Println("   ─────────────────────────────────────")
		
		// Generate location-aware script
		script := generator.GenerateExplorersGuideScriptWithLocation(bird, loc.latitude, loc.longitude)
		
		// Show the script
		fmt.Println("\n" + script)
		
		fmt.Printf("\n   [Duration: ~%d seconds, %d words]\n", 
			generator.EstimateReadingTime(script),
			len(strings.Fields(script)))
		
		// Analyze location-specific content
		analyzeLocationContent(script, loc.city, loc.state)
		
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}
	
	// Demonstrate simulated eBird data
	fmt.Println("\n📊 SIMULATED eBird SIGHTING DATA")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	demonstrateSimulatedSightings()
}

func analyzeLocationContent(script, city, state string) {
	fmt.Println("\n   📍 Location-specific content:")
	
	features := []struct {
		name    string
		phrases []string
	}{
		{"City mentions", []string{city, "your city", "near you"}},
		{"State mentions", []string{state, "your state", "your area"}},
		{"Recent sightings", []string{"spotted", "seen", "sighting", "days ago", "this week"}},
		{"Local landmarks", []string{"park", "trail", "lake", "river", "garden"}},
		{"Distance info", []string{"miles from you", "nearby", "close to"}},
		{"Seasonal presence", []string{"year-round", "summer", "winter", "migration"}},
		{"Local conservation", []string{"Audubon", "bird count", "report sightings"}},
	}
	
	for _, feature := range features {
		found := false
		for _, phrase := range feature.phrases {
			if strings.Contains(strings.ToLower(script), strings.ToLower(phrase)) {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("      ✓ %s\n", feature.name)
		}
	}
}

func demonstrateSimulatedSightings() {
	fmt.Println("\nExample of location-specific content that would be generated:")
	fmt.Println()
	
	examples := []string{
		"🔹 'Great news! American Robins have been spotted near you in Central Park!'",
		"🔹 'You're in luck! A Northern Cardinal was seen just 3 days ago at Prospect Park Lake!'",
		"🔹 'Exciting! Ruby-throated Hummingbirds are active in your area - one was spotted at Brooklyn Botanic Garden recently!'",
		"🔹 'Perfect timing! Blue Jays have been seen 12 times near Manhattan this month!'",
		"🔹 'Wow! A Red-tailed Hawk was spotted less than 2.3 miles from you!'",
		"🔹 'Bird watchers saw 5 American Robins near you just this week!'",
		"🔹 'Someone saw 3 Cardinals together at Central Park Reservoir!'",
		"🔹 'In New York City, check local parks and nature trails!'",
		"🔹 'American Robins live in New York all year long!'",
		"🔹 'Join the New York Audubon Society to help protect Cardinals!'",
	}
	
	for _, example := range examples {
		fmt.Println(example)
	}
	
	fmt.Println("\n📱 eBird Integration Features:")
	fmt.Println("• Real-time sighting data within 50-mile radius")
	fmt.Println("• Number of recent observations")
	fmt.Println("• Specific location names where birds were seen")
	fmt.Println("• Days since last sighting")
	fmt.Println("• Seasonal presence patterns")
	fmt.Println("• Distance to nearest observation")
	
	fmt.Println("\n🗺️ Location Context Features:")
	fmt.Println("• City and state-specific mentions")
	fmt.Println("• Local landmark references")
	fmt.Println("• Regional habitat descriptions")
	fmt.Println("• Local conservation organizations")
	fmt.Println("• Seasonal timing for the area")
}