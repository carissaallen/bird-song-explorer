package ebird

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const baseURL = "https://api.ebird.org/v2"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type Observation struct {
	SpeciesCode    string  `json:"speciesCode"`
	CommonName     string  `json:"comName"`
	ScientificName string  `json:"sciName"`
	LocationName   string  `json:"locName"`
	ObsDate        string  `json:"obsDt"`
	HowMany        int     `json:"howMany"`
	Latitude       float64 `json:"lat"`
	Longitude      float64 `json:"lng"`
}

type Species struct {
	SpeciesCode    string `json:"speciesCode"`
	CommonName     string `json:"comName"`
	ScientificName string `json:"sciName"`
	Family         string `json:"familyComName"`
	Order          string `json:"order"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *Client) GetRecentObservations(lat, lng float64, days int) ([]Observation, error) {
	endpoint := fmt.Sprintf("%s/data/obs/geo/recent", baseURL)

	params := url.Values{}
	params.Add("lat", fmt.Sprintf("%.4f", lat))
	params.Add("lng", fmt.Sprintf("%.4f", lng))
	params.Add("dist", "50")
	params.Add("back", fmt.Sprintf("%d", days))
	params.Add("maxResults", "100")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-eBirdApiToken", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eBird API error: %d", resp.StatusCode)
	}

	var observations []Observation
	if err := json.NewDecoder(resp.Body).Decode(&observations); err != nil {
		return nil, err
	}

	return observations, nil
}

func (c *Client) GetSpeciesInfo(speciesCode string) (*Species, error) {
	endpoint := fmt.Sprintf("%s/ref/taxonomy/ebird", baseURL)

	params := url.Values{}
	params.Add("species", speciesCode)
	params.Add("fmt", "json")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-eBirdApiToken", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eBird API error: %d", resp.StatusCode)
	}

	var species []Species
	if err := json.NewDecoder(resp.Body).Decode(&species); err != nil {
		return nil, err
	}

	if len(species) > 0 {
		return &species[0], nil
	}

	return nil, fmt.Errorf("species not found")
}
