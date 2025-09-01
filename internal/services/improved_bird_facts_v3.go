package services

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/pkg/inaturalist"
	"github.com/callen/bird-song-explorer/pkg/wikipedia"
)

// ImprovedFactGeneratorV3 generates comprehensive, engaging bird facts with better transitions
type ImprovedFactGeneratorV3 struct {
	wikiClient *wikipedia.Client
	inatClient *inaturalist.Client
	rng        *rand.Rand
}

// TransitionType represents different types of transitions
type TransitionType int

const (
	TransitionFact TransitionType = iota
	TransitionAction
	TransitionQuestion
	TransitionExcitement
	TransitionContinuation
)

// NewImprovedFactGeneratorV3 creates a new enhanced fact generator with better transitions
func NewImprovedFactGeneratorV3() *ImprovedFactGeneratorV3 {
	return &ImprovedFactGeneratorV3{
		wikiClient: wikipedia.NewClient(),
		inatClient: inaturalist.NewClient(),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// getTransition returns an appropriate transition phrase based on context
func (fg *ImprovedFactGeneratorV3) getTransition(transType TransitionType, usedTransitions map[string]bool) string {
	transitions := map[TransitionType][]string{
		TransitionFact: {
			"Here's an amazing fact:",
			"Did you know?",
			"Fun fact:",
			"Here's something cool:",
			"Guess what?",
			"Want to know something special?",
			"Check this out:",
			"Here's something interesting:",
			"Believe it or not,",
			"Amazingly,",
		},
		TransitionAction: {
			"Listen carefully!",
			"Watch for this:",
			"Look closely!",
			"Keep your eyes open:",
			"Pay attention to this:",
			"Notice how they",
			"You might see them",
			"Try to spot them when they",
			"See if you can observe them",
		},
		TransitionQuestion: {
			"Have you ever wondered",
			"Do you know",
			"Can you imagine",
			"Have you noticed",
			"Ever seen",
		},
		TransitionExcitement: {
			"How incredible!",
			"That's amazing!",
			"Wow!",
			"How cool is that?",
			"Isn't that wonderful?",
			"That's spectacular!",
		},
		TransitionContinuation: {
			"Also,",
			"Plus,",
			"What's more,",
			"In addition,",
			"Another thing:",
			"Not only that, but",
			"Even better,",
			"",  // Sometimes no transition is best
		},
	}
	
	options := transitions[transType]
	// Try to find an unused transition
	for attempts := 0; attempts < 10; attempts++ {
		choice := options[fg.rng.Intn(len(options))]
		if !usedTransitions[choice] {
			usedTransitions[choice] = true
			return choice
		}
	}
	
	// If all have been used, return a random one
	return options[fg.rng.Intn(len(options))]
}

// GenerateExplorersGuideScript creates an engaging script with better transitions
func (fg *ImprovedFactGeneratorV3) GenerateExplorersGuideScript(bird *models.Bird) string {
	sections := []string{}
	usedTransitions := make(map[string]bool)
	
	// Get comprehensive Wikipedia data
	wikiData, _ := fg.wikiClient.GetBirdSummary(bird.CommonName)
	
	// 1. Scientific Introduction (always first, no transition needed)
	scientificIntro := fg.generateScientificIntro(bird)
	if scientificIntro != "" {
		sections = append(sections, scientificIntro)
	}
	
	// 2. Physical Description
	physicalDesc := fg.generateEnhancedPhysicalDescription(bird, wikiData)
	if physicalDesc != "" {
		// Use a fact transition for physical description
		transition := fg.getTransition(TransitionFact, usedTransitions)
		if transition != "" {
			sections = append(sections, transition + " " + physicalDesc)
		} else {
			sections = append(sections, physicalDesc)
		}
	}
	
	// 3. Vocalizations
	vocalDesc := fg.generateVocalizationDescription(bird, wikiData)
	if vocalDesc != "" {
		// Don't add another transition - the vocalization section has its own intro
		sections = append(sections, vocalDesc)
	}
	
	// 4. Habitat and Behavior
	habitat := fg.generateEnhancedHabitatBehavior(bird, wikiData)
	if habitat != "" {
		transition := fg.getTransition(TransitionAction, usedTransitions)
		sections = append(sections, transition + " " + habitat)
	}
	
	// 5. Diet and Feeding
	diet := fg.generateEnhancedDietInfo(bird, wikiData)
	if diet != "" {
		// Diet descriptions often start with action, so minimal transition
		sections = append(sections, diet)
	}
	
	// 6. Nesting and Baby Birds
	nesting := fg.generateNestingInfo(bird, wikiData)
	if nesting != "" {
		transition := fg.getTransition(TransitionFact, usedTransitions)
		sections = append(sections, transition + " " + nesting)
	}
	
	// 7. Amazing Abilities
	abilities := fg.generateAmazingAbilities(bird, wikiData)
	if abilities != "" {
		// This section already has "Here's something incredible" built in
		sections = append(sections, abilities)
	}
	
	// 8. Conservation
	conservation := fg.generateConservationInfo(bird)
	if conservation != "" {
		sections = append(sections, conservation)
	}
	
	// 9. Fun Facts
	funFacts := fg.generateEnhancedFunFacts(bird, wikiData)
	if funFacts != "" {
		sections = append(sections, funFacts)
	}
	
	// Join sections with natural flow
	return fg.joinSectionsNaturally(sections, bird.CommonName)
}

// generateScientificIntro creates the scientific introduction
func (fg *ImprovedFactGeneratorV3) generateScientificIntro(bird *models.Bird) string {
	intros := []string{
		fmt.Sprintf("Let me tell you about the amazing %s! Its scientific name is %s.", bird.CommonName, bird.ScientificName),
		fmt.Sprintf("Today we're learning about the %s! Scientists call it %s.", bird.CommonName, bird.ScientificName),
		fmt.Sprintf("Get ready to discover the %s! Its scientific name is %s.", bird.CommonName, bird.ScientificName),
		fmt.Sprintf("Welcome to the world of the %s! In science, we call it %s.", bird.CommonName, bird.ScientificName),
	}
	
	intro := intros[fg.rng.Intn(len(intros))]
	
	if bird.Family != "" {
		familyName := bird.Family
		if strings.HasSuffix(familyName, "idae") {
			familyName = strings.TrimSuffix(familyName, "idae")
			familyPhrases := []string{
				fmt.Sprintf(" It belongs to the %s family of birds.", familyName),
				fmt.Sprintf(" This bird is part of the %s family.", familyName),
				fmt.Sprintf(" It's a member of the %s bird family.", familyName),
			}
			intro += familyPhrases[fg.rng.Intn(len(familyPhrases))]
		}
	}
	
	return intro
}

// generateEnhancedPhysicalDescription creates physical description without awkward transitions
func (fg *ImprovedFactGeneratorV3) generateEnhancedPhysicalDescription(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData == nil {
		return fg.generateBasicPhysicalDescription(bird)
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
		
		// Look for physical descriptions
		if (strings.Contains(lower, "color") || strings.Contains(lower, "colour") ||
			strings.Contains(lower, "plumage") || strings.Contains(lower, "feather") ||
			strings.Contains(lower, "size") || strings.Contains(lower, "length") ||
			strings.Contains(lower, "wingspan") || strings.Contains(lower, "cm") ||
			strings.Contains(lower, "inch") || strings.Contains(lower, "crest") ||
			strings.Contains(lower, "stripe") || strings.Contains(lower, "marking")) &&
			!strings.Contains(lower, "genus") && !strings.Contains(lower, "family") &&
			len(sentence) < 200 && len(physicalFacts) < 2 {
			
			// Make colors more vivid
			sentence = fg.makeColorDescriptionVivid(sentence)
			// Add size comparisons if applicable
			if strings.Contains(lower, "cm") || strings.Contains(lower, "inch") {
				sentence = fg.makeSizeComparison(sentence)
			}
			
			physicalFacts = append(physicalFacts, sentence)
			usedSentences[lower] = true
		}
		
		if len(physicalFacts) >= 2 {
			break
		}
	}
	
	if len(physicalFacts) > 0 {
		return strings.Join(physicalFacts, " ")
	}
	
	return fg.generateBasicPhysicalDescription(bird)
}

// generateVocalizationDescription creates sound description with appropriate intro
func (fg *ImprovedFactGeneratorV3) generateVocalizationDescription(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	// First check for Wikipedia content about sounds
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			if (strings.Contains(lower, "song") || strings.Contains(lower, "call") ||
				strings.Contains(lower, "sing") || strings.Contains(lower, "voice") ||
				strings.Contains(lower, "sound") || strings.Contains(lower, "whistle")) &&
				!strings.Contains(lower, "genus") && len(sentence) < 250 {
				
				// Use appropriate action-based intro
				intros := []string{
					"Now listen for their special sounds! ",
					"Their voice is unique! ",
					"You'll love their song! ",
					"Their calls are distinctive! ",
				}
				return intros[fg.rng.Intn(len(intros))] + sentence
			}
		}
	}
	
	// Bird-specific vocalizations with varied intros
	lowerName := strings.ToLower(bird.CommonName)
	soundIntros := []string{
		"Listen for their sound! ",
		"Their voice is special! ",
		"You can identify them by their call! ",
		"Here's what they sound like: ",
	}
	
	intro := soundIntros[fg.rng.Intn(len(soundIntros))]
	
	if strings.Contains(lowerName, "robin") {
		return intro + "Robins sing a cheerful melody that sounds like 'cheerily, cheer-up, cheerio!' They're often the first birds to sing at dawn."
	} else if strings.Contains(lowerName, "cardinal") {
		return intro + "Cardinals whistle clear, loud notes like 'birdy-birdy-birdy' or 'cheer-cheer-cheer.' Both males and females sing!"
	} else if strings.Contains(lowerName, "blue jay") {
		return intro + "Blue Jays make many sounds - they scream 'jay-jay!', make bell-like notes, and can even mimic hawk calls!"
	} else if strings.Contains(lowerName, "owl") {
		return intro + "Great Horned Owls make deep 'hoo-hoo-hoo' hoots, but can also shriek and hiss when protecting their young."
	} else if strings.Contains(lowerName, "hummingbird") {
		return intro + "The humming comes from their wings beating up to 80 times per second! They also make tiny chirping sounds."
	}
	
	return ""
}

