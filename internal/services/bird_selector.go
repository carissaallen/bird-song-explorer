package services

import (
	"fmt"
	"math/rand"
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
		bird.Description = strings.Join(descriptions, " ")
	} else {
		// Fallback description if no sources have data
		bird.Description = fmt.Sprintf("The %s is an amazing bird that you can hear in your area! Listen carefully to learn its unique song.", bird.CommonName)
	}
}
