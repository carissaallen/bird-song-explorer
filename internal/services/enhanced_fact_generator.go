package services

import (
	"github.com/callen/bird-song-explorer/internal/models"
)

// EnhancedFactGenerator wraps the existing ImprovedFactGeneratorV4
type EnhancedFactGenerator struct {
	v4Generator *ImprovedFactGeneratorV4
}

// NewEnhancedFactGenerator creates a new enhanced fact generator
func NewEnhancedFactGenerator(ebirdAPIKey string) *EnhancedFactGenerator {
	return &EnhancedFactGenerator{
		v4Generator: NewImprovedFactGeneratorV4(ebirdAPIKey),
	}
}

// GetGeneratorType returns the type of this generator
func (g *EnhancedFactGenerator) GetGeneratorType() string {
	return "enhanced"
}

// GenerateFactScript creates an enhanced fact script for a bird
func (g *EnhancedFactGenerator) GenerateFactScript(bird *models.Bird, latitude, longitude float64) string {
	// Use the existing V4 generator's method
	return g.v4Generator.GenerateExplorersGuideScriptWithLocation(bird, latitude, longitude)
}