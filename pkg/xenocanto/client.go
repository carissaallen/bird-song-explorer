package xenocanto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const baseURL = "https://xeno-canto.org/api/3"

type Client struct {
	httpClient *http.Client
	apiKey     string
}

type SearchResponse struct {
	NumRecordings string      `json:"numRecordings"`
	NumSpecies    string      `json:"numSpecies"`
	Page          int         `json:"page"`
	NumPages      int         `json:"numPages"`
	Recordings    []Recording `json:"recordings"`
}

type Recording struct {
	ID          string `json:"id"`
	Gen         string `json:"gen"`
	Sp          string `json:"sp"`
	En          string `json:"en"`
	Rec         string `json:"rec"`
	Cnt         string `json:"cnt"`
	Loc         string `json:"loc"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
	Type        string `json:"type"`
	File        string `json:"file"`
	FileName    string `json:"file-name"`
	Length      string `json:"length"`
	Time        string `json:"time"`
	Date        string `json:"date"`
	Quality     string `json:"q"`
	URL         string `json:"url"`
	License     string `json:"lic"`
	Attribution string
}

func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{},
		apiKey:     apiKey,
	}
}

func (c *Client) SearchRecordings(scientificName string, quality string) (*SearchResponse, error) {
	// Split scientific name into genus and species
	parts := strings.Split(scientificName, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid scientific name format: %s", scientificName)
	}

	// Build query using gen: and sp: tags
	searchQuery := fmt.Sprintf("gen:%s sp:%s", parts[0], parts[1])
	if quality != "" {
		searchQuery = fmt.Sprintf("%s q:%s", searchQuery, quality)
	}

	params := url.Values{}
	params.Add("query", searchQuery)
	// Xeno-canto API v3 requires an API key
	if c.apiKey != "" {
		params.Add("key", c.apiKey)
	}

	endpoint := fmt.Sprintf("%s/recordings?%s", baseURL, params.Encode())

	fmt.Printf("Xeno-canto API request: %s\n", endpoint)

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Xeno-canto API error: %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	for i := range result.Recordings {
		rec := &result.Recordings[i]
		rec.Attribution = fmt.Sprintf("%s, XC%s, %s, %s",
			rec.Rec, rec.ID, rec.License, rec.URL)

		if rec.File != "" && strings.HasPrefix(rec.File, "//") {
			rec.File = "https:" + rec.File
		}
	}

	return &result, nil
}

func (c *Client) GetBestRecording(scientificName string) (*Recording, error) {
	searchResp, err := c.SearchRecordings(scientificName, "A")
	if err != nil {
		return nil, err
	}

	if len(searchResp.Recordings) == 0 {
		searchResp, err = c.SearchRecordings(scientificName, "")
		if err != nil {
			return nil, err
		}
	}

	if len(searchResp.Recordings) == 0 {
		return nil, fmt.Errorf("no recordings found for %s", scientificName)
	}

	for _, rec := range searchResp.Recordings {
		if rec.Type == "song" || rec.Type == "call" {
			duration := c.parseDuration(rec.Length)
			if duration >= 15 && duration <= 60 {
				return &rec, nil
			}
		}
	}

	return &searchResp.Recordings[0], nil
}

func (c *Client) parseDuration(length string) int {
	parts := strings.Split(length, ":")
	if len(parts) != 2 {
		return 0
	}

	var minutes, seconds int
	fmt.Sscanf(parts[0], "%d", &minutes)
	fmt.Sscanf(parts[1], "%d", &seconds)

	return minutes*60 + seconds
}
