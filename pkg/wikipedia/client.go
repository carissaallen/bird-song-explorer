package wikipedia

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	englishURL string
}

type PageSummary struct {
	Title        string `json:"title"`
	DisplayTitle string `json:"displaytitle"`
	Extract      string `json:"extract"`
	Description  string `json:"description"`
	ContentURLs  struct {
		Desktop struct {
			Page string `json:"page"`
		} `json:"desktop"`
	} `json:"content_urls"`
}

type ComprehensiveSummary struct {
	BirdName       string
	SimpleExtract  string
	EnglishExtract string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		// Using both Simple and English Wikipedia for comprehensive content
		baseURL:    "https://simple.wikipedia.org/api/rest_v1",
		englishURL: "https://en.wikipedia.org/api/rest_v1",
	}
}

func (c *Client) GetBirdSummary(birdName string) (*PageSummary, error) {
	// Try Simple Wikipedia first for kid-friendly content
	summary, err := c.fetchSummary(birdName, c.baseURL)
	if err == nil && summary != nil && summary.Extract != "" {
		return summary, nil
	}

	// Fall back to English Wikipedia for more detailed content
	return c.fetchSummary(birdName, c.englishURL)
}

// GetComprehensiveSummary fetches from both Simple and English Wikipedia
func (c *Client) GetComprehensiveSummary(birdName string) (*ComprehensiveSummary, error) {
	result := &ComprehensiveSummary{
		BirdName: birdName,
	}

	// Get Simple Wikipedia content (kid-friendly)
	if simple, err := c.fetchSummary(birdName, c.baseURL); err == nil && simple != nil {
		result.SimpleExtract = simple.Extract
	}

	// Get English Wikipedia content (detailed)
	if english, err := c.fetchSummary(birdName, c.englishURL); err == nil && english != nil {
		result.EnglishExtract = english.Extract
	}

	if result.SimpleExtract == "" && result.EnglishExtract == "" {
		return nil, fmt.Errorf("no Wikipedia content found for %s", birdName)
	}

	return result, nil
}

func (c *Client) fetchSummary(birdName string, baseURL string) (*PageSummary, error) {
	encodedName := url.QueryEscape(strings.ReplaceAll(birdName, " ", "_"))
	apiURL := fmt.Sprintf("%s/page/summary/%s", baseURL, encodedName)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "BirdSongExplorer/1.0 (https://github.com/callen/bird-song-explorer)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Wikipedia summary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// Try with underscores replaced by spaces in URL
		encodedName = url.QueryEscape(strings.ReplaceAll(strings.ToLower(birdName), " ", "_"))
		apiURL = fmt.Sprintf("%s/page/summary/%s", baseURL, encodedName)

		req, err = http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", "BirdSongExplorer/1.0 (https://github.com/callen/bird-song-explorer)")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Wikipedia summary: %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Wikipedia API returned status %d", resp.StatusCode)
	}

	var summary PageSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, fmt.Errorf("failed to decode Wikipedia response: %w", err)
	}

	return &summary, nil
}

func (c *Client) FormatForKids(summary *PageSummary, birdName string) string {
	if summary == nil || summary.Extract == "" {
		return fmt.Sprintf("The %s is an amazing bird! Scientists and bird watchers love studying this species to learn more about how birds live in nature.", birdName)
	}

	extract := summary.Extract
	sentences := strings.Split(extract, ". ")

	var kidFriendlySentences []string
	maxSentences := 5 // Increased from 3 to 5 for longer content

	for _, sentence := range sentences {
		if len(kidFriendlySentences) >= maxSentences {
			break
		}

		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Skip overly technical sentences
		if !strings.Contains(strings.ToLower(sentence), "genus") &&
			!strings.Contains(strings.ToLower(sentence), "taxonomy") &&
			!strings.Contains(strings.ToLower(sentence), "subspecies") &&
			!strings.Contains(strings.ToLower(sentence), "binomial") &&
			!strings.Contains(strings.ToLower(sentence), "phylogen") &&
			len(sentence) < 250 {
			if !strings.HasSuffix(sentence, ".") {
				sentence += "."
			}
			kidFriendlySentences = append(kidFriendlySentences, sentence)
		}
	}

	if len(kidFriendlySentences) == 0 {
		return fmt.Sprintf("The %s is an amazing bird! Scientists and bird watchers love studying this species to learn more about how birds live in nature.", birdName)
	}

	result := strings.Join(kidFriendlySentences, " ")

	// Make language more kid-friendly
	result = strings.ReplaceAll(result, " is a species of bird", " is a type of bird")
	result = strings.ReplaceAll(result, " are a species of bird", " are a type of bird")
	result = strings.ReplaceAll(result, "It is found", "You can find them")
	result = strings.ReplaceAll(result, "They are found", "You can find them")
	result = strings.ReplaceAll(result, "It inhabits", "They live in")
	result = strings.ReplaceAll(result, "They inhabit", "They live in")
	result = strings.ReplaceAll(result, "endemic to", "only found in")

	return result
}
