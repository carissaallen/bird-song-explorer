package services

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/pkg/inaturalist"
	"github.com/callen/bird-song-explorer/pkg/wikipedia"
)

// ImprovedFactGeneratorV2 generates comprehensive, engaging bird facts for the Explorer's Guide
type ImprovedFactGeneratorV2 struct {
	wikiClient *wikipedia.Client
	inatClient *inaturalist.Client
}

// NewImprovedFactGeneratorV2 creates a new enhanced fact generator
func NewImprovedFactGeneratorV2() *ImprovedFactGeneratorV2 {
	return &ImprovedFactGeneratorV2{
		wikiClient: wikipedia.NewClient(),
		inatClient: inaturalist.NewClient(),
	}
}

// GenerateExplorersGuideScript creates an engaging, comprehensive script for the Bird Explorer's Guide track
func (fg *ImprovedFactGeneratorV2) GenerateExplorersGuideScript(bird *models.Bird) string {
	var sections []string
	
	// Get Wikipedia data
	wikiData, _ := fg.wikiClient.GetBirdSummary(bird.CommonName)
	
	// 1. Scientific Introduction (moved to start as requested)
	scientificIntro := fg.generateScientificIntro(bird)
	if scientificIntro != "" {
		sections = append(sections, scientificIntro)
	}
	
	// 2. Physical Description with size comparisons
	physicalDesc := fg.generateEnhancedPhysicalDescription(bird, wikiData)
	if physicalDesc != "" {
		sections = append(sections, physicalDesc)
	}
	
	// 3. Vocalizations and Sounds (NEW)
	vocalDesc := fg.generateVocalizationDescription(bird, wikiData)
	if vocalDesc != "" {
		sections = append(sections, vocalDesc)
	}
	
	// 4. Habitat and Behavior with seasonal info
	habitat := fg.generateEnhancedHabitatBehavior(bird, wikiData)
	if habitat != "" {
		sections = append(sections, habitat)
	}
	
	// 5. Diet and Feeding with action words
	diet := fg.generateEnhancedDietInfo(bird, wikiData)
	if diet != "" {
		sections = append(sections, diet)
	}
	
	// 6. Nesting and Baby Birds (NEW)
	nesting := fg.generateNestingInfo(bird, wikiData)
	if nesting != "" {
		sections = append(sections, nesting)
	}
	
	// 7. Amazing Abilities and Records (NEW)
	abilities := fg.generateAmazingAbilities(bird, wikiData)
	if abilities != "" {
		sections = append(sections, abilities)
	}
	
	// 8. Conservation and Observations
	conservation := fg.generateConservationInfo(bird)
	if conservation != "" {
		sections = append(sections, conservation)
	}
	
	// 9. Fun Facts or Special Features
	funFacts := fg.generateEnhancedFunFacts(bird, wikiData)
	if funFacts != "" {
		sections = append(sections, funFacts)
	}
	
	// Join sections with natural transitions
	return fg.joinSectionsNaturally(sections, bird.CommonName)
}

// generateScientificIntro creates the scientific introduction
func (fg *ImprovedFactGeneratorV2) generateScientificIntro(bird *models.Bird) string {
	intro := fmt.Sprintf("Let me tell you about the amazing %s! Its scientific name is %s.", 
		bird.CommonName, bird.ScientificName)
	
	if bird.Family != "" {
		// Make family names more kid-friendly
		familyName := bird.Family
		if strings.HasSuffix(familyName, "idae") {
			familyName = strings.TrimSuffix(familyName, "idae")
			intro += fmt.Sprintf(" It belongs to the %s family of birds.", familyName)
		} else {
			intro += fmt.Sprintf(" It's part of the %s bird family.", familyName)
		}
	}
	
	return intro
}