// generateEnhancedHabitatBehavior creates habitat info with better flow
func (fg *ImprovedFactGeneratorV3) generateEnhancedHabitatBehavior(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	var habitatFacts []string
	
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			if (strings.Contains(lower, "habitat") || strings.Contains(lower, "found in") ||
				strings.Contains(lower, "lives") || strings.Contains(lower, "inhabit") ||
				strings.Contains(lower, "migrate") || strings.Contains(lower, "breed")) &&
				!strings.Contains(lower, "genus") && len(sentence) < 250 {
				
				sentence = fg.makeHabitatEngaging(sentence)
				habitatFacts = append(habitatFacts, sentence)
				
				if len(habitatFacts) >= 2 {
					break
				}
			}
		}
	}
	
	// Add seasonal tip if needed
	seasonalTip := fg.getSeasonalWatchingTip(bird)
	if seasonalTip != "" && len(habitatFacts) < 2 {
		habitatFacts = append(habitatFacts, seasonalTip)
	}
	
	if len(habitatFacts) > 0 {
		return strings.Join(habitatFacts, " ")
	}
	
	return fmt.Sprintf("You might spot %ss in parks, gardens, or natural areas near you.", bird.CommonName)
}

// generateEnhancedDietInfo creates diet information
func (fg *ImprovedFactGeneratorV3) generateEnhancedDietInfo(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			if (strings.Contains(lower, "diet") || strings.Contains(lower, "eat") ||
				strings.Contains(lower, "feed") || strings.Contains(lower, "prey")) &&
				!strings.Contains(lower, "genus") && len(sentence) < 250 {
				
				sentence = fg.makeDietDescriptionVivid(sentence)
				
				// Add watching tip
				watchingTips := []string{
					"Watch how they hunt! ",
					"Look at their feeding style! ",
					"Notice their eating habits! ",
					"See how they find food! ",
				}
				return watchingTips[fg.rng.Intn(len(watchingTips))] + sentence
			}
		}
	}
	
	// Default diet descriptions with action
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "hawk") || strings.Contains(lowerName, "eagle") {
		return "Watch them hunt! They soar high above, spot prey with super-sharp eyes, then swoop down at incredible speeds!"
	} else if strings.Contains(lowerName, "hummingbird") {
		return "See them feed! They hover at flowers, zipping from bloom to bloom, using long tongues to slurp nectar and snatching tiny insects from the air!"
	} else if strings.Contains(lowerName, "woodpecker") {
		return "Watch them work! They hammer their beaks into bark - tap, tap, tap! Their long tongues snake into holes to grab hiding insects."
	}
	
	return "Notice how they search for food - hopping, pecking, and exploring to find seeds, insects, and berries."
}

