package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TestBirdFacts fetches and displays all available data for a bird
func TestBirdFacts(birdName string) {
	fmt.Printf("\n========================================\n")
	fmt.Printf("COMPREHENSIVE FACT TEST: %s\n", birdName)
	fmt.Printf("========================================\n")
	
	// 1. English Wikipedia REST API
	fmt.Println("\n1. ENGLISH WIKIPEDIA (REST API)")
	fmt.Println("--------------------------------")
	fetchWikipediaREST(birdName, "en")
	
	// 2. Simple Wikipedia REST API  
	fmt.Println("\n2. SIMPLE WIKIPEDIA (REST API)")
	fmt.Println("--------------------------------")
	fetchWikipediaREST(birdName, "simple")
	
	// 3. iNaturalist Full Data
	fmt.Println("\n3. INATURALIST FULL DATA")
	fmt.Println("--------------------------------")
	fetchINaturalistFull(birdName)
	
	// 4. Show what we're currently extracting
	fmt.Println("\n4. CURRENT EXTRACTION ANALYSIS")
	fmt.Println("--------------------------------")
	analyzeCurrentExtraction(birdName)
}

func fetchWikipediaREST(birdName, lang string) {
	encodedName := strings.ReplaceAll(birdName, " ", "_")
	apiURL := fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1/page/summary/%s", lang, url.QueryEscape(encodedName))
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "BirdFactTest/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		fmt.Println("Not found")
		return
	}
	
	var result struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Extract     string `json:"extract"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}
	
	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("Description: %s\n", result.Description)
	fmt.Printf("Extract (%d chars): %s\n", len(result.Extract), result.Extract)
	
	// Analyze content types
	fmt.Println("\n  Content Analysis:")
	analyzeContent(result.Extract)
}

func fetchINaturalistFull(birdName string) {
	encodedName := url.QueryEscape(birdName)
	apiURL := fmt.Sprintf("https://api.inaturalist.org/v1/taxa?q=%s&iconic_taxa=Aves&per_page=1", encodedName)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "BirdFactTest/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	var result struct {
		Results []struct {
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			PreferredCommonName string `json:"preferred_common_name"`
			WikipediaSummary    string `json:"wikipedia_summary"`
			ConservationStatus  *struct {
				Status     string `json:"status"`
				StatusName string `json:"status_name"`
			} `json:"conservation_status"`
			DefaultPhoto *struct {
				Medium string `json:"medium_url"`
			} `json:"default_photo"`
			AncestorIDs []int `json:"ancestor_ids"`
			ObservationsCount int `json:"observations_count"`
		} `json:"results"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}
	
	if len(result.Results) == 0 {
		fmt.Println("No results found")
		return
	}
	
	taxon := result.Results[0]
	fmt.Printf("Scientific Name: %s\n", taxon.Name)
	fmt.Printf("Common Name: %s\n", taxon.PreferredCommonName)
	fmt.Printf("Observation Count: %d\n", taxon.ObservationsCount)
	
	if taxon.ConservationStatus != nil {
		fmt.Printf("Conservation: %s (%s)\n", taxon.ConservationStatus.StatusName, taxon.ConservationStatus.Status)
	}
	
	if taxon.WikipediaSummary != "" {
		fmt.Printf("Wikipedia Summary from iNat: %s\n", taxon.WikipediaSummary)
	}
	
	if taxon.DefaultPhoto != nil {
		fmt.Printf("Has Photo: Yes\n")
	}
}

func analyzeContent(text string) {
	lower := strings.ToLower(text)
	
	// Check for different types of content
	categories := map[string][]string{
		"Physical": {"color", "colour", "size", "length", "wing", "tail", "beak", "bill", "plumage", "feather", "cm", "inch", "orange", "red", "black", "white", "brown"},
		"Behavior": {"sing", "song", "call", "fly", "perch", "hop", "walk", "swim", "dive", "hunt", "forage", "migrate", "nest", "breed", "mate"},
		"Diet": {"eat", "feed", "diet", "food", "seed", "insect", "worm", "berry", "fruit", "nectar", "prey", "fish"},
		"Habitat": {"habitat", "live", "found in", "inhabit", "forest", "woodland", "grassland", "wetland", "urban", "garden", "tree", "desert", "mountain"},
		"Fun Facts": {"unique", "special", "only", "fastest", "largest", "smallest", "amazing", "interesting", "can", "able to"},
	}
	
	for category, keywords := range categories {
		found := false
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				found = true
				break
			}
		}
		fmt.Printf("  - %s content: %v\n", category, found)
	}
}

func analyzeCurrentExtraction(birdName string) {
	fmt.Println("Current extraction looks for:")
	fmt.Println("  1. Scientific intro (taxonomy)")
	fmt.Println("  2. Physical description")
	fmt.Println("  3. Habitat and behavior")
	fmt.Println("  4. Diet and feeding")
	fmt.Println("  5. Conservation status")
	fmt.Println("  6. Fun facts")
	
	fmt.Println("\nPotential missing content:")
	fmt.Println("  - Sounds/vocalizations description")
	fmt.Println("  - Nesting behavior and eggs")
	fmt.Println("  - Baby birds/chicks information")
	fmt.Println("  - Migration patterns")
	fmt.Println("  - Interaction with humans")
	fmt.Println("  - Cultural significance")
	fmt.Println("  - Comparison with similar birds")
	fmt.Println("  - Seasonal changes")
	fmt.Println("  - Record-breaking facts (speed, distance, etc.)")
}

func main() {
	fmt.Println("BIRD FACT EXTRACTION - COMPREHENSIVE TEST")
	fmt.Println("==========================================")
	
	birds := []string{
		"American Robin",
		"Northern Cardinal", 
		"Blue Jay",
		"Ruby-throated Hummingbird",
		"Great Horned Owl",
		"Pileated Woodpecker",
		"Bald Eagle",
		"House Sparrow",
	}
	
	for _, bird := range birds {
		TestBirdFacts(bird)
		fmt.Println("\n\nPress Enter for next bird...")
		fmt.Scanln()
	}
	
	fmt.Println("\n==========================================")
	fmt.Println("RECOMMENDATIONS FOR IMPROVEMENT")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("1. Use BOTH Simple and English Wikipedia")
	fmt.Println("   - Simple for kid-friendly base")
	fmt.Println("   - English for detailed facts")
	fmt.Println()
	fmt.Println("2. Extract more engaging content:")
	fmt.Println("   - Record-breaking facts (fastest, highest, etc.)")
	fmt.Println("   - Unique abilities")
	fmt.Println("   - Baby bird information")
	fmt.Println("   - Interaction with humans")
	fmt.Println()
	fmt.Println("3. Add sensory descriptions:")
	fmt.Println("   - What their song sounds like")
	fmt.Println("   - How they move")
	fmt.Println("   - What they look like in flight")
	fmt.Println()
	fmt.Println("4. Include seasonal information:")
	fmt.Println("   - When to see them")
	fmt.Println("   - Migration timing")
	fmt.Println("   - Breeding season")
}