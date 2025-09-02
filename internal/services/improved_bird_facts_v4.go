package services

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/pkg/ebird"
	"github.com/callen/bird-song-explorer/pkg/inaturalist"
	"github.com/callen/bird-song-explorer/pkg/wikipedia"
)

// ImprovedFactGeneratorV4 generates bird facts with location-specific sightings
type ImprovedFactGeneratorV4 struct {
	wikiClient  *wikipedia.Client
	inatClient  *inaturalist.Client
	ebirdClient *ebird.Client
	rng         *rand.Rand
}

// LocationContext holds location-specific information for the script
type LocationContext struct {
	CityName         string
	StateName        string
	NearbyLandmarks  []string
	RecentSightings  []RecentSighting
	SeasonalPresence string  // "year-round", "summer", "winter", "migration"
	Distance         float64 // Distance to nearest sighting in miles
}

// RecentSighting represents a recent bird observation
type RecentSighting struct {
	LocationName string
	Date         string
	Count        int
	DaysAgo      int
}

// NewImprovedFactGeneratorV4 creates a new fact generator with location awareness
func NewImprovedFactGeneratorV4(ebirdAPIKey string) *ImprovedFactGeneratorV4 {
	return &ImprovedFactGeneratorV4{
		wikiClient:  wikipedia.NewClient(),
		inatClient:  inaturalist.NewClient(),
		ebirdClient: ebird.NewClient(ebirdAPIKey),
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateExplorersGuideScriptWithLocation creates a location-aware script
func (fg *ImprovedFactGeneratorV4) GenerateExplorersGuideScriptWithLocation(bird *models.Bird, lat, lng float64) string {
	sections := []string{}
	usedTransitions := make(map[string]bool)

	// Get Wikipedia data
	wikiData, _ := fg.wikiClient.GetBirdSummary(bird.CommonName)

	// Get location context from eBird
	locationContext := fg.getLocationContext(bird, lat, lng)

	// 1. Scientific Introduction
	scientificIntro := fg.generateScientificIntro(bird)
	if scientificIntro != "" {
		sections = append(sections, scientificIntro)
	}

	// 2. Location-specific introduction (NEW)
	locationIntro := fg.generateLocationIntro(bird, locationContext)
	if locationIntro != "" {
		sections = append(sections, locationIntro)
	}

	// 3. Physical Description
	physicalDesc := fg.generateEnhancedPhysicalDescription(bird, wikiData)
	if physicalDesc != "" {
		transition := fg.getTransition(0, usedTransitions) // TransitionFact
		if transition != "" {
			sections = append(sections, transition+" "+physicalDesc)
		} else {
			sections = append(sections, physicalDesc)
		}
	}

	// 4. Vocalizations
	vocalDesc := fg.generateVocalizationDescription(bird, wikiData)
	if vocalDesc != "" {
		sections = append(sections, vocalDesc)
	}

	// 5. Local habitat and behavior (ENHANCED)
	habitat := fg.generateLocalHabitatBehavior(bird, wikiData, locationContext)
	if habitat != "" {
		transition := fg.getTransition(1, usedTransitions) // TransitionAction
		sections = append(sections, transition+" "+habitat)
	}

	// 6. Diet and Feeding
	diet := fg.generateEnhancedDietInfo(bird, wikiData)
	if diet != "" {
		sections = append(sections, diet)
	}

	// 7. Nesting
	nesting := fg.generateNestingInfo(bird, wikiData)
	if nesting != "" {
		transition := fg.getTransition(0, usedTransitions) // TransitionFact
		sections = append(sections, transition+" "+nesting)
	}

	// 8. Amazing Abilities
	abilities := fg.generateAmazingAbilities(bird, wikiData)
	if abilities != "" {
		sections = append(sections, abilities)
	}

	// 9. Recent local sightings (NEW)
	sightings := fg.generateRecentSightingsInfo(bird, locationContext)
	if sightings != "" {
		sections = append(sections, sightings)
	}

	// 10. Conservation with local action
	conservation := fg.generateLocalConservationInfo(bird, locationContext)
	if conservation != "" {
		sections = append(sections, conservation)
	}

	// 11. Fun Facts
	funFacts := fg.generateEnhancedFunFacts(bird, wikiData)
	if funFacts != "" {
		sections = append(sections, funFacts)
	}

	// Join sections with natural flow
	return fg.joinSectionsNaturally(sections, bird.CommonName, locationContext)
}

// getLocationContext fetches location-specific information from eBird
func (fg *ImprovedFactGeneratorV4) getLocationContext(bird *models.Bird, lat, lng float64) LocationContext {
	context := LocationContext{
		CityName:  fg.getCityFromCoordinates(lat, lng),
		StateName: fg.getStateFromCoordinates(lat, lng),
	}

	// Get recent observations from eBird (last 30 days)
	observations, err := fg.ebirdClient.GetRecentObservations(lat, lng, 30)
	if err == nil {
		// Filter for this specific bird
		for _, obs := range observations {
			if strings.EqualFold(obs.CommonName, bird.CommonName) ||
				strings.EqualFold(obs.ScientificName, bird.ScientificName) {

				obsDate, _ := time.Parse("2006-01-02", obs.ObsDate)
				daysAgo := int(time.Since(obsDate).Hours() / 24)

				sighting := RecentSighting{
					LocationName: obs.LocationName,
					Date:         obs.ObsDate,
					Count:        obs.HowMany,
					DaysAgo:      daysAgo,
				}

				context.RecentSightings = append(context.RecentSightings, sighting)

				// Calculate distance to nearest sighting
				if context.Distance == 0 || context.Distance > fg.calculateDistance(lat, lng, obs.Latitude, obs.Longitude) {
					context.Distance = fg.calculateDistance(lat, lng, obs.Latitude, obs.Longitude)
				}
			}
		}

		// Determine seasonal presence based on observations
		context.SeasonalPresence = fg.determineSeasonalPresence(context.RecentSightings)
	}

	return context
}

// generateLocationIntro creates a location-specific introduction
func (fg *ImprovedFactGeneratorV4) generateLocationIntro(bird *models.Bird, context LocationContext) string {
	if len(context.RecentSightings) == 0 {
		return ""
	}

	mostRecent := context.RecentSightings[0]

	// Create kid-friendly location introductions without confusing street details
	intros := []string{
		fmt.Sprintf("Great news! %ss have been spotted near you in %s!", bird.CommonName, context.CityName),
		fmt.Sprintf("You're in luck! A %s was seen just %d days ago near you!", bird.CommonName, mostRecent.DaysAgo),
		fmt.Sprintf("Exciting! %ss are active in your area!", bird.CommonName),
		fmt.Sprintf("Perfect timing! %ss have been seen %d times near %s this month!", bird.CommonName, len(context.RecentSightings), context.CityName),
	}

	if context.Distance < 5 {
		intros = append(intros, fmt.Sprintf("Wow! A %s was spotted less than %.1f miles from you!", bird.CommonName, context.Distance))
	}

	return intros[fg.rng.Intn(len(intros))]
}

// generateLocalHabitatBehavior creates habitat info with local context
func (fg *ImprovedFactGeneratorV4) generateLocalHabitatBehavior(bird *models.Bird, wikiData *wikipedia.PageSummary, context LocationContext) string {
	baseHabitat := fg.generateEnhancedHabitatBehavior(bird, wikiData)

	// Add local context
	if len(context.RecentSightings) > 0 {
		localTips := []string{}

		// Analyze where birds have been seen locally
		parkCount := 0
		waterCount := 0
		urbanCount := 0

		for _, sighting := range context.RecentSightings {
			lower := strings.ToLower(sighting.LocationName)
			if strings.Contains(lower, "park") || strings.Contains(lower, "trail") {
				parkCount++
			}
			if strings.Contains(lower, "lake") || strings.Contains(lower, "river") || strings.Contains(lower, "pond") {
				waterCount++
			}
			if strings.Contains(lower, "garden") || strings.Contains(lower, "yard") || strings.Contains(lower, "feeder") {
				urbanCount++
			}
		}

		if parkCount > len(context.RecentSightings)/2 {
			localTips = append(localTips, fmt.Sprintf("In %s, check local parks and nature trails!", context.CityName))
		}
		if waterCount > 0 {
			localTips = append(localTips, "They love areas near water in your region!")
		}
		if urbanCount > 0 {
			localTips = append(localTips, "You might even see them in backyards and gardens nearby!")
		}

		if len(localTips) > 0 {
			baseHabitat += " " + localTips[fg.rng.Intn(len(localTips))]
		}
	}

	// Add seasonal context
	if context.SeasonalPresence != "" {
		switch context.SeasonalPresence {
		case "year-round":
			baseHabitat += fmt.Sprintf(" %ss live in %s all year long!", bird.CommonName, context.StateName)
		case "summer":
			baseHabitat += fmt.Sprintf(" They visit %s for breeding in summer!", context.StateName)
		case "winter":
			baseHabitat += fmt.Sprintf(" They spend winters in %s escaping the cold!", context.StateName)
		case "migration":
			baseHabitat += fmt.Sprintf(" They pass through %s during migration!", context.StateName)
		}
	}

	return baseHabitat
}

// generateRecentSightingsInfo creates exciting info about recent local sightings
func (fg *ImprovedFactGeneratorV4) generateRecentSightingsInfo(bird *models.Bird, context LocationContext) string {
	if len(context.RecentSightings) == 0 {
		return ""
	}

	// Group sightings by how recent
	thisWeek := 0
	thisMonth := 0

	for _, sighting := range context.RecentSightings {
		if sighting.DaysAgo <= 7 {
			thisWeek++
		}
		if sighting.DaysAgo <= 30 {
			thisMonth++
		}
	}

	sightingPhrases := []string{}

	if thisWeek > 0 {
		sightingPhrases = append(sightingPhrases,
			fmt.Sprintf("Bird watchers saw %d %ss near you just this week!", thisWeek, bird.CommonName))
	}

	if len(context.RecentSightings) > 5 {
		sightingPhrases = append(sightingPhrases,
			fmt.Sprintf("Wow! %ss have been spotted %d times in your area this month!", bird.CommonName, thisMonth))
	}

	// Mention group sightings without confusing location details
	for _, sighting := range context.RecentSightings[:min(3, len(context.RecentSightings))] {
		if sighting.Count > 1 {
			sightingPhrases = append(sightingPhrases,
				fmt.Sprintf("Someone saw %d %ss together in your neighborhood!", sighting.Count, bird.CommonName))
			break
		}
	}

	if len(sightingPhrases) > 0 {
		// Add pauses before and after "Local bird alert!" using spaced periods
		return ". . . Local bird alert! . . . " + sightingPhrases[fg.rng.Intn(len(sightingPhrases))]
	}

	return ""
}

// generateLocalConservationInfo creates conservation info with local actions
func (fg *ImprovedFactGeneratorV4) generateLocalConservationInfo(bird *models.Bird, context LocationContext) string {
	base := fg.generateConservationInfo(bird)

	// Add local conservation actions based on whether we have actual location
	var localActions []string
	
	hasActualState := context.StateName != "your state" && context.StateName != ""
	hasActualCity := context.CityName != "your city" && context.CityName != ""
	
	if hasActualState && hasActualCity {
		// Use specific location names
		localActions = []string{
			fmt.Sprintf("Join the %s Audubon Society to help protect %ss!", context.StateName, bird.CommonName),
			fmt.Sprintf("Report your %s sightings to eBird to help scientists!", bird.CommonName),
			fmt.Sprintf("Participate in the %s Bird Count to track local populations!", context.CityName),
			"Create a bird-friendly yard with native plants and fresh water!",
		}
	} else {
		// Use generic phrasing
		localActions = []string{
			fmt.Sprintf("Join your local Audubon Society to help protect %ss!", bird.CommonName),
			fmt.Sprintf("Report your %s sightings to eBird to help scientists!", bird.CommonName),
			"Participate in your local Bird Count to track populations!",
			"Create a bird-friendly yard with native plants and fresh water!",
		}
	}

	if strings.Contains(base, "help") {
		base += " " + localActions[fg.rng.Intn(len(localActions))]
	}

	return base
}

// Helper functions for location

func (fg *ImprovedFactGeneratorV4) getCityFromCoordinates(lat, lng float64) string {
	// If coordinates are zero, return generic text
	if lat == 0 && lng == 0 {
		return "your city"
	}
	
	// Use reverse geocoding to get actual city name
	cityName := fg.reverseGeocode(lat, lng, "city")
	if cityName != "" {
		return cityName
	}
	
	// Fallback to generic
	return "your city"
}

func (fg *ImprovedFactGeneratorV4) getStateFromCoordinates(lat, lng float64) string {
	// If coordinates are zero, return generic text
	if lat == 0 && lng == 0 {
		return "your state"
	}
	
	// Use reverse geocoding to get actual state name
	stateName := fg.reverseGeocode(lat, lng, "state")
	if stateName != "" {
		return stateName
	}
	
	// Fallback to generic
	return "your state"
}

func (fg *ImprovedFactGeneratorV4) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Haversine formula for distance between two points
	const earthRadius = 3959.0 // miles

	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func (fg *ImprovedFactGeneratorV4) reverseGeocode(lat, lng float64, locationType string) string {
	// Use eBird hotspots to get location names
	// This is a simplified approach using nearby eBird hotspot names
	hotspots, err := fg.ebirdClient.GetNearbyHotspots(lat, lng, 25)
	if err != nil || len(hotspots) == 0 {
		return ""
	}
	
	// Collect city and state names, filtering out street-level details
	cities := make(map[string]int)
	states := make(map[string]int)
	
	for _, hotspot := range hotspots {
		parts := strings.Split(hotspot.LocationName, ", ")
		if len(parts) >= 2 {
			// Last part is usually state (or country)
			statePart := strings.TrimSpace(parts[len(parts)-1])
			// Clean the state name to remove dates/years
			statePart = cleanLocationName(statePart)
			
			// Filter out common state abbreviations and full names
			if statePart != "" && isValidState(statePart) {
				states[statePart]++
			}
			
			// Look for city name - usually second to last, but skip if it contains street indicators
			if len(parts) >= 2 {
				cityPart := strings.TrimSpace(parts[len(parts)-2])
				// Clean the city name to remove dates/years
				cityPart = cleanLocationName(cityPart)
				
				// Skip if it looks like a street address
				if cityPart != "" && !containsStreetIndicators(cityPart) && !containsNumbers(cityPart) {
					cities[cityPart]++
				} else if len(parts) >= 3 {
					// Try third to last if second to last was a street
					cityPart = strings.TrimSpace(parts[len(parts)-3])
					// Clean the city name to remove dates/years
					cityPart = cleanLocationName(cityPart)
					if cityPart != "" && !containsStreetIndicators(cityPart) && !containsNumbers(cityPart) {
						cities[cityPart]++
					}
				}
			}
		}
	}
	
	// Return the most common city or state
	if locationType == "state" {
		return getMostCommon(states)
	} else if locationType == "city" {
		return getMostCommon(cities)
	}
	
	return ""
}

// Helper function to check if string contains street indicators
func containsStreetIndicators(s string) bool {
	streetWords := []string{"St", "Street", "Ave", "Avenue", "Rd", "Road", "Blvd", 
		"Boulevard", "Dr", "Drive", "Ln", "Lane", "Way", "Ct", "Court", 
		"Pl", "Place", "Block", "Park", "Trail", "Path"}
	lower := strings.ToLower(s)
	for _, word := range streetWords {
		if strings.Contains(lower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// Helper function to check if string contains numbers (likely address)
func containsNumbers(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// Helper function to clean location names by removing dates/years
func cleanLocationName(s string) string {
	// Remove common year patterns (4-digit numbers)
	words := strings.Fields(s)
	var cleaned []string
	
	for _, word := range words {
		// Skip if it's a 4-digit year (1900-2099)
		if len(word) == 4 {
			if year, err := strconv.Atoi(word); err == nil && year >= 1900 && year <= 2099 {
				continue
			}
		}
		
		// Skip if it contains date patterns like "2023-01-15" or "01/15/2023"
		if strings.Contains(word, "-") || strings.Contains(word, "/") {
			hasOnlyNumbersAndSeparators := true
			for _, r := range word {
				if !((r >= '0' && r <= '9') || r == '-' || r == '/') {
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

// Helper function to validate state names/abbreviations
func isValidState(s string) bool {
	// Common US state abbreviations and some full names
	// This is a simplified check - could be expanded
	if len(s) == 2 {
		// Likely a state abbreviation
		return true
	}
	
	// Check for common state name patterns
	stateNames := []string{"California", "Oregon", "Washington", "Texas", "Florida", 
		"New York", "Colorado", "Arizona", "Nevada", "Utah", "Idaho", "Montana"}
	
	for _, state := range stateNames {
		if strings.EqualFold(s, state) {
			return true
		}
	}
	
	return len(s) > 2 && len(s) < 20 // Reasonable length for a state name
}

// Helper function to get most common item from map
func getMostCommon(items map[string]int) string {
	if len(items) == 0 {
		return ""
	}
	
	var mostCommon string
	maxCount := 0
	
	for item, count := range items {
		if count > maxCount {
			maxCount = count
			mostCommon = item
		}
	}
	
	// Clean the final result one more time to ensure no dates/years
	return cleanLocationName(mostCommon)
}

func (fg *ImprovedFactGeneratorV4) determineSeasonalPresence(sightings []RecentSighting) string {
	if len(sightings) == 0 {
		return ""
	}

	// Analyze sighting patterns
	// This is simplified - a real implementation would look at historical data
	currentMonth := time.Now().Month()

	if len(sightings) > 10 {
		return "year-round"
	}

	switch currentMonth {
	case time.June, time.July, time.August:
		return "summer"
	case time.December, time.January, time.February:
		return "winter"
	case time.March, time.April, time.September, time.October:
		return "migration"
	default:
		return ""
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// joinSectionsNaturally combines sections with location-aware closing
func (fg *ImprovedFactGeneratorV4) joinSectionsNaturally(sections []string, birdName string, context LocationContext) string {
	if len(sections) == 0 {
		return fmt.Sprintf("The %s is an incredible bird! Listen for its unique song and watch for it in %s.",
			birdName, context.CityName)
	}

	result := strings.Join(sections, " ")
	result = strings.ReplaceAll(result, "  ", " ")

	// Location-aware closings with proper grammar for actual vs generic locations
	var closings []string
	
	// Check if we have actual location names or generic ones
	hasActualLocation := context.CityName != "your city" && context.CityName != ""
	
	if hasActualLocation {
		// Use specific location names when available
		closings = []string{
			fmt.Sprintf(" Now you're a %s expert! Look for one in %s today!", birdName, context.CityName),
			fmt.Sprintf(" Keep watching for %ss around %s - you might be the next to spot one!", birdName, context.CityName),
			fmt.Sprintf(" The %s is waiting to be discovered in %s! Happy bird watching!", birdName, context.CityName),
			fmt.Sprintf(" Now go explore %s and find a %s!", context.CityName, birdName),
		}
	} else {
		// Use generic phrasing when location is unknown
		closings = []string{
			fmt.Sprintf(" Now you're a %s expert! Look for one in your city today!", birdName),
			fmt.Sprintf(" Keep watching for %ss in your neighborhood - you might be the next to spot one!", birdName),
			fmt.Sprintf(" The %s is waiting to be discovered near you! Happy bird watching!", birdName),
			fmt.Sprintf(" Now go explore your area and find a %s!", birdName),
		}
	}

	if len(context.RecentSightings) > 0 {
		closings = append(closings,
			fmt.Sprintf(" With %d recent sightings near you, today might be your lucky day to see a %s!",
				len(context.RecentSightings), birdName))
	}

	if len(result) < 1500 {
		result += closings[fg.rng.Intn(len(closings))]
	}

	return result
}

// Include all the helper functions from V3 (transitions, physical description, etc.)
// These remain the same but I'm including key ones for completeness

func (fg *ImprovedFactGeneratorV4) getTransition(transType int, usedTransitions map[string]bool) string {
	transitions := map[int][]string{
		0: { // TransitionFact
			"Here's an amazing fact:",
			"Did you know?",
			"Fun fact:",
			"Here's something cool:",
			"Guess what?",
			"Want to know something special?",
			"Check this out:",
		},
		1: { // TransitionAction
			"Listen carefully!",
			"Watch for this:",
			"Look closely!",
			"Keep your eyes open:",
			"Pay attention to this:",
		},
	}

	options := transitions[transType]
	for attempts := 0; attempts < 10; attempts++ {
		choice := options[fg.rng.Intn(len(options))]
		if !usedTransitions[choice] {
			usedTransitions[choice] = true
			return choice
		}
	}

	return options[fg.rng.Intn(len(options))]
}

// Include other essential methods from V3
func (fg *ImprovedFactGeneratorV4) generateScientificIntro(bird *models.Bird) string {
	var intro string

	// Only include scientific name if it's available
	if bird.ScientificName != "" {
		intros := []string{
			fmt.Sprintf("Let me tell you about the amazing %s! Its scientific name is %s.", bird.CommonName, bird.ScientificName),
			fmt.Sprintf("Today we're learning about the %s! Scientists call it %s.", bird.CommonName, bird.ScientificName),
			fmt.Sprintf("Get ready to discover the %s! Its scientific name is %s.", bird.CommonName, bird.ScientificName),
		}
		intro = intros[fg.rng.Intn(len(intros))]
	} else {
		// If no scientific name, use a simpler introduction
		intros := []string{
			fmt.Sprintf("Let me tell you about the amazing %s!", bird.CommonName),
			fmt.Sprintf("Today we're learning about the %s!", bird.CommonName),
			fmt.Sprintf("Get ready to discover the %s!", bird.CommonName),
		}
		intro = intros[fg.rng.Intn(len(intros))]
	}

	if bird.Family != "" {
		familyName := bird.Family
		if strings.HasSuffix(familyName, "idae") {
			familyName = strings.TrimSuffix(familyName, "idae")
			intro += fmt.Sprintf(" It belongs to the %s family of birds.", familyName)
		}
	}

	return intro
}

// Copy remaining methods from V3...
// (generateEnhancedPhysicalDescription, generateVocalizationDescription, etc.)
// These would be identical to V3 implementation

func (fg *ImprovedFactGeneratorV4) generateEnhancedPhysicalDescription(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	// Same as V3
	if wikiData == nil {
		return fmt.Sprintf("The %s has unique markings and colors that make it special.", bird.CommonName)
	}

	combinedText := wikiData.Extract
	sentences := strings.Split(combinedText, ". ")

	var physicalFacts []string
	usedSentences := make(map[string]bool)

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		lower := strings.ToLower(sentence)

		if usedSentences[lower] {
			continue
		}

		if (strings.Contains(lower, "color") || strings.Contains(lower, "size") ||
			strings.Contains(lower, "wing") || strings.Contains(lower, "marking")) &&
			!strings.Contains(lower, "genus") && len(sentence) < 200 {

			physicalFacts = append(physicalFacts, sentence)
			usedSentences[lower] = true

			if len(physicalFacts) >= 2 {
				break
			}
		}
	}

	if len(physicalFacts) > 0 {
		return strings.Join(physicalFacts, " ")
	}

	return fmt.Sprintf("The %s has unique markings and colors that make it special.", bird.CommonName)
}

func (fg *ImprovedFactGeneratorV4) generateVocalizationDescription(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	// Same implementation as V3
	lowerName := strings.ToLower(bird.CommonName)
	soundIntros := []string{
		"Listen for their sound! ",
		"Their voice is special! ",
		"You can identify them by their call! ",
	}

	intro := soundIntros[fg.rng.Intn(len(soundIntros))]

	if strings.Contains(lowerName, "robin") {
		return intro + "Robins sing a cheerful melody that sounds like 'cheerily, cheer-up, cheerio!'"
	} else if strings.Contains(lowerName, "cardinal") {
		return intro + "Cardinals whistle clear notes like 'birdy-birdy-birdy' or 'cheer-cheer-cheer.'"
	}

	return ""
}

func (fg *ImprovedFactGeneratorV4) generateEnhancedHabitatBehavior(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	// Basic implementation - enhanced version uses generateLocalHabitatBehavior
	return fmt.Sprintf("You might spot %ss in parks, gardens, or natural areas.", bird.CommonName)
}

func (fg *ImprovedFactGeneratorV4) generateEnhancedDietInfo(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	// Same as V3
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "hummingbird") {
		return "Watch them feed! They hover at flowers, sipping nectar and catching tiny insects!"
	}
	return "Notice how they search for food - hopping, pecking, and exploring!"
}

func (fg *ImprovedFactGeneratorV4) generateNestingInfo(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	// Same as V3
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "robin") {
		return "Robin parents lay 3-5 blue eggs. Tiny pink babies hatch after two weeks!"
	}
	return ""
}

func (fg *ImprovedFactGeneratorV4) generateAmazingAbilities(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	// Same as V3
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "hummingbird") {
		return "Incredible ability: They can fly backwards and hover! Hearts beat 1,200 times per minute!"
	}
	return ""
}

func (fg *ImprovedFactGeneratorV4) generateConservationInfo(bird *models.Bird) string {
	// Basic version - enhanced version uses generateLocalConservationInfo
	return fmt.Sprintf("You can help %ss by providing bird feeders and keeping cats indoors!", bird.CommonName)
}

func (fg *ImprovedFactGeneratorV4) generateEnhancedFunFacts(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	// Same as V3
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "cardinal") {
		return "Fun fact: Cardinals are the state bird of seven states!"
	}
	return "Keep watching - every bird has its own special story!"
}

// EstimateReadingTime calculates approximate speech duration
func (fg *ImprovedFactGeneratorV4) EstimateReadingTime(text string) int {
	words := len(strings.Fields(text))
	return int(math.Ceil(float64(words) / 150.0 * 60))
}