// generateEnhancedPhysicalDescription creates detailed physical description with comparisons
func (fg *ImprovedFactGeneratorV2) generateEnhancedPhysicalDescription(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData == nil {
		return fg.generateBasicPhysicalDescription(bird)
	}
	
	// Use Wikipedia extract for content
	combinedText := wikiData.Extract
	sentences := strings.Split(combinedText, ". ")
	
	var physicalFacts []string
	usedSentences := make(map[string]bool) // Track used sentences to avoid repetition
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		lower := strings.ToLower(sentence)
		
		// Skip if already used similar content
		if usedSentences[lower] {
			continue
		}
		
		// Extract size information with measurements
		if (strings.Contains(lower, "cm") || strings.Contains(lower, "inch") || 
		   strings.Contains(lower, "length") || strings.Contains(lower, "wingspan")) &&
		   !strings.Contains(lower, "genus") && len(sentence) < 200 {
			// Convert measurements to kid-friendly comparisons
			sizeInfo := fg.makeSizeComparison(sentence)
			if sizeInfo != "" {
				physicalFacts = append(physicalFacts, sizeInfo)
				usedSentences[lower] = true
			}
		}
		
		// Extract color descriptions
		if (strings.Contains(lower, "color") || strings.Contains(lower, "colour") ||
			strings.Contains(lower, "plumage") || strings.Contains(lower, "feather") ||
			strings.Contains(lower, "red") || strings.Contains(lower, "blue") ||
			strings.Contains(lower, "black") || strings.Contains(lower, "white")) &&
			!strings.Contains(lower, "genus") && !strings.Contains(lower, "family") && 
			len(sentence) < 200 && len(physicalFacts) < 2 {
			colorInfo := fg.makeColorDescriptionVivid(sentence)
			if colorInfo != "" && !usedSentences[lower] {
				physicalFacts = append(physicalFacts, colorInfo)
				usedSentences[lower] = true
			}
		}
		
		// Look for distinctive features
		if (strings.Contains(lower, "crest") || strings.Contains(lower, "stripe") ||
			strings.Contains(lower, "spot") || strings.Contains(lower, "patch") ||
			strings.Contains(lower, "distinctive") || strings.Contains(lower, "marking")) &&
			!strings.Contains(lower, "genus") && len(sentence) < 200 && 
			len(physicalFacts) < 3 && !usedSentences[lower] {
			physicalFacts = append(physicalFacts, sentence)
			usedSentences[lower] = true
		}
		
		if len(physicalFacts) >= 2 {
			break
		}
	}
	
	if len(physicalFacts) > 0 {
		return strings.Join(physicalFacts, ". ") + "."
	}
	
	return fg.generateBasicPhysicalDescription(bird)
}

// generateVocalizationDescription creates description of bird sounds (NEW)
func (fg *ImprovedFactGeneratorV2) generateVocalizationDescription(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Look for sound descriptions
			if (strings.Contains(lower, "song") || strings.Contains(lower, "call") ||
				strings.Contains(lower, "sing") || strings.Contains(lower, "voice") ||
				strings.Contains(lower, "sound") || strings.Contains(lower, "whistle") ||
				strings.Contains(lower, "chirp") || strings.Contains(lower, "tweet") ||
				strings.Contains(lower, "trill") || strings.Contains(lower, "warble") ||
				strings.Contains(lower, "hoot") || strings.Contains(lower, "screech")) &&
				!strings.Contains(lower, "genus") && !strings.Contains(lower, "species") &&
				len(sentence) < 250 {
				
				// Make it more descriptive
				sentence = fg.enhanceVocalizationDescription(sentence)
				return "Listen for their special sound! " + sentence + "."
			}
		}
	}
	
	// Bird-specific vocalizations
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "robin") {
		return "Listen for their cheerful song! Robins sing a lovely melody that sounds like 'cheerily, cheer-up, cheerio!' They're often the first birds to sing at dawn."
	} else if strings.Contains(lowerName, "cardinal") {
		return "Listen carefully! Cardinals sing a clear, loud whistle that sounds like 'birdy-birdy-birdy' or 'cheer-cheer-cheer.' Both males and females sing these beautiful songs!"
	} else if strings.Contains(lowerName, "blue jay") {
		return "Blue Jays make many different sounds! They can scream 'jay-jay!', make bell-like sounds, and even copy the calls of hawks to scare other birds."
	} else if strings.Contains(lowerName, "owl") {
		return "Owls don't just hoot! Great Horned Owls make deep 'hoo-hoo-hoo' sounds, but they can also shriek, hiss, and make bill-snapping sounds when protecting their young."
	} else if strings.Contains(lowerName, "woodpecker") {
		return "Woodpeckers don't sing - they drum! They rapidly peck on trees making a loud 'rat-a-tat-tat' sound that can be heard from far away. Each species has its own drumming pattern!"
	} else if strings.Contains(lowerName, "hummingbird") {
		return "Hummingbirds make tiny chirping sounds, but the most amazing sound is the humming of their wings - beating up to 80 times per second!"
	}
	
	return ""
}