// generateNestingInfo creates nesting information
func (fg *ImprovedFactGeneratorV3) generateNestingInfo(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			if (strings.Contains(lower, "nest") || strings.Contains(lower, "egg") ||
				strings.Contains(lower, "chick") || strings.Contains(lower, "young")) &&
				!strings.Contains(lower, "genus") && len(sentence) < 250 {
				
				sentence = strings.ReplaceAll(sentence, "The female", "The mother bird")
				sentence = strings.ReplaceAll(sentence, "The male", "The father bird")
				return sentence
			}
		}
	}
	
	// Bird-specific nesting facts
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "robin") {
		return "Robin parents are amazing! Mom lays 3-5 beautiful blue eggs. After two weeks, tiny pink babies hatch with no feathers! Both parents work hard feeding them."
	} else if strings.Contains(lowerName, "cardinal") {
		return "Cardinal parents work as a team! They build cozy nests in bushes and can raise 2-3 families each year. Babies leave after just 10 days!"
	} else if strings.Contains(lowerName, "hummingbird") {
		return "Hummingbird nests are tiny treasures - only walnut-sized, made from spider silk and plant down. The eggs are smaller than jelly beans!"
	}
	
	return ""
}

// generateAmazingAbilities finds the most amazing ability
func (fg *ImprovedFactGeneratorV3) generateAmazingAbilities(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		bestSentence := ""
		bestScore := 0
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			if strings.Contains(lower, "genus") || len(sentence) > 250 {
				continue
			}
			
			score := 0
			if strings.Contains(lower, "largest") || strings.Contains(lower, "smallest") {
				score += 3
			}
			if strings.Contains(lower, "fastest") || strings.Contains(lower, "speed") {
				score += 3
			}
			if strings.Contains(lower, "only") || strings.Contains(lower, "unique") {
				score += 2
			}
			if strings.Contains(lower, "record") || strings.Contains(lower, "extreme") {
				score += 2
			}
			
			if score > bestScore {
				bestScore = score
				bestSentence = sentence
			}
		}
		
		if bestSentence != "" {
			exclamations := []string{
				"Here's something incredible: ",
				"Amazing fact: ",
				"You won't believe this: ",
				"Incredible ability: ",
			}
			return exclamations[fg.rng.Intn(len(exclamations))] + bestSentence + "!"
		}
	}
	
	// Bird-specific abilities
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "hummingbird") {
		return "Incredible ability: Hummingbirds can fly backwards, upside down, and hover! Their hearts beat 1,200 times per minute!"
	} else if strings.Contains(lowerName, "owl") {
		return "Amazing fact: Owls can turn their heads 270 degrees and fly in complete silence thanks to special feathers!"
	}
	
	return ""
}

