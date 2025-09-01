package services

import (
	"fmt"
	"strings"

	"github.com/callen/bird-song-explorer/internal/models"
	"github.com/callen/bird-song-explorer/pkg/inaturalist"
	"github.com/callen/bird-song-explorer/pkg/wikipedia"
)

// ImprovedFactGenerator generates comprehensive, organized bird facts for the Explorer's Guide
type ImprovedFactGenerator struct {
	wikiClient *wikipedia.Client
	inatClient *inaturalist.Client
}

// NewImprovedFactGenerator creates a new fact generator
func NewImprovedFactGenerator() *ImprovedFactGenerator {
	return &ImprovedFactGenerator{
		wikiClient: wikipedia.NewClient(),
		inatClient: inaturalist.NewClient(),
	}
}

// GenerateExplorersGuideScript creates a structured script for the Bird Explorer's Guide track
// Returns the full text to be converted to speech, organized for better flow
func (fg *ImprovedFactGenerator) GenerateExplorersGuideScript(bird *models.Bird) string {
	var sections []string
	
	// 1. Scientific Introduction (moved to start as requested)
	scientificIntro := fg.generateScientificIntro(bird)
	if scientificIntro != "" {
		sections = append(sections, scientificIntro)
	}
	
	// 2. Physical Description
	physicalDesc := fg.generatePhysicalDescription(bird)
	if physicalDesc != "" {
		sections = append(sections, physicalDesc)
	}
	
	// 3. Habitat and Behavior
	habitat := fg.generateHabitatBehavior(bird)
	if habitat != "" {
		sections = append(sections, habitat)
	}
	
	// 4. Diet and Feeding
	diet := fg.generateDietInfo(bird)
	if diet != "" {
		sections = append(sections, diet)
	}
	
	// 5. Conservation and Observations
	conservation := fg.generateConservationInfo(bird)
	if conservation != "" {
		sections = append(sections, conservation)
	}
	
	// 6. Fun Facts or Special Features
	funFacts := fg.generateFunFacts(bird)
	if funFacts != "" {
		sections = append(sections, funFacts)
	}
	
	// Join sections with natural transitions
	return fg.joinSectionsNaturally(sections, bird.CommonName)
}

// generateScientificIntro creates the scientific introduction
func (fg *ImprovedFactGenerator) generateScientificIntro(bird *models.Bird) string {
	intro := fmt.Sprintf("Let me tell you about the %s! Its scientific name is %s.", 
		bird.CommonName, bird.ScientificName)
	
	if bird.Family != "" {
		intro += fmt.Sprintf(" It belongs to the %s family of birds.", bird.Family)
	}
	
	return intro
}

// generatePhysicalDescription fetches and formats physical description from Wikipedia
func (fg *ImprovedFactGenerator) generatePhysicalDescription(bird *models.Bird) string {
	// Try to get Wikipedia summary (try lowercase first for better results)
	lowerName := strings.ToLower(bird.CommonName)
	summary, err := fg.wikiClient.GetBirdSummary(lowerName)
	if err != nil {
		// Try with original case
		summary, err = fg.wikiClient.GetBirdSummary(bird.CommonName)
	}
	if err != nil {
		// Try with scientific name
		summary, err = fg.wikiClient.GetBirdSummary(bird.ScientificName)
	}
	
	if err == nil && summary != nil && summary.Extract != "" {
		// Extract physical description sentences
		sentences := strings.Split(summary.Extract, ". ")
		var physicalSentences []string
		
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Look for physical characteristics
			if (strings.Contains(lower, "color") || strings.Contains(lower, "colour") ||
				strings.Contains(lower, "breast") || strings.Contains(lower, "wing") ||
				strings.Contains(lower, "tail") || strings.Contains(lower, "beak") ||
				strings.Contains(lower, "bill") || strings.Contains(lower, "feather") ||
				strings.Contains(lower, "plumage") || strings.Contains(lower, "size") ||
				strings.Contains(lower, "length") || strings.Contains(lower, "wingspan") ||
				strings.Contains(lower, "cm") || strings.Contains(lower, "inch") ||
				strings.Contains(lower, "orange") || strings.Contains(lower, "red") ||
				strings.Contains(lower, "blue") || strings.Contains(lower, "black") ||
				strings.Contains(lower, "white") || strings.Contains(lower, "brown") ||
				strings.Contains(lower, "gray") || strings.Contains(lower, "grey")) &&
				len(sentence) < 200 &&
				!strings.Contains(lower, "genus") &&
				!strings.Contains(lower, "taxonomy") {
				
				// Make it more kid-friendly
				sentence = strings.ReplaceAll(sentence, "It is", "This bird is")
				sentence = strings.ReplaceAll(sentence, "They are", "These birds are")
				physicalSentences = append(physicalSentences, sentence)
				
				if len(physicalSentences) >= 2 {
					break
				}
			}
		}
		
		if len(physicalSentences) > 0 {
			return strings.Join(physicalSentences, ". ") + "."
		}
	}
	
	// Fallback description
	return fmt.Sprintf("The %s is a beautiful bird with unique markings and colors.", bird.CommonName)
}