// generateEnhancedHabitatBehavior creates habitat description with seasonal info
func (fg *ImprovedFactGeneratorV2) generateEnhancedHabitatBehavior(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	var habitatFacts []string
	
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Look for habitat and seasonal information
			if (strings.Contains(lower, "habitat") || strings.Contains(lower, "found in") ||
				strings.Contains(lower, "lives in") || strings.Contains(lower, "inhabit") ||
				strings.Contains(lower, "migrate") || strings.Contains(lower, "winter") ||
				strings.Contains(lower, "summer") || strings.Contains(lower, "spring") ||
				strings.Contains(lower, "autumn") || strings.Contains(lower, "fall") ||
				strings.Contains(lower, "breed") || strings.Contains(lower, "nest")) &&
				!strings.Contains(lower, "genus") && len(sentence) < 250 {
				
				// Make it more engaging
				sentence = fg.makeHabitatEngaging(sentence)
				habitatFacts = append(habitatFacts, sentence)
				
				if len(habitatFacts) >= 2 {
					break
				}
			}
		}
	}
	
	// Add seasonal watching tips
	if len(habitatFacts) < 2 {
		seasonalTip := fg.getSeasonalWatchingTip(bird)
		if seasonalTip != "" {
			habitatFacts = append(habitatFacts, seasonalTip)
		}
	}
	
	if len(habitatFacts) > 0 {
		return strings.Join(habitatFacts, ". ") + "."
	}
	
	return fmt.Sprintf("You might spot a %s in parks, gardens, or natural areas near you! Keep your eyes open when you're outside.", bird.CommonName)
}

// generateEnhancedDietInfo creates diet info with action words
func (fg *ImprovedFactGeneratorV2) generateEnhancedDietInfo(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Look for diet information with action words
			if (strings.Contains(lower, "diet") || strings.Contains(lower, "eat") ||
				strings.Contains(lower, "feed") || strings.Contains(lower, "food") ||
				strings.Contains(lower, "hunt") || strings.Contains(lower, "catch") ||
				strings.Contains(lower, "forage") || strings.Contains(lower, "peck") ||
				strings.Contains(lower, "swallow") || strings.Contains(lower, "prey")) &&
				!strings.Contains(lower, "genus") && len(sentence) < 250 {
				
				// Add action words and make it vivid
				sentence = fg.makeDietDescriptionVivid(sentence)
				return sentence + "."
			}
		}
	}
	
	// Enhanced diet descriptions with action words
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "hawk") || strings.Contains(lowerName, "eagle") || strings.Contains(lowerName, "owl") {
		return "These amazing hunters soar high above, using their super-sharp eyes to spot prey. They swoop down at incredible speeds to catch small animals!"
	} else if strings.Contains(lowerName, "hummingbird") {
		return "Watch them hover at flowers! They zip from bloom to bloom, using their long tongues to slurp up sweet nectar. They also snatch tiny insects right out of the air!"
	} else if strings.Contains(lowerName, "woodpecker") {
		return "They hammer their strong beaks into tree bark - tap, tap, tap! Their long, sticky tongues snake into the holes to grab hiding insects. Yum!"
	} else if strings.Contains(lowerName, "robin") {
		return "Robins hop across lawns, tilting their heads to listen for worms moving underground. When they spot one, they quickly tug it out of the ground!"
	}
	
	return "Watch how they search for food! They hop, peck, and explore to find tasty seeds, juicy insects, and sweet berries."
}

// generateNestingInfo creates info about nests and baby birds (NEW)
func (fg *ImprovedFactGeneratorV2) generateNestingInfo(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Look for nesting and breeding information
			if (strings.Contains(lower, "nest") || strings.Contains(lower, "egg") ||
				strings.Contains(lower, "chick") || strings.Contains(lower, "young") ||
				strings.Contains(lower, "baby") || strings.Contains(lower, "fledg") ||
				strings.Contains(lower, "hatch") || strings.Contains(lower, "incubat") ||
				strings.Contains(lower, "clutch")) &&
				!strings.Contains(lower, "genus") && len(sentence) < 250 {
				
				// Make it kid-friendly
				sentence = strings.ReplaceAll(sentence, "The female", "The mother bird")
				sentence = strings.ReplaceAll(sentence, "The male", "The father bird")
				sentence = strings.ReplaceAll(sentence, "clutch", "group of eggs")
				
				return "About their babies: " + sentence + "."
			}
		}
	}
	
	// Bird-specific nesting facts
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "robin") {
		return "Baby robins are amazing! The mother lays 3-5 beautiful blue eggs. After two weeks, tiny pink babies hatch with no feathers! Both parents work hard feeding them worms all day long."
	} else if strings.Contains(lowerName, "cardinal") {
		return "Cardinal parents are great teamwork! They build cozy nests in thick bushes and raise 2-3 broods each year. The babies leave the nest after just 10 days!"
	} else if strings.Contains(lowerName, "hummingbird") {
		return "Hummingbird nests are tiny treasures! They're only as big as a walnut, made from spider silk and plant down. The eggs are smaller than jelly beans!"
	} else if strings.Contains(lowerName, "woodpecker") {
		return "Woodpeckers are nature's carpenters! They chisel out holes in trees for their nests. The babies stay safe inside for about a month before they're ready to fly."
	}
	
	return ""
}

