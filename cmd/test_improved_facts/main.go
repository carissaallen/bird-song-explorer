package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WikipediaFullPage struct {
	Parse struct {
		Title string `json:"title"`
		Text  struct {
			Content string `json:"*"`
		} `json:"text"`
		Sections []struct {
			Line   string `json:"line"`
			Number string `json:"number"`
			Index  string `json:"index"`
		} `json:"sections"`
	} `json:"parse"`
}

type WikipediaExtract struct {
	Query struct {
		Pages []struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

type INatTaxonDetails struct {
	Results []struct {
		ID                  int    `json:"id"`
		Name                string `json:"name"`
		PreferredCommonName string `json:"preferred_common_name"`
		WikipediaSummary    string `json:"wikipedia_summary"`
		ConservationStatus  *struct {
			Status     string `json:"status"`
			StatusName string `json:"status_name"`
		} `json:"conservation_status"`
		EstablishmentMeans struct {
			EstablishmentMeans string `json:"establishment_means"`
		} `json:"establishment_means"`
		MinSpeciesAncestors []struct {
			Name       string `json:"name"`
			CommonName string `json:"common_name"`
			Rank       string `json:"rank"`
		} `json:"min_species_ancestors"`
	} `json:"results"`
}

func fetchWikipediaExtract(birdName string) {
	fmt.Println("\n=== Wikipedia Extract API (English) ===")
	encodedName := url.QueryEscape(birdName)
	apiURL := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&format=json&prop=extracts&exintro=true&explaintext=true&exlimit=1&titles=%s&formatversion=2", encodedName)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "BirdSongExplorer/1.0 Testing")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	var result WikipediaExtract
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}
	
	if len(result.Query.Pages) > 0 {
		page := result.Query.Pages[0]
		fmt.Printf("Title: %s\n", page.Title)
		fmt.Printf("Extract Length: %d characters\n", len(page.Extract))
		
		// Show first 1500 characters
		if len(page.Extract) > 1500 {
			fmt.Printf("Extract (first 1500 chars):\n%s...\n", page.Extract[:1500])
		} else {
			fmt.Printf("Extract:\n%s\n", page.Extract)
		}
		
		// Extract interesting sentences for kids
		sentences := strings.Split(page.Extract, ". ")
		fmt.Println("\n--- Potential Kid-Friendly Sentences ---")
		count := 0
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if sentence == "" {
				continue
			}
			
			// Look for interesting facts
			lowerSentence := strings.ToLower(sentence)
			isInteresting := false
			
			// Positive indicators for kid-friendly content
			if strings.Contains(lowerSentence, "color") ||
				strings.Contains(lowerSentence, "colour") ||
				strings.Contains(lowerSentence, "size") ||
				strings.Contains(lowerSentence, "length") ||
				strings.Contains(lowerSentence, "wingspan") ||
				strings.Contains(lowerSentence, "eat") ||
				strings.Contains(lowerSentence, "feed") ||
				strings.Contains(lowerSentence, "diet") ||
				strings.Contains(lowerSentence, "nest") ||
				strings.Contains(lowerSentence, "egg") ||
				strings.Contains(lowerSentence, "baby") ||
				strings.Contains(lowerSentence, "chick") ||
				strings.Contains(lowerSentence, "habitat") ||
				strings.Contains(lowerSentence, "live") ||
				strings.Contains(lowerSentence, "found in") ||
				strings.Contains(lowerSentence, "migrate") ||
				strings.Contains(lowerSentence, "winter") ||
				strings.Contains(lowerSentence, "summer") ||
				strings.Contains(lowerSentence, "fly") ||
				strings.Contains(lowerSentence, "song") ||
				strings.Contains(lowerSentence, "call") ||
				strings.Contains(lowerSentence, "sound") {
				isInteresting = true
			}
			
			// Skip overly technical sentences
			if strings.Contains(lowerSentence, "genus") ||
				strings.Contains(lowerSentence, "taxonomy") ||
				strings.Contains(lowerSentence, "subspecies") ||
				strings.Contains(lowerSentence, "phylogen") ||
				strings.Contains(lowerSentence, "described by") ||
				strings.Contains(lowerSentence, "classified") ||
				len(sentence) > 200 {
				isInteresting = false
			}
			
			if isInteresting && count < 10 {
				fmt.Printf("  • %s.\n", sentence)
				count++
			}
		}
	} else {
		fmt.Println("No pages found in Wikipedia response")
	}
}

func fetchSimpleWikipedia(birdName string) {
	fmt.Println("\n=== Simple Wikipedia Summary ===")
	encodedName := url.QueryEscape(strings.ReplaceAll(birdName, " ", "_"))
	apiURL := fmt.Sprintf("https://simple.wikipedia.org/api/rest_v1/page/summary/%s", encodedName)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "BirdSongExplorer/1.0 Testing")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		fmt.Println("Not found on Simple Wikipedia")
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
	fmt.Printf("Extract: %s\n", result.Extract)
}