// generateConservationInfo creates conservation awareness
func (fg *ImprovedFactGeneratorV3) generateConservationInfo(bird *models.Bird) string {
	taxon, err := fg.inatClient.SearchTaxon(bird.CommonName)
	if err != nil {
		taxon, _ = fg.inatClient.SearchTaxon(bird.ScientificName)
	}
	
	if taxon != nil && taxon.ConservationStatus != nil {
		status := taxon.ConservationStatus.Status
		switch status {
		case "LC":
			return fmt.Sprintf("Good news! %s populations are healthy. You can help by providing bird feeders and keeping cats indoors!", bird.CommonName)
		case "NT":
			return fmt.Sprintf("Scientists watch %ss carefully to keep them safe. You can help by joining bird counts!", bird.CommonName)
		case "VU":
			return fmt.Sprintf("%ss need our help! Protect their homes and keep the environment clean.", bird.CommonName)
		case "EN":
			return fmt.Sprintf("%ss are endangered and special. Seeing one is witnessing something rare!", bird.CommonName)
		}
	}
	
	// Location-based sightings
	if taxon != nil && bird.Latitude != 0 {
		observations, _ := fg.inatClient.GetRecentObservations(taxon.ID, bird.Latitude, bird.Longitude)
		if len(observations) > 0 {
			return fmt.Sprintf("Bird watchers have spotted %ss in your area recently! Look for them in early morning or late afternoon.", bird.CommonName)
		}
	}
	
	return "You can be a scientist too! When you see this bird, you're helping researchers learn about them."
}