// generateHabitatBehavior creates habitat and behavior description
func (fg *ImprovedFactGenerator) generateHabitatBehavior(bird *models.Bird) string {
	// Try to get Wikipedia summary for habitat info
	lowerName := strings.ToLower(bird.CommonName)
	summary, _ := fg.wikiClient.GetBirdSummary(lowerName)
	
	var habitatInfo []string
	
	if summary != nil && summary.Extract != "" {
		sentences := strings.Split(summary.Extract, ". ")
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Look for habitat and behavior information
			if (strings.Contains(lower, "habitat") || strings.Contains(lower, "found in") ||
				strings.Contains(lower, "lives in") || strings.Contains(lower, "inhabit") ||
				strings.Contains(lower, "forest") || strings.Contains(lower, "woodland") ||
				strings.Contains(lower, "grassland") || strings.Contains(lower, "wetland") ||
				strings.Contains(lower, "urban") || strings.Contains(lower, "garden") ||
				strings.Contains(lower, "park") || strings.Contains(lower, "tree") ||
				strings.Contains(lower, "migrate") || strings.Contains(lower, "winter") ||
				strings.Contains(lower, "summer") || strings.Contains(lower, "breed") ||
				strings.Contains(lower, "nest") || strings.Contains(lower, "flock") ||
				strings.Contains(lower, "active") || strings.Contains(lower, "day") ||
				strings.Contains(lower, "night") || strings.Contains(lower, "dawn")) &&
				len(sentence) < 200 &&
				!strings.Contains(lower, "genus") {
				
				// Make it more kid-friendly
				sentence = strings.ReplaceAll(sentence, "It is found", "You can find them")
				sentence = strings.ReplaceAll(sentence, "They are found", "You can find them")
				sentence = strings.ReplaceAll(sentence, "It inhabits", "They live in")
				sentence = strings.ReplaceAll(sentence, "They inhabit", "They live in")
				habitatInfo = append(habitatInfo, sentence)
				
				if len(habitatInfo) >= 2 {
					break
				}
			}
		}
	}
	
	// Add regional information if available
	if bird.Region != "" && len(habitatInfo) < 2 {
		habitatInfo = append(habitatInfo, fmt.Sprintf("These birds can be found in %s", bird.Region))
	}
	
	if len(habitatInfo) > 0 {
		return strings.Join(habitatInfo, ". ") + "."
	}
	
	return fmt.Sprintf("You might spot a %s in parks, gardens, or natural areas near you!", bird.CommonName)
}

// generateDietInfo creates diet and feeding information
func (fg *ImprovedFactGenerator) generateDietInfo(bird *models.Bird) string {
	// Try to get Wikipedia summary for diet info
	lowerName := strings.ToLower(bird.CommonName)
	summary, _ := fg.wikiClient.GetBirdSummary(lowerName)
	
	if summary != nil && summary.Extract != "" {
		sentences := strings.Split(summary.Extract, ". ")
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Look for diet information
			if (strings.Contains(lower, "diet") || strings.Contains(lower, "eat") ||
				strings.Contains(lower, "feed") || strings.Contains(lower, "food") ||
				strings.Contains(lower, "seed") || strings.Contains(lower, "insect") ||
				strings.Contains(lower, "worm") || strings.Contains(lower, "berry") ||
				strings.Contains(lower, "fruit") || strings.Contains(lower, "nectar") ||
				strings.Contains(lower, "fish") || strings.Contains(lower, "prey") ||
				strings.Contains(lower, "hunt") || strings.Contains(lower, "forage")) &&
				len(sentence) < 200 {
				
				// Make it more kid-friendly
				sentence = strings.ReplaceAll(sentence, "Its diet consists of", "They love to eat")
				sentence = strings.ReplaceAll(sentence, "It feeds on", "They eat")
				sentence = strings.ReplaceAll(sentence, "They feed on", "They eat")
				
				if !strings.HasSuffix(sentence, ".") {
					sentence += "."
				}
				
				return sentence
			}
		}
	}
	
	// Generic diet info based on bird type
	if strings.Contains(strings.ToLower(bird.CommonName), "hawk") ||
		strings.Contains(strings.ToLower(bird.CommonName), "eagle") ||
		strings.Contains(strings.ToLower(bird.CommonName), "owl") {
		return "As a bird of prey, they hunt for small animals to eat."
	} else if strings.Contains(strings.ToLower(bird.CommonName), "hummingbird") {
		return "They drink sweet nectar from flowers and also catch tiny insects."
	} else if strings.Contains(strings.ToLower(bird.CommonName), "woodpecker") {
		return "They use their strong beaks to find insects hiding in tree bark."
	}
	
	return "Like many birds, they search for seeds, insects, and berries to eat."
}