func fetchINaturalistDetails(birdName string) {
	fmt.Println("\n=== iNaturalist Taxon Details ===")
	encodedName := url.QueryEscape(birdName)
	apiURL := fmt.Sprintf("https://api.inaturalist.org/v1/taxa?q=%s&iconic_taxa=Aves&per_page=1", encodedName)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "BirdSongExplorer/1.0 Testing")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	var result INatTaxonDetails
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
	fmt.Printf("ID: %d\n", taxon.ID)
	
	if taxon.WikipediaSummary != "" {
		fmt.Printf("Wikipedia Summary from iNat:\n%s\n", taxon.WikipediaSummary)
	}
	
	if taxon.ConservationStatus != nil {
		fmt.Printf("Conservation Status: %s (%s)\n", taxon.ConservationStatus.StatusName, taxon.ConservationStatus.Status)
	}
	
	// Show taxonomy
	fmt.Println("\nTaxonomy Hierarchy:")
	for _, ancestor := range taxon.MinSpeciesAncestors {
		if ancestor.CommonName != "" {
			fmt.Printf("  %s: %s (%s)\n", ancestor.Rank, ancestor.CommonName, ancestor.Name)
		} else {
			fmt.Printf("  %s: %s\n", ancestor.Rank, ancestor.Name)
		}
	}
	
	// Fetch species account (more detailed info)
	fmt.Println("\n=== iNaturalist Species Account ===")
	accountURL := fmt.Sprintf("https://api.inaturalist.org/v1/taxa/%d", taxon.ID)
	req2, _ := http.NewRequest("GET", accountURL, nil)
	req2.Header.Set("User-Agent", "BirdSongExplorer/1.0 Testing")
	
	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("Error fetching account: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	
	var accountResult struct {
		Results []struct {
			TaxonNames []struct {
				Name       string `json:"name"`
				Locale     string `json:"locale"`
				Vernacular bool   `json:"is_vernacular"`
			} `json:"taxon_names"`
			Atlas struct {
				Presence bool `json:"presence"`
			} `json:"atlas"`
		} `json:"results"`
	}
	
	if err := json.NewDecoder(resp2.Body).Decode(&accountResult); err == nil && len(accountResult.Results) > 0 {
		// Show common names in different languages
		fmt.Println("\nCommon Names:")
		nameCount := 0
		for _, name := range accountResult.Results[0].TaxonNames {
			if name.Vernacular && nameCount < 5 {
				fmt.Printf("  • %s (%s)\n", name.Name, name.Locale)
				nameCount++
			}
		}
	}
}

func testBird(birdName string) {
	fmt.Printf("\n############################################\n")
	fmt.Printf("Testing: %s\n", birdName)
	fmt.Printf("############################################\n")
	
	// Test all APIs
	fetchWikipediaExtract(birdName)
	fetchSimpleWikipedia(birdName)
	fetchINaturalistDetails(birdName)
	
	// Also test with scientific name if different
	if birdName == "American Robin" {
		fmt.Println("\n--- Testing with Scientific Name ---")
		fetchWikipediaExtract("Turdus migratorius")
	}
}

func main() {
	fmt.Println("Bird Song Explorer - Enhanced Fact Gathering Test")
	fmt.Println("================================================")
	fmt.Println("This test fetches detailed information from Wikipedia and iNaturalist APIs")
	fmt.Println("to see what content is available for the Bird Explorer's Guide track.")
	
	// Test with common birds
	birds := []string{
		"American Robin",
		"Northern Cardinal",
		"Blue Jay",
		"House Sparrow",
		"Red-tailed Hawk",
		"Ruby-throated Hummingbird",
		"Great Horned Owl",
		"Pileated Woodpecker",
	}
	
	for _, bird := range birds {
		testBird(bird)
		// fmt.Println("\n\nPress Enter to continue to next bird...")
		// fmt.Scanln()
	}
	
	fmt.Println("\n=== Summary ===")
	fmt.Println("The APIs provide rich content including:")
	fmt.Println("1. Wikipedia: Detailed descriptions, habitat, behavior, diet")
	fmt.Println("2. Simple Wikipedia: Simplified language better for kids")
	fmt.Println("3. iNaturalist: Conservation status, taxonomy, common names")
	fmt.Println("\nRecommendations for Bird Explorer's Guide:")
	fmt.Println("1. Start with scientific name and family")
	fmt.Println("2. Add 2-3 physical description sentences (size, colors)")
	fmt.Println("3. Include habitat and diet information")
	fmt.Println("4. Add conservation status if notable")
	fmt.Println("5. End with fun fact or behavior note")
}