// generateEnhancedFunFacts creates fun facts
func (fg *ImprovedFactGeneratorV3) generateEnhancedFunFacts(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			if (strings.Contains(lower, "state bird") || strings.Contains(lower, "national") ||
				strings.Contains(lower, "symbol") || strings.Contains(lower, "million")) &&
				!strings.Contains(lower, "genus") && len(sentence) < 200 {
				
				return "Cool fact: " + sentence + "!"
			}
		}
	}
	
	// Bird-specific fun facts
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "robin") {
		return "Fun fact: Robins can have three families in one year! They're also symbols of spring!"
	} else if strings.Contains(lowerName, "cardinal") {
		return "Did you know? Cardinals get their red color from their food! They're the state bird of seven states!"
	}
	
	return "Keep watching - every bird has its own special story waiting to be discovered!"
}

// Helper functions

func (fg *ImprovedFactGeneratorV3) makeSizeComparison(sentence string) string {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(cm|centimeter|inch|in)`)
	matches := re.FindAllStringSubmatch(sentence, -1)
	
	if len(matches) > 0 {
		unit := matches[0][2]
		var floatVal float64
		fmt.Sscanf(matches[0][1], "%f", &floatVal)
		
		var comparison string
		if unit == "cm" || unit == "centimeter" {
			comparison = fg.getSizeComparisonFromCM(floatVal)
		} else if unit == "inch" || unit == "in" {
			comparison = fg.getSizeComparisonFromInches(floatVal)
		}
		
		if comparison != "" {
			return sentence + " That's " + comparison + "!"
		}
	}
	return sentence
}

func (fg *ImprovedFactGeneratorV3) getSizeComparisonFromCM(cm float64) string {
	switch {
	case cm < 10:
		return "smaller than your thumb"
	case cm < 15:
		return "about as long as a pencil"
	case cm < 25:
		return "about as long as a banana"
	case cm < 35:
		return "about as long as a ruler"
	default:
		return "bigger than a basketball"
	}
}

func (fg *ImprovedFactGeneratorV3) getSizeComparisonFromInches(inches float64) string {
	return fg.getSizeComparisonFromCM(inches * 2.54)
}

func (fg *ImprovedFactGeneratorV3) makeColorDescriptionVivid(sentence string) string {
	replacements := map[string]string{
		" red ":    " bright red ",
		" blue ":   " brilliant blue ",
		" yellow ": " sunny yellow ",
		" orange ": " vibrant orange ",
		" black ":  " jet black ",
		" white ":  " snowy white ",
		" brown ":  " rich brown ",
	}
	
	result := sentence
	for old, new := range replacements {
		if strings.Contains(result, old) && !strings.Contains(result, new) {
			result = strings.ReplaceAll(result, old, new)
		}
	}
	return result
}

func (fg *ImprovedFactGeneratorV3) makeHabitatEngaging(sentence string) string {
	sentence = strings.ReplaceAll(sentence, "It is found", "You can find them")
	sentence = strings.ReplaceAll(sentence, "They are found", "Look for them")
	sentence = strings.ReplaceAll(sentence, "inhabits", "makes its home in")
	return sentence
}

func (fg *ImprovedFactGeneratorV3) makeDietDescriptionVivid(sentence string) string {
	sentence = strings.ReplaceAll(sentence, "eats", "gobbles up")
	sentence = strings.ReplaceAll(sentence, "feeds on", "hunts for")
	sentence = strings.ReplaceAll(sentence, "diet consists of", "favorite foods include")
	return sentence
}

func (fg *ImprovedFactGeneratorV3) getSeasonalWatchingTip(bird *models.Bird) string {
	lowerName := strings.ToLower(bird.CommonName)
	
	tips := map[string][]string{
		"robin": {
			"In spring, watch robins hop on lawns hunting worms. In winter, they gather in flocks for berries!",
			"Robins return in spring as a sign of warmer weather. Look for them on grassy areas!",
			"Watch for robins pulling worms from wet grass after rain!",
		},
		"hummingbird": {
			"Look for hummingbirds April through October near colorful flowers!",
			"Put out a sugar-water feeder in summer to attract hummingbirds!",
			"Hummingbirds love red and orange flowers - plant some to attract them!",
		},
		"cardinal": {
			"Cardinals stay all year! They're easiest to spot in winter against snow.",
			"Cardinals visit feeders year-round - offer sunflower seeds!",
			"Look for bright red males and brownish females together!",
		},
	}
	
	for key, tipList := range tips {
		if strings.Contains(lowerName, key) {
			return tipList[fg.rng.Intn(len(tipList))]
		}
	}
	
	return ""
}

func (fg *ImprovedFactGeneratorV3) generateBasicPhysicalDescription(bird *models.Bird) string {
	return fmt.Sprintf("The %s has unique markings and colors that make it special.", bird.CommonName)
}

// joinSectionsNaturally combines sections with smooth flow
func (fg *ImprovedFactGeneratorV3) joinSectionsNaturally(sections []string, birdName string) string {
	if len(sections) == 0 {
		return fmt.Sprintf("The %s is an incredible bird! Listen for its unique song and watch for its special behaviors.", birdName)
	}
	
	// Join sections with single spaces
	result := strings.Join(sections, " ")
	
	// Clean up any double spaces
	result = strings.ReplaceAll(result, "  ", " ")
	
	// Add engaging closing
	closings := []string{
		fmt.Sprintf(" Now you're a %s expert! See if you can spot one on your next adventure!", birdName),
		fmt.Sprintf(" Keep watching for %ss - you might discover something new!", birdName),
		fmt.Sprintf(" The %s is truly amazing! Happy bird watching!", birdName),
		fmt.Sprintf(" Now go outside and look for a %s!", birdName),
	}
	
	if len(result) < 1500 {
		result += closings[fg.rng.Intn(len(closings))]
	}
	
	return result
}

// EstimateReadingTime calculates approximate speech duration
func (fg *ImprovedFactGeneratorV3) EstimateReadingTime(text string) int {
	words := len(strings.Fields(text))
	return int(math.Ceil(float64(words) / 150.0 * 60))
}