// generateConservationInfo creates conservation and observation information
func (fg *ImprovedFactGenerator) generateConservationInfo(bird *models.Bird) string {
	// Try to get iNaturalist data
	taxon, err := fg.inatClient.SearchTaxon(bird.CommonName)
	if err != nil {
		taxon, _ = fg.inatClient.SearchTaxon(bird.ScientificName)
	}
	
	if taxon != nil && taxon.ConservationStatus != nil {
		status := taxon.ConservationStatus.Status
		switch status {
		case "LC":
			return fmt.Sprintf("Great news! The %s population is healthy and stable. These birds are doing well in nature!", bird.CommonName)
		case "NT":
			return fmt.Sprintf("Scientists are watching the %s carefully to make sure they stay safe and healthy.", bird.CommonName)
		case "VU":
			return fmt.Sprintf("The %s needs our help! We can protect them by taking care of the places where they live.", bird.CommonName)
		case "EN":
			return fmt.Sprintf("The %s is endangered and very special. If you see one, you're incredibly lucky!", bird.CommonName)
		case "CR":
			return fmt.Sprintf("The %s is one of the rarest birds. Every sighting helps scientists protect them!", bird.CommonName)
		}
	}
	
	// Get recent observations if location is available
	if taxon != nil && bird.Latitude != 0 && bird.Longitude != 0 {
		observations, _ := fg.inatClient.GetRecentObservations(taxon.ID, bird.Latitude, bird.Longitude)
		if len(observations) > 0 {
			return fmt.Sprintf("Bird watchers have recently spotted %ss in your area! Keep your eyes and ears open - you might see one too!", bird.CommonName)
		}
	}
	
	return "When you see this bird, you're helping scientists by being a nature observer!"
}

// generateFunFacts creates engaging fun facts
func (fg *ImprovedFactGenerator) generateFunFacts(bird *models.Bird) string {
	// Try to extract interesting facts from Wikipedia
	lowerName := strings.ToLower(bird.CommonName)
	summary, _ := fg.wikiClient.GetBirdSummary(lowerName)
	
	if summary != nil && summary.Extract != "" {
		sentences := strings.Split(summary.Extract, ". ")
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			lower := strings.ToLower(sentence)
			
			// Look for interesting behavioral facts
			if (strings.Contains(lower, "song") || strings.Contains(lower, "call") ||
				strings.Contains(lower, "sing") || strings.Contains(lower, "voice") ||
				strings.Contains(lower, "earliest") || strings.Contains(lower, "dawn") ||
				strings.Contains(lower, "loud") || strings.Contains(lower, "unique") ||
				strings.Contains(lower, "special") || strings.Contains(lower, "amazing") ||
				strings.Contains(lower, "fastest") || strings.Contains(lower, "largest") ||
				strings.Contains(lower, "smallest") || strings.Contains(lower, "most")) &&
				len(sentence) < 200 &&
				!strings.Contains(lower, "genus") {
				
				if !strings.HasSuffix(sentence, ".") {
					sentence += "."
				}
				return "Here's something amazing: " + sentence
			}
		}
	}
	
	// Bird-specific fun facts
	if strings.ToLower(bird.CommonName) == "american robin" {
		return "Fun fact: American Robins are often the first birds to sing in the morning and the last to stop singing at night!"
	} else if strings.Contains(strings.ToLower(bird.CommonName), "cardinal") {
		return "Fun fact: Cardinals are one of the few birds where both males and females sing beautiful songs!"
	} else if strings.Contains(strings.ToLower(bird.CommonName), "blue jay") {
		return "Fun fact: Blue Jays are super smart and can mimic the calls of hawks to scare other birds away from food!"
	} else if strings.Contains(strings.ToLower(bird.CommonName), "owl") {
		return "Fun fact: Owls can turn their heads almost all the way around - up to 270 degrees!"
	} else if strings.Contains(strings.ToLower(bird.CommonName), "hummingbird") {
		return "Fun fact: Hummingbirds are the only birds that can fly backwards!"
	} else if strings.Contains(strings.ToLower(bird.CommonName), "woodpecker") {
		return "Fun fact: Woodpeckers have special cushioning in their heads to protect their brains when they peck on trees!"
	}
	
	return "Listen carefully to learn this bird's special song - each species has its own unique way of singing!"
}

// joinSectionsNaturally combines sections with smooth transitions
func (fg *ImprovedFactGenerator) joinSectionsNaturally(sections []string, birdName string) string {
	if len(sections) == 0 {
		return fmt.Sprintf("The %s is an amazing bird! Listen carefully to learn its unique song.", birdName)
	}
	
	// Add transitions between sections for natural flow
	result := sections[0]
	
	for i := 1; i < len(sections); i++ {
		// Add varied transitions
		transitions := []string{
			" ",
			" Did you know? ",
			" Also, ",
			" Here's something interesting: ",
			" ",
		}
		
		transition := transitions[i % len(transitions)]
		result += transition + sections[i]
	}
	
	// Add closing if the text is short
	if len(result) < 500 {
		result += fmt.Sprintf(" The %s is truly a wonderful bird to observe and learn about!", birdName)
	}
	
	return result
}