// generateAmazingAbilities creates info about special abilities and records (NEW)
func (fg *ImprovedFactGeneratorV2) generateAmazingAbilities(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		bestSentence := ""
		bestScore := 0
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Skip technical or repetitive content
			if strings.Contains(lower, "genus") || strings.Contains(lower, "family") ||
			   strings.Contains(lower, "species") || len(sentence) > 250 {
				continue
			}
			
			// Score sentences based on interesting ability keywords
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
			if strings.Contains(lower, "can") || strings.Contains(lower, "able") {
				score += 1
			}
			if strings.Contains(lower, "mph") || strings.Contains(lower, "kilometer") {
				score += 2
			}
			
			if score > bestScore {
				bestScore = score
				bestSentence = sentence
			}
		}
		
		if bestSentence != "" {
			return "Here's something incredible: " + bestSentence + "!"
		}
	}
	
	// Bird-specific amazing abilities
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "hummingbird") {
		return "Here's something incredible: Hummingbirds can fly backwards, upside down, and hover in place! Their hearts beat over 1,200 times per minute!"
	} else if strings.Contains(lowerName, "owl") {
		return "Here's something incredible: Owls can turn their heads 270 degrees - that's three-quarters of the way around! Their flight is completely silent thanks to special feathers."
	} else if strings.Contains(lowerName, "woodpecker") {
		return "Here's something incredible: Woodpeckers can peck 20 times per second! Their brains are protected by special spongy bone that works like a helmet."
	} else if strings.Contains(lowerName, "cardinal") {
		return "Here's something incredible: Cardinals can live up to 15 years in the wild! They're one of the few birds where both males and females sing."
	} else if strings.Contains(lowerName, "eagle") {
		return "Here's something incredible: Bald Eagles can see four times better than humans and spot prey from 2 miles away! They can dive at speeds up to 100 mph!"
	}
	
	return ""
}

// generateConservationInfo creates conservation and observation information
func (fg *ImprovedFactGeneratorV2) generateConservationInfo(bird *models.Bird) string {
	// Try to get iNaturalist data
	taxon, err := fg.inatClient.SearchTaxon(bird.CommonName)
	if err != nil {
		taxon, _ = fg.inatClient.SearchTaxon(bird.ScientificName)
	}
	
	if taxon != nil && taxon.ConservationStatus != nil {
		status := taxon.ConservationStatus.Status
		switch status {
		case "LC":
			return fmt.Sprintf("Great news! The %s population is healthy and stable. You can help them by putting out bird feeders and keeping cats indoors!", bird.CommonName)
		case "NT":
			return fmt.Sprintf("Scientists are keeping a close eye on %ss to make sure they stay safe. You can help by participating in bird counts!", bird.CommonName)
		case "VU":
			return fmt.Sprintf("The %s needs our help! We can protect them by taking care of the places where they live and keeping our environment clean.", bird.CommonName)
		case "EN":
			return fmt.Sprintf("The %s is endangered and very special. If you're lucky enough to see one, you're witnessing something rare and precious!", bird.CommonName)
		case "CR":
			return fmt.Sprintf("The %s is one of the rarest birds on Earth. Every sighting helps scientists protect them. You could be a hero for these birds!", bird.CommonName)
		}
	}
	
	// Get recent observations if location is available
	if taxon != nil && bird.Latitude != 0 && bird.Longitude != 0 {
		observations, _ := fg.inatClient.GetRecentObservations(taxon.ID, bird.Latitude, bird.Longitude)
		if len(observations) > 0 {
			return fmt.Sprintf("Exciting news! Other bird watchers have spotted %ss in your area recently! Look for them in the early morning or late afternoon when they're most active.", bird.CommonName)
		}
	}
	
	return "You can be a citizen scientist! When you see this bird, you're helping researchers learn more about them. Try drawing or photographing what you see!"
}

