package services

import (
	"math/rand"
	"strings"
	"time"

	"github.com/callen/bird-song-explorer/internal/models"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type AvailableBird struct {
	CommonName     string
	ScientificName string
	Region         string
	Regions        []string
}

type AvailableBirdsService struct {
	birds []AvailableBird
}

func NewAvailableBirdsService() *AvailableBirdsService {
	birds := []AvailableBird{
		{
			CommonName:     "Western Meadowlark",
			ScientificName: "Sturnella neglecta",
			Region:         "north_america",
			Regions:        []string{"north_america", "us", "canada", "mexico", "global"},
		},
		{
			CommonName:     "Atlantic Puffin",
			ScientificName: "Fratercula arctica",
			Region:         "global",
			Regions:        []string{"north_america", "us", "canada", "europe", "iceland", "norway", "uk", "global"},
		},
		{
			CommonName:     "Great Spotted Woodpecker",
			ScientificName: "Dendrocopos major",
			Region:         "europe",
			Regions:        []string{"europe", "uk", "germany", "france", "spain", "russia", "china", "japan", "global"},
		},
		{
			CommonName:     "Brown Kiwi",
			ScientificName: "Apteryx mantelli",
			Region:         "oceania",
			Regions:        []string{"oceania", "new_zealand", "australia", "global"},
		},
		//{
		//	CommonName:     "Bald Eagle",
		//	ScientificName: "Haliaeetus leucocephalus",
		//	Region:         "north_america",
		//	Regions:        []string{"north_america", "us", "canada", "global"},
		//},
		//{
		//	CommonName:     "Common Kingfisher",
		//	ScientificName: "Alcedo atthis",
		//	Region:         "europe",
		//	Regions:        []string{"europe", "asia", "uk", "germany", "france", "spain", "russia", "china", "japan", "global"},
		//},
		//{
		//	CommonName:     "Laughing Kookaburra",
		//	ScientificName: "Dacelo novaeguineae",
		//	Region:         "oceania",
		//	Regions:        []string{"oceania", "australia", "global"},
		//},
	}

	return &AvailableBirdsService{
		birds: birds,
	}
}

func (s *AvailableBirdsService) GetAllAvailableBirds() []AvailableBird {
	return s.birds
}

func (s *AvailableBirdsService) GetBirdsByRegion(region string) []AvailableBird {
	var regionalBirds []AvailableBird
	regionLower := strings.ToLower(region)

	for _, bird := range s.birds {
		for _, birdRegion := range bird.Regions {
			if strings.ToLower(birdRegion) == regionLower {
				regionalBirds = append(regionalBirds, bird)
				break
			}
		}
	}

	return regionalBirds
}

func (s *AvailableBirdsService) GetRandomBird() *models.Bird {
	if len(s.birds) == 0 {
		return nil
	}

	selected := s.birds[rand.Intn(len(s.birds))]

	return &models.Bird{
		CommonName:     selected.CommonName,
		ScientificName: selected.ScientificName,
		Region:         selected.Region,
	}
}

func (s *AvailableBirdsService) GetRandomBirdForLocation(location *models.Location) *models.Bird {
	var matchingBirds []AvailableBird

	if location != nil && location.Country != "" {
		regionLower := strings.ToLower(location.Country)
		for _, bird := range s.birds {
			for _, region := range bird.Regions {
				if strings.ToLower(region) == regionLower {
					matchingBirds = append(matchingBirds, bird)
					break
				}
			}
		}
	}

	if len(matchingBirds) == 0 {
		matchingBirds = s.birds
	}

	if len(matchingBirds) == 0 {
		return &models.Bird{
			CommonName:     "Western Meadowlark",
			ScientificName: "Sturnella neglecta",
			Region:         "north_america",
		}
	}

	selected := matchingBirds[rand.Intn(len(matchingBirds))]

	return &models.Bird{
		CommonName:     selected.CommonName,
		ScientificName: selected.ScientificName,
		Region:         selected.Region,
	}
}

func (s *AvailableBirdsService) HasAvailableBirds() bool {
	return len(s.birds) > 0
}

func (s *AvailableBirdsService) GetCyclingBird() *models.Bird {
	if len(s.birds) == 0 {
		return nil
	}

	now := time.Now()
	daySeed := now.Year()*365 + now.YearDay()
	birdIndex := daySeed % len(s.birds)

	selected := s.birds[birdIndex]

	return &models.Bird{
		CommonName:     selected.CommonName,
		ScientificName: selected.ScientificName,
		Region:         selected.Region,
	}
}
