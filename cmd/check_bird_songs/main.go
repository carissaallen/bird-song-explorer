package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Map of bird directory names to their scientific names
var birdScientificNames = map[string]string{
	// North American birds
	"american-robin":          "Turdus migratorius",
	"northern-cardinal":       "Cardinalis cardinalis",
	"blue-jay":                "Cyanocitta cristata",
	"mourning-dove":           "Zenaida macroura",
	"black-capped-chickadee":  "Poecile atricapillus", // Adding this one

	// European birds
	"european-robin":     "Erithacus rubecula",
	"common-chaffinch":   "Fringilla coelebs", // Adding this one
	"eurasian-blue-tit":  "Cyanistes caeruleus", // Adding this one
	"great-tit":          "Parus major",

	// Asian birds
	"house-sparrow":          "Passer domesticus",
	"oriental-magpie-robin":  "Copsychus saularis", // Adding this one
	"japanese-white-eye":     "Zosterops japonicus", // Adding this one

	// Australian birds
	"australian-magpie": "Gymnorhina tibicen",
	"rainbow-lorikeet":  "Trichoglossus moluccanus",
	"kookaburra":        "Dacelo novaeguineae", // Adding common name

	// South American birds
	"great-kiskadee":  "Pitangus sulphuratus", // Adding this one
	"rufous-hornero":  "Furnarius rufus", // Adding this one
}

// XenoCantoResponse represents the API v2 response
type XenoCantoResponse struct {
	Recordings []XenoCantoRecording `json:"recordings"`
}

// XenoCantoRecording represents a recording from xeno-canto
type XenoCantoRecording struct {
	ID       string `json:"id"`
	Gen      string `json:"gen"`
	Sp       string `json:"sp"`
	En       string `json:"en"`
	Type     string `json:"type"`
	File     string `json:"file"`
	FileName string `json:"file-name"`
	Length   string `json:"length"`
	Quality  string `json:"q"`
}

// searchXenoCanto uses API v2 which doesn't require authentication
func searchXenoCanto(scientificName string) (*XenoCantoResponse, error) {
	parts := strings.Split(scientificName, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid scientific name: %s", scientificName)
	}

	// Build query for API v2
	baseURL := "https://www.xeno-canto.org/api/2/recordings"
	query := fmt.Sprintf("gen:%s sp:%s q:A", parts[0], parts[1])

	params := url.Values{}
	params.Set("query", query)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query xeno-canto: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("xeno-canto API error %d: %s", resp.StatusCode, string(body))
	}

	var result XenoCantoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func main() {
	// Check birds in the unavailable directory
	unavailableDir := "prerecorded_tts/bird-song-unavailable"

	// Read all birds in the unavailable directory
	birds, err := os.ReadDir(unavailableDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No bird-song-unavailable directory found")
			return
		}
		log.Fatalf("Failed to read unavailable directory: %v", err)
	}

	var birdsWithSongs []string
	var birdsWithoutSongs []string

	fmt.Println("Checking birds in unavailable directory against xeno-canto API v2...")
	fmt.Println("=" + strings.Repeat("=", 60))

	for _, bird := range birds {
		if !bird.IsDir() {
			continue
		}

		birdName := bird.Name()
		// Remove region suffix if present
		cleanBirdName := birdName
		if strings.Contains(birdName, "-europe") || strings.Contains(birdName, "-north-america") ||
		   strings.Contains(birdName, "-south-america") {
			parts := strings.Split(birdName, "-")
			if len(parts) >= 2 {
				// Rejoin all parts except the last one if it's a region
				lastPart := parts[len(parts)-1]
				if lastPart == "europe" || lastPart == "america" {
					cleanBirdName = strings.Join(parts[:len(parts)-1], "-")
				} else if len(parts) >= 3 && parts[len(parts)-2] == "north" || parts[len(parts)-2] == "south" {
					cleanBirdName = strings.Join(parts[:len(parts)-2], "-")
				}
			}
		}

		birdPath := filepath.Join(unavailableDir, birdName)

		// Get scientific name
		scientificName, exists := birdScientificNames[cleanBirdName]
		if !exists {
			fmt.Printf("❌ %s - No scientific name mapping found\n", birdName)
			birdsWithoutSongs = append(birdsWithoutSongs, birdPath)
			continue
		}

		// Try to get recording from xeno-canto
		fmt.Printf("Checking %s (%s)... ", birdName, scientificName)

		response, err := searchXenoCanto(scientificName)
		if err != nil {
			fmt.Printf("❌ API Error\n")
			fmt.Printf("   Error: %v\n", err)
			birdsWithoutSongs = append(birdsWithoutSongs, birdPath)
		} else if len(response.Recordings) == 0 {
			fmt.Printf("❌ No recordings found\n")
			birdsWithoutSongs = append(birdsWithoutSongs, birdPath)
		} else {
			recording := response.Recordings[0]
			// Fix URL if needed
			if recording.File != "" && strings.HasPrefix(recording.File, "//") {
				recording.File = "https:" + recording.File
			}
			fmt.Printf("✅ Found %d recordings\n", len(response.Recordings))
			fmt.Printf("   First recording: %s\n", recording.File)
			fmt.Printf("   Type: %s, Length: %s, Quality: %s\n", recording.Type, recording.Length, recording.Quality)
			birdsWithSongs = append(birdsWithSongs, birdPath)
		}
	}

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("SUMMARY: %d birds with songs, %d birds without songs\n",
		len(birdsWithSongs), len(birdsWithoutSongs))

	if len(birdsWithoutSongs) > 0 {
		fmt.Println("\nBirds WITHOUT available songs:")
		for _, bird := range birdsWithoutSongs {
			fmt.Printf("  - %s\n", filepath.Base(bird))
		}
	}

	if len(birdsWithSongs) > 0 {
		fmt.Println("\nBirds WITH available songs (can be moved back to audio):")
		for _, bird := range birdsWithSongs {
			fmt.Printf("  - %s\n", filepath.Base(bird))
		}
	}
}