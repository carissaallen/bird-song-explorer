package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func fetchWikipediaData(birdName string) {
	fmt.Printf("\n========== WIKIPEDIA (English) ==========\n")
	fmt.Printf("Bird: %s\n", birdName)
	fmt.Println("-----------------------------------------")
	
	// Try lowercase first (works better)
	lowerName := strings.ToLower(birdName)
	encodedName := url.QueryEscape(lowerName)
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
	
	var result struct {
		Query struct {
			Pages []struct {
				Title   string `json:"title"`
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}
	
	if len(result.Query.Pages) > 0 && result.Query.Pages[0].Extract != "" {
		page := result.Query.Pages[0]
		fmt.Printf("Title: %s\n", page.Title)
		fmt.Printf("Full Extract (%d characters):\n", len(page.Extract))
		fmt.Println("---")
		fmt.Println(page.Extract)
	} else {
		fmt.Println("No extract found in Wikipedia")
	}
}

func fetchSimpleWikipediaData(birdName string) {
	fmt.Printf("\n========== SIMPLE WIKIPEDIA ==========\n")
	fmt.Printf("Bird: %s\n", birdName)
	fmt.Println("---------------------------------------")
	
	encodedName := url.QueryEscape(strings.ReplaceAll(strings.ToLower(birdName), " ", "_"))
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
	fmt.Printf("Full Extract:\n")
	fmt.Println("---")
	fmt.Println(result.Extract)
}

func fetchINaturalistData(birdName string) {
	fmt.Printf("\n========== iNATURALIST ==========\n")
	fmt.Printf("Bird: %s\n", birdName)
	fmt.Println("----------------------------------")
	
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
	
	var result struct {
		Results []struct {
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			PreferredCommonName string `json:"preferred_common_name"`
			WikipediaSummary    string `json:"wikipedia_summary"`
			WikipediaURL        string `json:"wikipedia_url"`
			ConservationStatus  *struct {
				Status      string `json:"status"`
				StatusName  string `json:"status_name"`
				Authority   string `json:"authority"`
				IUCNStatus  string `json:"iucn"`
				Description string `json:"description"`
			} `json:"conservation_status"`
			EstablishmentMeans struct {
				EstablishmentMeans string `json:"establishment_means"`
			} `json:"establishment_means"`
			TaxonNames []struct {
				Name       string `json:"name"`
				Locale     string `json:"locale"`
				Vernacular bool   `json:"is_vernacular"`
			} `json:"taxon_names"`
			ObservationsCount int `json:"observations_count"`
			MinSpeciesAncestors []struct {
				Name       string `json:"name"`
				CommonName string `json:"common_name"`
				Rank       string `json:"rank"`
			} `json:"min_species_ancestors"`
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
	fmt.Printf("Taxon ID: %d\n", taxon.ID)
	fmt.Printf("Total Observations: %d\n", taxon.ObservationsCount)
	
	if taxon.WikipediaSummary != "" {
		fmt.Println("\niNaturalist's Wikipedia Summary:")
		fmt.Println("---")
		fmt.Println(taxon.WikipediaSummary)
	}
	
	if taxon.ConservationStatus != nil {
		fmt.Println("\nConservation Status:")
		fmt.Printf("  Status: %s (%s)\n", taxon.ConservationStatus.StatusName, taxon.ConservationStatus.Status)
		fmt.Printf("  Authority: %s\n", taxon.ConservationStatus.Authority)
		if taxon.ConservationStatus.Description != "" {
			fmt.Printf("  Description: %s\n", taxon.ConservationStatus.Description)
		}
	}
	
	fmt.Println("\nTaxonomy Hierarchy:")
	for _, ancestor := range taxon.MinSpeciesAncestors {
		if ancestor.CommonName != "" {
			fmt.Printf("  %s: %s (%s)\n", ancestor.Rank, ancestor.CommonName, ancestor.Name)
		} else {
			fmt.Printf("  %s: %s\n", ancestor.Rank, ancestor.Name)
		}
	}
	
	// Fetch recent observations near NYC (example location)
	fmt.Println("\n--- Recent Observations (NYC area) ---")
	obsURL := fmt.Sprintf("https://api.inaturalist.org/v1/observations?taxon_id=%d&lat=40.7128&lng=-74.0060&radius=50&order_by=observed_on&order=desc&per_page=3", taxon.ID)
	
	req2, _ := http.NewRequest("GET", obsURL, nil)
	req2.Header.Set("User-Agent", "BirdSongExplorer/1.0 Testing")
	
	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("Error fetching observations: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	
	var obsResult struct {
		Results []struct {
			ID          int    `json:"id"`
			PlaceGuess  string `json:"place_guess"`
			ObservedOn  string `json:"observed_on"`
			Description string `json:"description"`
			Notes       string `json:"notes"`
			Quality     string `json:"quality_grade"`
		} `json:"results"`
	}
	
	if err := json.NewDecoder(resp2.Body).Decode(&obsResult); err == nil {
		if len(obsResult.Results) > 0 {
			fmt.Printf("Found %d recent observations:\n", len(obsResult.Results))
			for i, obs := range obsResult.Results {
				fmt.Printf("%d. Location: %s (Date: %s)\n", i+1, obs.PlaceGuess, obs.ObservedOn)
				if obs.Notes != "" {
					fmt.Printf("   Observer notes: %s\n", obs.Notes)
				}
			}
		} else {
			fmt.Println("No recent observations in this area")
		}
	}
}

func main() {
	fmt.Println("=========================================================")
	fmt.Println("Bird Song Explorer - RAW API DATA")
	fmt.Println("=========================================================")
	fmt.Println("This shows the complete, unfiltered data from APIs")
	fmt.Println("so you can see everything available for selection.")
	fmt.Println()
	
	// Test with a few birds
	birds := []string{
		"American Robin",
		"Northern Cardinal",
		"Blue Jay",
	}
	
	for _, bird := range birds {
		fmt.Printf("\n#########################################################\n")
		fmt.Printf("# %s\n", bird)
		fmt.Printf("#########################################################\n")
		
		fetchWikipediaData(bird)
		fetchSimpleWikipediaData(bird)
		fetchINaturalistData(bird)
		
		fmt.Println("\n\nPress Enter to continue to next bird...")
		fmt.Scanln()
	}
	
	fmt.Println("\n=========================================================")
	fmt.Println("END OF RAW DATA")
	fmt.Println("=========================================================")
	fmt.Println("\nAs you can see, the APIs provide:")
	fmt.Println("- Wikipedia: Extensive descriptions, behavior, habitat, diet")
	fmt.Println("- Simple Wikipedia: Shorter, kid-friendly summaries")
	fmt.Println("- iNaturalist: Conservation data, taxonomy, observations")
	fmt.Println("\nOur improved generator selects the best parts from each source.")
}