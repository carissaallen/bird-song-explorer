package yoto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// IconSearcher handles searching for icons from various sources
type IconSearcher struct {
	client      *Client
	cache       map[string]*IconSearchResult
	cacheMu     sync.RWMutex
	rateLimiter *RateLimiter
}

// IconSearchResult represents an icon found through search
type IconSearchResult struct {
	MediaID  string    `json:"mediaId"`
	Title    string    `json:"title"`
	URL      string    `json:"url"`
	Source   string    `json:"source"` // "yoto-public" or "yotoicons"
	Author   string    `json:"author,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	CachedAt time.Time `json:"cachedAt"`
}

// YotoPublicIcon represents an icon from Yoto's public library
type YotoPublicIcon struct {
	MediaID    string   `json:"mediaId"`
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	PublicTags []string `json:"publicTags,omitempty"`
}

// RateLimiter controls request frequency
type RateLimiter struct {
	lastRequest time.Time
	minInterval time.Duration
	mu          sync.Mutex
}

func NewIconSearcher(client *Client) *IconSearcher {
	return &IconSearcher{
		client: client,
		cache:  make(map[string]*IconSearchResult),
		rateLimiter: &RateLimiter{
			minInterval: 1 * time.Second,
		},
	}
}

// SearchBirdIcon searches for an icon matching the bird name
func (is *IconSearcher) SearchBirdIcon(birdName string) (string, error) {

	// Check cache first
	is.cacheMu.RLock()
	if cached, exists := is.cache[birdName]; exists && time.Since(cached.CachedAt) < 24*time.Hour {
		is.cacheMu.RUnlock()
		return FormatIconID(cached.MediaID), nil
	}
	is.cacheMu.RUnlock()

	// Generate variations first (more generic terms)
	variations := is.generateBirdNameVariations(birdName)

	// Try variations first (they're more likely to have icons)
	for _, variation := range variations {
		fmt.Printf("Searching yotoicons.com for variation: %s...\n", variation)
		yotoiconsResult, err := is.searchYotoicons(variation)
		if err == nil && yotoiconsResult != nil {
			// Upload the icon from yotoicons.com to Yoto
			mediaID, err := is.uploadYotoiconsIcon(yotoiconsResult)
			if err == nil && mediaID != "" {
				result := &IconSearchResult{
					MediaID:  mediaID,
					Title:    yotoiconsResult.Title,
					Source:   "yotoicons",
					Author:   yotoiconsResult.Author,
					CachedAt: time.Now(),
				}

				// Cache with original name
				is.cacheMu.Lock()
				is.cache[birdName] = result
				is.cacheMu.Unlock()

				fmt.Printf("Found and uploaded icon from yotoicons for %s (variation: %s): %s\n", birdName, variation, mediaID)
				return FormatIconID(mediaID), nil
			}
		}
	}

	// If variations didn't work, try the full bird name as a last resort
	fmt.Printf("Searching yotoicons.com for full name: %s\n", birdName)
	yotoiconsResult, err := is.searchYotoicons(birdName)
	if err == nil && yotoiconsResult != nil {
		// Upload the icon from yotoicons.com to Yoto
		mediaID, err := is.uploadYotoiconsIcon(yotoiconsResult)
		if err == nil && mediaID != "" {
			result := &IconSearchResult{
				MediaID:  mediaID,
				Title:    yotoiconsResult.Title,
				Source:   "yotoicons",
				Author:   yotoiconsResult.Author,
				CachedAt: time.Now(),
			}

			// Cache the result
			is.cacheMu.Lock()
			is.cache[birdName] = result
			is.cacheMu.Unlock()

			fmt.Printf("Found icon for %s using full name\n", birdName)
			return FormatIconID(mediaID), nil
		}
	}

	// We no longer search Yoto public icons to avoid generic "bird" matches
	// Only use specific matches from yotoicons.com
	fmt.Printf("No specific icon found for %s on yotoicons.com, will use meadowlark default\n", birdName)
	return "", nil
}

// searchYotoPublicIcons searches Yoto's public icon library
func (is *IconSearcher) searchYotoPublicIcons(query string) (*IconSearchResult, error) {
	// Ensure authenticated
	if err := is.client.ensureAuthenticated(); err != nil {
		return nil, err
	}

	// Get public icons from Yoto
	url := fmt.Sprintf("%s/media/displayIcons/user/yoto", is.client.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+is.client.accessToken)

	resp, err := is.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get public icons: %d", resp.StatusCode)
	}

	var icons []YotoPublicIcon
	if err := json.NewDecoder(resp.Body).Decode(&icons); err != nil {
		return nil, err
	}

	// Search for matching icons
	lowerQuery := strings.ToLower(query)
	for _, icon := range icons {
		// Check title and tags
		allText := strings.ToLower(icon.Title)
		for _, tag := range icon.PublicTags {
			allText += " " + strings.ToLower(tag)
		}

		if strings.Contains(allText, lowerQuery) {
			return &IconSearchResult{
				MediaID:  icon.MediaID,
				Title:    icon.Title,
				URL:      icon.URL,
				Source:   "yoto-public",
				Tags:     icon.PublicTags,
				CachedAt: time.Now(),
			}, nil
		}
	}

	// Try partial matches on individual words
	words := strings.Fields(lowerQuery)
	for _, icon := range icons {
		allText := strings.ToLower(icon.Title)
		for _, tag := range icon.PublicTags {
			allText += " " + strings.ToLower(tag)
		}

		for _, word := range words {
			if len(word) > 3 && strings.Contains(allText, word) {
				return &IconSearchResult{
					MediaID:  icon.MediaID,
					Title:    icon.Title,
					URL:      icon.URL,
					Source:   "yoto-public",
					Tags:     icon.PublicTags,
					CachedAt: time.Now(),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no matching icon found")
}

// searchYotoicons searches yotoicons.com for matching icons
func (is *IconSearcher) searchYotoicons(query string) (*IconSearchResult, error) {
	// Rate limiting
	is.rateLimiter.Wait()

	searchURL := fmt.Sprintf("https://www.yotoicons.com/icons?tag=%s", url.QueryEscape(query))

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yotoicons search failed: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	lowerHTML := strings.ToLower(html)

	// Parse HTML to find icon images
	iconRegex := regexp.MustCompile(`<img[^>]+src=["']/static/uploads/(\d+)\.png["'][^>]*>`)
	matches := iconRegex.FindAllStringSubmatch(html, 10)

	// Check if we're on a "no results" page
	if strings.Contains(html, "No icons found") || strings.Contains(html, "no results") {
		fmt.Printf("No results found on yotoicons.com for: %s\n", query)
		return nil, fmt.Errorf("no icons found on yotoicons")
	}

	if len(matches) > 0 {
		// Check if we found the search term on the page
		lowerQuery := strings.ToLower(query)
		hasSearchTerm := strings.Contains(lowerHTML, lowerQuery)

		// Accept the result if we found the search term
		// We're being less strict now - if searching for "duck" finds a duck icon, that's good enough
		if hasSearchTerm {
			fmt.Printf("Found icon for %s on yotoicons.com\n", query)
		} else {
			// Log if we're getting results but not for our search term
			fmt.Printf("Warning: Search for %s returned results but search term not found on page\n", query)
		}

		// Avoid truly generic results only when we have no bird association
		if strings.Contains(lowerHTML, "generic") && !hasSearchTerm {
			fmt.Printf("Found only generic icon for %s, skipping\n", query)
			return nil, fmt.Errorf("only generic icon found")
		}

		// Use the first match
		iconID := matches[0][1]
		iconURL := fmt.Sprintf("https://www.yotoicons.com/static/uploads/%s.png", iconID)

		// Try to extract author
		authorRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
		authorMatch := authorRegex.FindStringSubmatch(html)
		author := "unknown"
		if len(authorMatch) > 1 {
			author = authorMatch[1]
		}

		return &IconSearchResult{
			MediaID:  iconID, // Temporary, will be replaced after upload
			Title:    fmt.Sprintf("%s icon", query),
			URL:      iconURL,
			Source:   "yotoicons",
			Author:   author,
			CachedAt: time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("no icons found on yotoicons")
}

// uploadYotoiconsIcon downloads and uploads an icon from yotoicons.com
func (is *IconSearcher) uploadYotoiconsIcon(icon *IconSearchResult) (string, error) {
	// Download the icon
	resp, err := http.Get(icon.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	iconData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Upload to Yoto using the icon uploader
	uploader := NewIconUploader(is.client)

	// Save temporarily and upload
	tempFile := fmt.Sprintf("/tmp/yotoicon_%s.png", icon.MediaID)
	if err := os.WriteFile(tempFile, iconData, 0644); err != nil {
		return "", err
	}
	defer os.Remove(tempFile)

	mediaID, err := uploader.UploadIcon(tempFile, fmt.Sprintf("bird_%s", icon.Title))
	if err != nil {
		return "", err
	}

	return mediaID, nil
}

// generateBirdNameVariations creates search variations for bird names
func (is *IconSearcher) generateBirdNameVariations(birdName string) []string {
	variations := []string{}

	// Clean up the name
	cleanName := strings.ReplaceAll(birdName, "'s", "")
	cleanName = strings.ReplaceAll(cleanName, "-", " ")

	words := strings.Fields(cleanName)

	// Try the last word if it's a bird type (e.g., "Blue Jay" -> "Jay")
	if len(words) > 1 {
		lastWord := words[len(words)-1]
		if isDistinctiveBirdType(lastWord) {
			variations = append(variations, lastWord)
		}

		// Try first word if descriptive (e.g., "Bald Eagle" -> "Eagle")
		if len(words) == 2 && isDistinctiveBirdType(words[1]) {
			variations = append(variations, words[1])
		}

		// Try the first word alone (e.g., "Cardinal" from "Northern Cardinal")
		if isDistinctiveBirdType(words[0]) {
			variations = append(variations, words[0])
		}
	}

	// Extract the main bird type from compound names
	// Use word boundaries to avoid false matches like "owl" in "Meadowlark"
	lowerBirdName := strings.ToLower(birdName)
	for _, birdType := range getCommonBirdTypes() {
		// Check for whole word match, not substring
		for _, word := range strings.Fields(lowerBirdName) {
			if word == birdType {
				variations = append(variations, birdType)
				break
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	unique := []string{}
	for _, v := range variations {
		lower := strings.ToLower(v)
		if !seen[lower] {
			seen[lower] = true
			unique = append(unique, v)
		}
	}

	return unique
}

// isDistinctiveBirdType checks if a word is a distinctive bird type
func isDistinctiveBirdType(word string) bool {
	types := getCommonBirdTypes()
	lower := strings.ToLower(word)
	for _, t := range types {
		if lower == t {
			return true
		}
	}
	return false
}

// getCommonBirdTypes returns common bird type names
func getCommonBirdTypes() []string {
	return []string{
		"eagle", "hawk", "owl", "duck", "goose", "swan", "crow", "raven",
		"sparrow", "robin", "cardinal", "heron", "crane", "pelican", "penguin", "parrot",
		"jay", "finch", "warbler", "thrush", "wren", "chickadee", "nuthatch", "woodpecker",
		"flycatcher", "vireo", "tanager", "bunting", "grosbeak", "blackbird", "oriole",
		"dove", "pigeon", "quail", "grouse", "turkey", "pheasant", "partridge",
		"gull", "tern", "sandpiper", "plover", "cormorant", "grebe", "loon",
		"ibis", "stork", "flamingo", "spoonbill", "egret", "bittern",
		"kestrel", "falcon", "vulture", "condor", "osprey", "harrier",
		"kingfisher", "bee-eater", "roller", "hoopoe", "cuckoo", "roadrunner",
		"swift", "hummingbird", "trogon", "toucan", "hornbill", "puffin",
		// Added common North American birds that were missing
		"mockingbird", "meadowlark", "bluebird", "catbird", "thrasher", "towhee",
		"junco", "siskin", "goldfinch", "redpoll", "crossbill", "starling",
		"cowbird", "bobolink", "shrike", "magpie", "nutcracker", "dipper",
		"wagtail", "pipit", "lark", "swallow", "martin", "bushtit",
		"kinglet", "gnatcatcher", "creeper", "waxwing", "phalarope", "yellowlegs",
		"dowitcher", "snipe", "godwit", "curlew", "whimbrel", "turnstone",
		"knot", "dunlin", "sanderling", "stilt", "avocet", "oystercatcher",
		"skimmer", "murre", "guillemot", "auklet", "murrelet", "razorbill",
		"jaeger", "skua", "kittiwake", "fulmar", "shearwater", "petrel",
		"gannet", "booby", "tropicbird", "frigatebird", "anhinga", "rail",
		"coot", "moorhen", "gallinule", "sora", "bittern", "night-heron",
	}
}

// Wait implements rate limiting
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	timeSince := time.Since(rl.lastRequest)
	if timeSince < rl.minInterval {
		time.Sleep(rl.minInterval - timeSince)
	}
	rl.lastRequest = time.Now()
}
