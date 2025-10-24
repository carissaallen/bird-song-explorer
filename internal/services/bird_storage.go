package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// BirdMetadata represents the metadata for a bird species
type BirdMetadata struct {
	CommonName          string            `json:"common_name"`
	ScientificName      string            `json:"scientific_name"`
	Family              string            `json:"family"`
	Regions             []string          `json:"regions"`
	PrimaryHabitat      string            `json:"primary_habitat"`
	Habitats            []string          `json:"habitats"`
	ConservationStatus  string            `json:"conservation_status"`
	Size                BirdSize          `json:"size"`
	Diet                []string          `json:"diet"`
	BreedingSeason      string            `json:"breeding_season"`
	MigrationPattern    string            `json:"migration_pattern"`
	DistinctiveFeatures []string          `json:"distinctive_features"`
	FunFacts            []string          `json:"fun_facts"`
	SongLocations       map[string]string `json:"song_locations"`
	RecordedDate        string            `json:"recorded_date"`
	Narrator            string            `json:"narrator"`
}

// BirdSize represents the physical dimensions of a bird
type BirdSize struct {
	LengthCM   string `json:"length_cm"`
	WingspanCM string `json:"wingspan_cm"`
	WeightG    string `json:"weight_g"`
}

// BirdStorage handles bird data storage and retrieval
type BirdStorage struct {
	basePath string
}

// NewBirdStorage creates a new BirdStorage instance
func NewBirdStorage(basePath string) *BirdStorage {
	if basePath == "" {
		basePath = "./birds"
	}
	return &BirdStorage{
		basePath: basePath,
	}
}

// GetBirdMetadata retrieves metadata for a specific bird
func (bs *BirdStorage) GetBirdMetadata(birdName string) (*BirdMetadata, error) {
	// Convert bird name to directory format (lowercase, underscores)
	dirName := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))

	metadataPath := filepath.Join(bs.basePath, "_global_species", dirName, "metadata.json")

	data, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata for %s: %w", birdName, err)
	}

	var metadata BirdMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata for %s: %w", birdName, err)
	}

	return &metadata, nil
}

// GetSongPath returns the full path to a bird's song recording
func (bs *BirdStorage) GetSongPath(birdName, songType string) string {
	dirName := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	return filepath.Join(bs.basePath, "_global_species", dirName, "songs", songType+".mp3")
}

// GetNarrationPath returns the full path to a bird's narration file
func (bs *BirdStorage) GetNarrationPath(birdName, narrationType string) string {
	dirName := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	return filepath.Join(bs.basePath, "_global_species", dirName, "narration", narrationType+".mp3")
}

// GetIconPath returns the full path to a bird's icon
func (bs *BirdStorage) GetIconPath(birdName, iconType string) string {
	dirName := strings.ToLower(strings.ReplaceAll(birdName, " ", "_"))
	return filepath.Join(bs.basePath, "_global_species", dirName, "icons", iconType+".png")
}

// GetGlobalNarrationPath returns the path to global narration files
func (bs *BirdStorage) GetGlobalNarrationPath(timeOfDay, narType string, index int) string {
	filename := fmt.Sprintf("narration_%02d.mp3", index)
	return filepath.Join(bs.basePath, "global_narration", timeOfDay, narType, filename)
}

// ListBirdsByRegion returns all birds in a specific region
func (bs *BirdStorage) ListBirdsByRegion(region string) ([]string, error) {
	// This would read all metadata files and filter by region
	// For now, returning a placeholder implementation
	var birds []string

	// Read all birds from _global_species
	globalPath := filepath.Join(bs.basePath, "_global_species")
	entries, err := ioutil.ReadDir(globalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global species directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			// Check if this bird is in the requested region
			metadata, err := bs.GetBirdMetadata(entry.Name())
			if err != nil {
				continue
			}

			for _, r := range metadata.Regions {
				if r == region {
					birds = append(birds, metadata.CommonName)
					break
				}
			}
		}
	}

	return birds, nil
}

// GetBirdsByHabitat returns all birds that live in a specific habitat
func (bs *BirdStorage) GetBirdsByHabitat(habitat string) ([]string, error) {
	var birds []string

	globalPath := filepath.Join(bs.basePath, "_global_species")
	entries, err := ioutil.ReadDir(globalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global species directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			metadata, err := bs.GetBirdMetadata(entry.Name())
			if err != nil {
				continue
			}

			// Check primary habitat and habitats list
			if metadata.PrimaryHabitat == habitat {
				birds = append(birds, metadata.CommonName)
				continue
			}

			for _, h := range metadata.Habitats {
				if h == habitat {
					birds = append(birds, metadata.CommonName)
					break
				}
			}
		}
	}

	return birds, nil
}
