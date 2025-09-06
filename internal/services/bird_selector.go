package services

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/pkg/ebird"
	"github.com/callen/bird-song-explorer/pkg/inaturalist"
	"github.com/callen/bird-song-explorer/pkg/wikipedia"
	"github.com/callen/bird-song-explorer/pkg/xenocanto"
)

type BirdSelector struct {
	ebirdClient *ebird.Client
	xcClient    *xenocanto.Client
	wikiClient  *wikipedia.Client
	inatClient  *inaturalist.Client
}

func NewBirdSelector(ebirdAPIKey, xenoCantoAPIKey string) *BirdSelector {
	return &BirdSelector{
		ebirdClient: ebird.NewClient(ebirdAPIKey),
		xcClient:    xenocanto.NewClient(xenoCantoAPIKey),
		wikiClient:  wikipedia.NewClient(),
		inatClient:  inaturalist.NewClient(),
	}
}

func (bs *BirdSelector) SelectBirdOfDay(location *models.Location) (*models.Bird, error) {
	observations, err := bs.ebirdClient.GetRecentObservations(
		location.Latitude,
		location.Longitude,
		7,
	)
	if err != nil {
		return bs.getFallbackBird()
	}

	if len(observations) == 0 {
		return bs.getFallbackBird()
	}

	speciesMap := make(map[string]ebird.Observation)
	for _, obs := range observations {
		if _, exists := speciesMap[obs.SpeciesCode]; !exists {
			speciesMap[obs.SpeciesCode] = obs
		}
	}

	var uniqueSpecies []ebird.Observation
	for _, obs := range speciesMap {
		uniqueSpecies = append(uniqueSpecies, obs)
	}

	rand.Seed(time.Now().UnixNano())
	maxAttempts := 5

	for i := 0; i < maxAttempts && len(uniqueSpecies) > 0; i++ {
		idx := rand.Intn(len(uniqueSpecies))
		selected := uniqueSpecies[idx]

		recording, err := bs.xcClient.GetBestRecording(selected.ScientificName)
		if err != nil {
			uniqueSpecies = append(uniqueSpecies[:idx], uniqueSpecies[idx+1:]...)
			continue
		}

		speciesInfo, _ := bs.ebirdClient.GetSpeciesInfo(selected.SpeciesCode)

		bird := &models.Bird{
			CommonName:       selected.CommonName,
			ScientificName:   selected.ScientificName,
			Region:           location.Region,
			AudioURL:         recording.File,
			AudioAttribution: recording.Attribution,
			Latitude:         location.Latitude,
			Longitude:        location.Longitude,
		}

		if speciesInfo != nil {
			bird.Family = speciesInfo.Family
			bird.Order = speciesInfo.Order
		}

		bird.Facts = bs.generateBirdFacts(bird)

		bs.enrichWithWikipedia(bird)

		return bird, nil
	}

	return bs.getFallbackBird()
}

// GetBirdByName retrieves a bird by its common name
func (bs *BirdSelector) GetBirdByName(commonName string) (*models.Bird, error) {
	// Try to get bird details from Wikipedia
	wikiSummary, err := bs.wikiClient.GetBirdSummary(commonName)
	if err != nil {
		// If Wikipedia fails, create a basic bird entry
		return bs.getBirdFromXenoCanto(commonName)
	}

	// Extract scientific name from Wikipedia summary
	scientificName := ""
	if wikiSummary != nil && wikiSummary.Extract != "" {
		// Look for scientific name in parentheses
		if strings.Contains(wikiSummary.Extract, "(") && strings.Contains(wikiSummary.Extract, ")") {
			start := strings.Index(wikiSummary.Extract, "(")
			end := strings.Index(wikiSummary.Extract, ")")
			if start < end && end-start < 50 {
				potentialName := wikiSummary.Extract[start+1 : end]
				words := strings.Fields(potentialName)
				if len(words) == 2 && strings.Title(words[0]) == words[0] {
					scientificName = potentialName
				}
			}
		}
	}

	// Get audio recording
	recording, err := bs.xcClient.GetBestRecording(scientificName)
	if err != nil {
		// Try with common name if scientific name fails
		recording, err = bs.xcClient.GetBestRecording(commonName)
		if err != nil {
			return nil, fmt.Errorf("failed to get recording for %s: %w", commonName, err)
		}
	}

	bird := &models.Bird{
		CommonName:       commonName,
		ScientificName:   scientificName,
		AudioURL:         recording.File,
		AudioAttribution: recording.Attribution,
		Description:      wikiSummary.Extract,
	}

	// Generate basic facts
	bird.Facts = bs.generateBirdFacts(bird)

	return bird, nil
}

// getBirdFromXenoCanto gets bird info using only XenoCanto (fallback)
func (bs *BirdSelector) getBirdFromXenoCanto(commonName string) (*models.Bird, error) {
	recording, err := bs.xcClient.GetBestRecording(commonName)
	if err != nil {
		return nil, fmt.Errorf("failed to get recording for %s: %w", commonName, err)
	}

	bird := &models.Bird{
		CommonName:       commonName,
		AudioURL:         recording.File,
		AudioAttribution: recording.Attribution,
		Description:      fmt.Sprintf("The %s is a fascinating bird with unique characteristics.", commonName),
	}

	// Generate basic facts
	bird.Facts = bs.generateBirdFacts(bird)

	return bird, nil
}

