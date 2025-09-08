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

type Hotspot struct {
	LocationID   string  `json:"locId"`
	LocationName string  `json:"locName"`
	CountryCode  string  `json:"countryCode"`
	SubNational1 string  `json:"subnational1Code"`
	SubNational2 string  `json:"subnational2Code"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lng"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *Client) GetRecentObservations(lat, lng float64, days int) ([]Observation, error) {
	return c.GetRecentObservationsWithRadius(lat, lng, 50, days)
}

// GetRecentObservationsWithRadius gets recent bird observations within a specified radius
func (c *Client) GetRecentObservationsWithRadius(lat, lng float64, radiusKm, days int) ([]Observation, error) {
	endpoint := fmt.Sprintf("%s/data/obs/geo/recent", baseURL)

	params := url.Values{}
	params.Add("lat", fmt.Sprintf("%.4f", lat))
	params.Add("lng", fmt.Sprintf("%.4f", lng))
	params.Add("dist", fmt.Sprintf("%d", radiusKm))
	params.Add("back", fmt.Sprintf("%d", days))
	params.Add("maxResults", "200")  // Increase for wider searches

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

func (c *Client) GetNearbyHotspots(lat, lng float64, dist int) ([]Hotspot, error) {
	endpoint := fmt.Sprintf("%s/ref/hotspot/geo", baseURL)

	params := url.Values{}
	params.Add("lat", fmt.Sprintf("%.4f", lat))
	params.Add("lng", fmt.Sprintf("%.4f", lng))
	params.Add("dist", fmt.Sprintf("%d", dist))
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

	var hotspots []Hotspot
	if err := json.NewDecoder(resp.Body).Decode(&hotspots); err != nil {
		return nil, err
	}

	return hotspots, nil
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