// generateEnhancedFunFacts creates more engaging fun facts
func (fg *ImprovedFactGeneratorV2) generateEnhancedFunFacts(bird *models.Bird, wikiData *wikipedia.PageSummary) string {
	if wikiData != nil {
		combinedText := wikiData.Extract
		sentences := strings.Split(combinedText, ". ")
		
		// Look for interesting trivia
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Cultural or interesting facts
			if (strings.Contains(lower, "state bird") || strings.Contains(lower, "national") ||
				strings.Contains(lower, "symbol") || strings.Contains(lower, "folklore") ||
				strings.Contains(lower, "legend") || strings.Contains(lower, "named") ||
				strings.Contains(lower, "famous") || strings.Contains(lower, "popular") ||
				strings.Contains(lower, "million") || strings.Contains(lower, "population")) &&
				!strings.Contains(lower, "genus") && !strings.Contains(lower, "species") &&
				len(sentence) < 200 {
				
				return "Did you know? " + sentence + "!"
			}
		}
	}
	
	// Enhanced bird-specific fun facts
	lowerName := strings.ToLower(bird.CommonName)
	if strings.Contains(lowerName, "robin") {
		return "Did you know? American Robins can produce three successful broods in one year! They're also symbols of spring - many people get excited when they see the first robin of the year!"
	} else if strings.Contains(lowerName, "cardinal") {
		return "Did you know? Cardinals get their bright red color from the foods they eat! They're also the state bird of seven different states - more than any other bird!"
	} else if strings.Contains(lowerName, "blue jay") {
		return "Did you know? Blue Jays aren't really blue! Their feathers have no blue pigment - it's a trick of light called structural coloration. They're also super smart and can use tools!"
	} else if strings.Contains(lowerName, "eagle") {
		return "Did you know? Bald Eagles aren't actually bald - they have white feathers on their heads! They're the national bird of the United States and can live over 30 years!"
	}
	
	return "Keep watching and listening - every bird has its own special story. You might discover something no one else has noticed before!"
}

// Helper functions for making content more engaging