func (bs *BirdSelector) getFallbackBird() (*models.Bird, error) {
	commonBirds := []struct {
		common     string
		scientific string
	}{
		{"American Robin", "Turdus migratorius"},
		{"Northern Cardinal", "Cardinalis cardinalis"},
		{"Blue Jay", "Cyanocitta cristata"},
		{"House Sparrow", "Passer domesticus"},
		{"Mourning Dove", "Zenaida macroura"},
	}

	rand.Seed(time.Now().UnixNano())
	selected := commonBirds[rand.Intn(len(commonBirds))]

	recording, err := bs.xcClient.GetBestRecording(selected.scientific)
	if err != nil {
		return nil, fmt.Errorf("failed to get fallback bird recording: %w", err)
	}

	bird := &models.Bird{
		CommonName:       selected.common,
		ScientificName:   selected.scientific,
		Region:           "North America",
		AudioURL:         recording.File,
		AudioAttribution: recording.Attribution,
		Facts:            bs.generateBirdFacts(&models.Bird{CommonName: selected.common, ScientificName: selected.scientific}),
	}

	bs.enrichWithWikipedia(bird)

	return bird, nil
}

func (bs *BirdSelector) generateBirdFacts(bird *models.Bird) []string {
	facts := []string{
		fmt.Sprintf("The %s's scientific name is %s.", bird.CommonName, bird.ScientificName),
	}

	if bird.Family != "" {
		facts = append(facts, fmt.Sprintf("It belongs to the %s family.", bird.Family))
	}

	baseFactTemplates := []string{
		"This bird can be found in %s.",
		"Listen carefully to hear its distinctive call!",
		"Birds use songs to communicate with each other.",
		"Every bird species has its own unique song pattern.",
		"Bird songs are loudest during the early morning hours.",
	}

	for _, template := range baseFactTemplates {
		if len(facts) >= 5 {
			break
		}
		if template == "This bird can be found in %s." && bird.Region != "" {
			facts = append(facts, fmt.Sprintf(template, bird.Region))
		} else if !strings.Contains(template, "%s") {
			facts = append(facts, template)
		}
	}

	return facts
}

func (bs *BirdSelector) enrichWithWikipedia(bird *models.Bird) {
	var descriptions []string

	// 1. Simple Wikipedia
	summary, err := bs.wikiClient.GetBirdSummary(bird.CommonName)
	if err != nil {
		summary, err = bs.wikiClient.GetBirdSummary(bird.ScientificName)
	}

	if err == nil && summary != nil {
		wikiDesc := bs.wikiClient.FormatForKids(summary, bird.CommonName)
		if wikiDesc != "" {
			descriptions = append(descriptions, wikiDesc)
		}

		if summary.ContentURLs.Desktop.Page != "" {
			bird.WikipediaURL = summary.ContentURLs.Desktop.Page
		}
	}

	// 2. iNaturalist for additional facts
	taxon, err := bs.inatClient.SearchTaxon(bird.CommonName)
	if err != nil {
		// Try with scientific name
		taxon, err = bs.inatClient.SearchTaxon(bird.ScientificName)
	}

	if err == nil && taxon != nil {
		// Get recent observations if we have location data
		var observations []inaturalist.Observation
		if bird.Latitude != 0 && bird.Longitude != 0 {
			observations, _ = bs.inatClient.GetRecentObservations(taxon.ID, bird.Latitude, bird.Longitude)
		}

		// Get kid-friendly facts from iNaturalist
		inatFacts := bs.inatClient.FormatForKids(taxon, observations)
		for _, fact := range inatFacts {
			descriptions = append(descriptions, fact)
		}
	}

	// 3. Add some of the generated bird facts for more content
	if len(bird.Facts) > 0 {
		for i, fact := range bird.Facts {
			if i < 2 && fact != "" {
				descriptions = append(descriptions, fact)
			}
		}
	}

	// Combine descriptions from multiple sources
	if len(descriptions) > 0 {
		combinedDescription := strings.Join(descriptions, " ")
		// Clean up any dates/years from the description
		bird.Description = bs.cleanDescriptionText(combinedDescription)
	} else {
		// Fallback description if no sources have data
		bird.Description = fmt.Sprintf("The %s is an amazing bird that you can hear in your area! Listen carefully to learn its unique song.", bird.CommonName)
	}
}

// cleanDescriptionText removes dates and years from description text
func (bs *BirdSelector) cleanDescriptionText(text string) string {
	// Split into words
	words := strings.Fields(text)
	var cleaned []string

	for _, word := range words {
		// Remove standalone 4-digit years (1900-2099)
		if len(word) == 4 {
			if year, err := strconv.Atoi(strings.Trim(word, ".,!?")); err == nil && year >= 1900 && year <= 2099 {
				continue
			}
		}

		// Skip date patterns like "2023-01-15" or "01/15/2023"
		if strings.Contains(word, "-") || strings.Contains(word, "/") {
			hasOnlyNumbersAndSeparators := true
			for _, r := range word {
				if !((r >= '0' && r <= '9') || r == '-' || r == '/' || r == '.' || r == ',') {
					hasOnlyNumbersAndSeparators = false
					break
				}
			}
			if hasOnlyNumbersAndSeparators {
				continue
			}
		}

		cleaned = append(cleaned, word)
	}

	return strings.Join(cleaned, " ")
}
