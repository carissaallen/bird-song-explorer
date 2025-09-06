package inaturalist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type TaxonResult struct {
	Results []Taxon `json:"results"`
}

type Taxon struct {
	ID                  int                 `json:"id"`
	Name                string              `json:"name"`
	PreferredCommonName string              `json:"preferred_common_name"`
	Rank                string              `json:"rank"`
	WikipediaURL        string              `json:"wikipedia_url"`
	ConservationStatus  *ConservationStatus `json:"conservation_status"`
	DefaultPhoto        *Photo              `json:"default_photo"`
	TaxonPhotos         []TaxonPhoto        `json:"taxon_photos"`
}

type ConservationStatus struct {
	Status     string `json:"status"`
	Authority  string `json:"authority"`
	StatusName string `json:"status_name"`
}

type Photo struct {
	URL         string `json:"url"`
	Attribution string `json:"attribution"`
	LicenseCode string `json:"license_code"`
	MediumURL   string `json:"medium_url"`
	SquareURL   string `json:"square_url"`
}

type TaxonPhoto struct {
	Photo Photo `json:"photo"`
}

type ObservationSearch struct {
	Results []Observation `json:"results"`
}

type Observation struct {
	ID          int     `json:"id"`
	PlaceGuess  string  `json:"place_guess"`
	ObservedOn  string  `json:"observed_on"`
	Description string  `json:"description"`
	Taxon       *Taxon  `json:"taxon"`
	Photos      []Photo `json:"photos"`
	Sounds      []Sound `json:"sounds"`
}

type Sound struct {
	ID          int    `json:"id"`
	URL         string `json:"file_url"`
	Attribution string `json:"attribution"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.inaturalist.org/v1",
	}
}

// SearchTaxon searches for a bird species in iNaturalist
func (c *Client) SearchTaxon(birdName string) (*Taxon, error) {
	// URL encode the bird name
	encodedName := url.QueryEscape(birdName)

	// Search for the taxon, specifically birds (Aves)
	apiURL := fmt.Sprintf("%s/taxa?q=%s&iconic_taxa=Aves&per_page=1", c.baseURL, encodedName)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "BirdSongExplorer/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch iNaturalist data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("iNaturalist API returned status %d", resp.StatusCode)
	}

	var result TaxonResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no results found for %s", birdName)
	}

	return &result.Results[0], nil
}

// GetRecentObservations gets recent observations of a bird species
func (c *Client) GetRecentObservations(taxonID int, lat, lng float64) ([]Observation, error) {
	// Search for recent observations near the location
	apiURL := fmt.Sprintf("%s/observations?taxon_id=%d&lat=%f&lng=%f&radius=50&order_by=observed_on&order=desc&per_page=5&photos=true",
		c.baseURL, taxonID, lat, lng)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "BirdSongExplorer/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch observations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("iNaturalist API returned status %d", resp.StatusCode)
	}

	var result ObservationSearch
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Results, nil
}

// FormatForKids creates kid-friendly facts from iNaturalist data
func (c *Client) FormatForKids(taxon *Taxon, observations []Observation) []string {
	var facts []string

	// Conservation status fact (more detailed)
	if taxon.ConservationStatus != nil {
		status := taxon.ConservationStatus.StatusName
		if status != "" {
			switch taxon.ConservationStatus.Status {
			case "LC":
				facts = append(facts, "Good news! This bird is doing well and there are lots of them in nature. Scientists call this 'Least Concern' which means we don't need to worry about them disappearing.")
			case "NT":
				facts = append(facts, "Scientists are keeping a close eye on this bird to make sure it stays safe. It's called 'Near Threatened' which means we need to be careful to protect their homes.")
			case "VU":
				facts = append(facts, "This bird needs our help! It's called 'Vulnerable' because there aren't as many as there used to be. We can help by protecting the places where they live.")
			case "EN":
				facts = append(facts, "This is a very special and endangered bird that needs protection. You're incredibly lucky if you see one! Scientists are working hard to help save them.")
			case "CR":
				facts = append(facts, "This is one of the rarest birds in the world! It's 'Critically Endangered' which means every sighting is precious and important for scientists.")
			default:
				facts = append(facts, "Scientists study how many of these birds are in the wild to make sure they stay healthy and safe.")
			}
		}
	}

	// Recent sighting facts with more detail
	if len(observations) > 0 {
		locations := make([]string, 0)
		recentDates := make([]string, 0)
		hasPhotos := false
		hasSounds := false

		for i, obs := range observations {
			if obs.PlaceGuess != "" && len(locations) < 4 { // Increased from 3
				// Simplify location names for kids
				place := obs.PlaceGuess
				if strings.Contains(place, ",") {
					parts := strings.Split(place, ",")
					place = strings.TrimSpace(parts[0])
				}
				// Avoid duplicate locations
				isDuplicate := false
				for _, loc := range locations {
					if loc == place {
						isDuplicate = true
						break
					}
				}
				if !isDuplicate {
					locations = append(locations, place)
				}
			}

			// Check for recent dates
			if i < 2 && obs.ObservedOn != "" {
				recentDates = append(recentDates, obs.ObservedOn)
			}

			// Check for media
			if len(obs.Photos) > 0 {
				hasPhotos = true
			}
			if len(obs.Sounds) > 0 {
				hasSounds = true
			}
		}

		// Location-based facts - DISABLED
		// We handle location awareness properly in Track 4 with actual eBird sightings
		// This prevents generic/incorrect location claims
		/*
		if len(locations) > 0 {
			if len(locations) == 1 {
				facts = append(facts, fmt.Sprintf("Someone recently spotted this bird near %s! Bird watchers love to record where they see different birds.", locations[0]))
			} else if len(locations) == 2 {
				facts = append(facts, fmt.Sprintf("People have recently seen this bird in places like %s and %s. Isn't it amazing how birds can live in different neighborhoods?", locations[0], locations[1]))
			} else {
				facts = append(facts, fmt.Sprintf("Bird watchers have spotted this bird in many places nearby, including %s! These birds might even live in your neighborhood.", strings.Join(locations[:2], " and ")))
			}
		}
		*/

		// Media observation fact
		if hasPhotos && hasSounds {
			facts = append(facts, "Nature lovers have shared both photos and recordings of this bird, helping us learn more about how they look and sound.")
		} else if hasPhotos {
			facts = append(facts, "People love taking pictures of this bird! Each photo helps scientists understand more about where these birds live and what they do.")
		}

		// Timing fact
		if len(recentDates) > 0 {
			facts = append(facts, "This bird has been seen in your area recently, so keep your eyes and ears open - you might spot one too!")
		}
	}

	// Citizen science fact (educational)
	if len(facts) < 4 { // Add if we need more content
		facts = append(facts, "Did you know that when you watch birds and tell others what you see, you're being a citizen scientist? You're helping real scientists learn about birds!")
	}

	// Add a fun engagement fact if we still need content
	if len(facts) < 5 {
		facts = append(facts, "Try to listen carefully for this bird's special song. Each type of bird has its own unique way of singing!")
	}

	return facts
}
