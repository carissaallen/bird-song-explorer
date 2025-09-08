package services

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
)

// BasicFactGenerator generates simple, TTS-friendly bird facts
type BasicFactGenerator struct{}

// NewBasicFactGenerator creates a new basic fact generator
func NewBasicFactGenerator() *BasicFactGenerator {
	return &BasicFactGenerator{}
}

// GetGeneratorType returns the type of this generator
func (g *BasicFactGenerator) GetGeneratorType() string {
	return "basic"
}

// GenerateFactScript creates a simple fact script for a bird
func (g *BasicFactGenerator) GenerateFactScript(bird *models.Bird, latitude, longitude float64) string {
	// Extract scientific name if available
	scientificName := bird.ScientificName
	if scientificName == "" && bird.Description != "" {
		scientificName = g.extractScientificName(bird.Description)
	}

	// Get a simple fact from the description
	simpleFact := g.extractSimpleFact(bird.Description, bird.CommonName)
	
	// Get an additional generic fact
	additionalFact := g.getGenericBirdFact(bird.CommonName, simpleFact)

	// Build the text
	var script string
	if scientificName != "" {
		// Format with scientific name and enhanced fact
		script = fmt.Sprintf("The scientific name for the %s is %s. Did you know? %s %s Birds are found all over the world, each one perfectly adapted to its home!",
			bird.CommonName, scientificName, simpleFact, additionalFact)
	} else {
		// Enhanced version without scientific name
		script = fmt.Sprintf("Let me tell you about the amazing %s! Did you know? %s %s Every bird has its own special story. Listen carefully to learn its unique song!",
			bird.CommonName, simpleFact, additionalFact)
	}

	return script
}

// extractScientificName extracts the scientific name from a description
func (g *BasicFactGenerator) extractScientificName(description string) string {
	if strings.Contains(description, "(") && strings.Contains(description, ")") {
		start := strings.Index(description, "(")
		end := strings.Index(description, ")")
		if start < end && end-start < 50 {
			potentialName := description[start+1 : end]
			words := strings.Fields(potentialName)
			if len(words) == 2 && strings.Title(words[0]) == words[0] {
				return potentialName
			}
		}
	}
	return ""
}

// extractSimpleFact extracts a simple fact from the description
func (g *BasicFactGenerator) extractSimpleFact(description string, birdName string) string {
	if description == "" {
		return fmt.Sprintf("The %s is an amazing bird!", birdName)
	}

	// Remove scientific name if present
	simpleFact := description
	if scientificName := g.extractScientificName(description); scientificName != "" {
		start := strings.Index(description, "(")
		end := strings.Index(description, ")")
		if start < end {
			simpleFact = description[:start] + description[end+1:]
		}
	}

	// Clean up and simplify
	simpleFact = strings.TrimSpace(simpleFact)
	parts := strings.Split(simpleFact, ".")
	if len(parts) > 0 {
		if strings.Contains(strings.ToLower(parts[0]), "is a") {
			simpleFact = strings.TrimSpace(parts[0]) + "."
		} else if len(parts) > 1 {
			simpleFact = strings.TrimSpace(parts[0]) + ". " + strings.TrimSpace(parts[1]) + "."
		}
	}

	// Remove overly technical content
	if strings.Contains(strings.ToLower(simpleFact), "derived from") ||
		strings.Contains(strings.ToLower(simpleFact), "greek") ||
		strings.Contains(strings.ToLower(simpleFact), "latin") {
		simpleFact = fmt.Sprintf("The %s is an amazing bird!", birdName)
	}

	return simpleFact
}

// getGenericBirdFact returns an interesting generic fact based on bird characteristics
func (g *BasicFactGenerator) getGenericBirdFact(birdName string, existingFact string) string {
	lowerName := strings.ToLower(birdName)

	// Size-based facts
	if strings.Contains(lowerName, "eagle") || strings.Contains(lowerName, "hawk") ||
		strings.Contains(lowerName, "owl") {
		return "Birds of prey have incredible eyesight - they can spot tiny movements from far away!"
	}

	if strings.Contains(lowerName, "hummingbird") {
		return "Hummingbirds are the only birds that can fly backwards and their hearts beat over one thousand two-hundred times per minute!"
	}

	// Water birds
	if strings.Contains(lowerName, "duck") || strings.Contains(lowerName, "goose") ||
		strings.Contains(lowerName, "swan") {
		return "Water birds have special oil glands that keep their feathers waterproof!"
	}

	// Songbirds
	if strings.Contains(lowerName, "robin") || strings.Contains(lowerName, "sparrow") ||
		strings.Contains(lowerName, "finch") || strings.Contains(lowerName, "warbler") {
		return "Songbirds learn their songs by listening to their parents, just like you learned to talk!"
	}

	// Colorful birds
	if strings.Contains(lowerName, "cardinal") || strings.Contains(lowerName, "blue") ||
		strings.Contains(lowerName, "gold") {
		return "Bright colors help birds recognize their own species and attract mates!"
	}

	// Nocturnal birds
	if strings.Contains(lowerName, "owl") || strings.Contains(lowerName, "nightjar") {
		return "Night birds have special feathers that let them fly almost silently!"
	}

	// Migration-related
	if strings.Contains(lowerName, "swallow") || strings.Contains(lowerName, "crane") ||
		strings.Contains(lowerName, "arctic") {
		return "Some birds travel thousands of miles each year, using the stars and Earth's magnetic field to navigate!"
	}

	// Intelligence
	if strings.Contains(lowerName, "crow") || strings.Contains(lowerName, "raven") ||
		strings.Contains(lowerName, "jay") {
		return "These birds are super smart - they can use tools and even recognize human faces!"
	}

	// Default interesting facts
	defaultFacts := []string{
		"Birds are the only animals with feathers - no other creature has them!",
		"A bird's bones are hollow, making them light enough to fly!",
		"Birds can see colors that humans can't even imagine!",
		"Most birds have excellent memories and can remember hundreds of food hiding spots!",
		"Baby birds can eat their own body weight in food every day!",
		"Birds help plants grow by spreading seeds in their droppings!",
		"Some birds can sleep with one half of their brain while the other stays awake!",
		"Birds existed alongside dinosaurs - they're living dinosaurs themselves!",
	}

	rand.Seed(time.Now().UnixNano())
	return defaultFacts[rand.Intn(len(defaultFacts))]
}