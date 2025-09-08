package services

import "github.com/callen/bird-song-explorer/internal/models"

// FactGenerator defines the interface for bird fact generation
type FactGenerator interface {
	// GenerateFactScript creates a fact script for a bird
	// Returns the generated text script (not audio)
	GenerateFactScript(bird *models.Bird, latitude, longitude float64) string
	
	// GetGeneratorType returns the type of generator (basic or enhanced)
	GetGeneratorType() string
}

// FactGeneratorFactory creates the appropriate fact generator based on configuration
func NewFactGenerator(generatorType string, ebirdAPIKey string) FactGenerator {
	switch generatorType {
	case "enhanced":
		// Use the enhanced generator (formerly V4)
		return NewEnhancedFactGenerator(ebirdAPIKey)
	default:
		// Use the basic generator (current standard)
		return NewBasicFactGenerator()
	}
}