func (fg *ImprovedFactGeneratorV2) makeSizeComparison(sentence string) string {
	// Extract measurements and convert to comparisons
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(cm|centimeter|inch|in|mm|m)`)
	matches := re.FindAllStringSubmatch(sentence, -1)
	
	if len(matches) > 0 {
		unit := matches[0][2]
		
		var comparison string
		switch unit {
		case "cm", "centimeter":
			if value, _ := fmt.Sscanf(matches[0][1], "%f", new(float64)); value > 0 {
				floatVal := *new(float64)
				fmt.Sscanf(matches[0][1], "%f", &floatVal)
				comparison = fg.getSizeComparisonFromCM(floatVal)
			}
		case "inch", "in":
			if value, _ := fmt.Sscanf(matches[0][1], "%f", new(float64)); value > 0 {
				floatVal := *new(float64)
				fmt.Sscanf(matches[0][1], "%f", &floatVal)
				comparison = fg.getSizeComparisonFromInches(floatVal)
			}
		}
		
		if comparison != "" {
			return sentence + " That's " + comparison + "!"
		}
	}
	
	return sentence
}

func (fg *ImprovedFactGeneratorV2) getSizeComparisonFromCM(cm float64) string {
	switch {
	case cm < 10:
		return "smaller than your thumb"
	case cm < 15:
		return "about as long as a pencil"
	case cm < 25:
		return "about as long as a banana"
	case cm < 35:
		return "about as long as a ruler"
	case cm < 50:
		return "about as long as your arm from elbow to hand"
	case cm < 100:
		return "about as long as a baseball bat"
	default:
		return "bigger than a bicycle"
	}
}

func (fg *ImprovedFactGeneratorV2) getSizeComparisonFromInches(inches float64) string {
	cm := inches * 2.54
	return fg.getSizeComparisonFromCM(cm)
}

func (fg *ImprovedFactGeneratorV2) makeColorDescriptionVivid(sentence string) string {
	// Add vivid descriptors to colors
	replacements := map[string]string{
		"red":    "bright red",
		"blue":   "brilliant blue",
		"yellow": "sunny yellow",
		"orange": "vibrant orange",
		"green":  "emerald green",
		"black":  "jet black",
		"white":  "snowy white",
		"brown":  "rich brown",
		"gray":   "soft gray",
		"grey":   "soft grey",
	}
	
	result := sentence
	for old, new := range replacements {
		if strings.Contains(strings.ToLower(result), old) && !strings.Contains(strings.ToLower(result), new) {
			result = regexp.MustCompile(`(?i)\b`+old+`\b`).ReplaceAllString(result, new)
		}
	}
	
	return result
}

func (fg *ImprovedFactGeneratorV2) enhanceVocalizationDescription(sentence string) string {
	// Make sound descriptions more vivid
	sentence = strings.ReplaceAll(sentence, "calls", "makes exciting calls")
	sentence = strings.ReplaceAll(sentence, "sings", "sings beautifully")
	sentence = strings.ReplaceAll(sentence, "song is", "song sounds like")
	return sentence
}

func (fg *ImprovedFactGeneratorV2) makeHabitatEngaging(sentence string) string {
	// Make habitat descriptions more relatable
	sentence = strings.ReplaceAll(sentence, "It is found", "Look for them")
	sentence = strings.ReplaceAll(sentence, "They are found", "You can find them")
	sentence = strings.ReplaceAll(sentence, "inhabits", "makes its home in")
	sentence = strings.ReplaceAll(sentence, "inhabit", "make their homes in")
	sentence = strings.ReplaceAll(sentence, "resides in", "lives in")
	return sentence
}

func (fg *ImprovedFactGeneratorV2) makeDietDescriptionVivid(sentence string) string {
	// Add action words to diet descriptions
	sentence = strings.ReplaceAll(sentence, "eats", "gobbles up")
	sentence = strings.ReplaceAll(sentence, "feeds on", "hunts for")
	sentence = strings.ReplaceAll(sentence, "diet consists of", "favorite foods include")
	sentence = strings.ReplaceAll(sentence, "consumes", "devours")
	return sentence
}

func (fg *ImprovedFactGeneratorV2) getSeasonalWatchingTip(bird *models.Bird) string {
	lowerName := strings.ToLower(bird.CommonName)
	
	if strings.Contains(lowerName, "robin") {
		return "In spring, watch for robins hopping on lawns looking for worms. In winter, they gather in flocks to eat berries!"
	} else if strings.Contains(lowerName, "hummingbird") {
		return "Look for hummingbirds from April to October near flowers. Put out a feeder with sugar water to attract them!"
	} else if strings.Contains(lowerName, "cardinal") {
		return "Cardinals don't migrate, so you can see them all year! They're easiest to spot in winter against the snow."
	}
	
	return ""
}

func (fg *ImprovedFactGeneratorV2) generateBasicPhysicalDescription(bird *models.Bird) string {
	return fmt.Sprintf("The %s is a beautiful bird with unique markings and colors that help you identify it in the wild.", bird.CommonName)
}

// joinSectionsNaturally combines sections with smooth transitions
func (fg *ImprovedFactGeneratorV2) joinSectionsNaturally(sections []string, birdName string) string {
	if len(sections) == 0 {
		return fmt.Sprintf("The %s is an incredible bird! Listen carefully to learn its unique song and watch for its special behaviors.", birdName)
	}
	
	// Add varied transitions between sections for natural flow
	result := sections[0]
	
	transitions := []string{
		" ",
		" Here's something cool: ",
		" And guess what? ",
		" Want to know something amazing? ",
		" ",
		" Check this out! ",
		" ",
		" Here's a fun fact: ",
		" ",
	}
	
	for i := 1; i < len(sections); i++ {
		if sections[i] != "" {
			transition := transitions[i % len(transitions)]
			result += transition + sections[i]
		}
	}
	
	// Add engaging closing if not too long
	if len(result) < 1500 {
		result += fmt.Sprintf(" Now you're a %s expert! See if you can spot one on your next outdoor adventure!", birdName)
	}
	
	return result
}

// Calculate approximate reading time
func (fg *ImprovedFactGeneratorV2) EstimateReadingTime(text string) int {
	words := len(strings.Fields(text))
	// Approximate 150 words per minute for TTS
	return int(math.Ceil(float64(words) / 150.0 * 60))
}