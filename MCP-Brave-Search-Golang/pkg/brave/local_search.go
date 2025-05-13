package brave

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers/MCP-Brave-Search-Golang/internal/ratelimit"
)

// LocationResult represents a location result in the local search
type LocationResult struct {
	ID string `json:"id"`
}

// LocationSearchResponse represents the response from the location search
type LocationSearchResponse struct {
	Locations struct {
		Results []LocationResult `json:"results"`
	} `json:"locations"`
}

// Address represents an address in a POI result
type Address struct {
	StreetAddress   string `json:"streetAddress"`
	AddressLocality string `json:"addressLocality"`
	AddressRegion   string `json:"addressRegion"`
	PostalCode      string `json:"postalCode"`
}

// Rating represents a rating in a POI result
type Rating struct {
	RatingValue float64 `json:"ratingValue"`
	RatingCount int     `json:"ratingCount"`
}

// POI represents a point of interest result
type POI struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Address      Address  `json:"address"`
	Phone        string   `json:"phone"`
	Rating       Rating   `json:"rating"`
	PriceRange   string   `json:"priceRange"`
	OpeningHours []string `json:"openingHours"`
}

// POIsResponse represents the response from the POIs API
type POIsResponse struct {
	Results []POI `json:"results"`
}

// DescriptionsResponse represents the response from the descriptions API
type DescriptionsResponse struct {
	Descriptions map[string]string `json:"descriptions"`
}

// LocalSearch performs a local search using the Brave Search API
func LocalSearch(
	apiKey string,
	query string,
	count int,
	rateLimiter *ratelimit.RateLimiter,
) (string, error) {
	// Check rate limits
	if err := rateLimiter.CheckLimit(); err != nil {
		return "", err
	}

	// Ensure count is within API limits
	if count <= 0 {
		count = 5 // Default value
	} else if count > 20 {
		count = 20 // API maximum
	}

	// Step 1: Perform initial search to get location IDs
	locationIDs, err := getLocationIDs(apiKey, query, count, rateLimiter)
	if err != nil {
		return "", err
	}

	// If no locations found, fall back to web search
	if len(locationIDs) == 0 {
		return WebSearch(apiKey, query, count, 0, rateLimiter)
	}

	// Step 2: Get POIs and descriptions in parallel
	poisChan := make(chan POIsResponse)
	poisErrChan := make(chan error)
	descChan := make(chan DescriptionsResponse)
	descErrChan := make(chan error)

	go func() {
		pois, err := getPOIsData(apiKey, locationIDs, rateLimiter)
		if err != nil {
			poisErrChan <- err
			return
		}
		poisChan <- pois
	}()

	go func() {
		desc, err := getDescriptionsData(apiKey, locationIDs, rateLimiter)
		if err != nil {
			descErrChan <- err
			return
		}
		descChan <- desc
	}()

	// Wait for both goroutines to complete
	var poisResp POIsResponse
	var descResp DescriptionsResponse

	select {
	case poisResp = <-poisChan:
	case err := <-poisErrChan:
		return "", fmt.Errorf("failed to get POIs data: %w", err)
	}

	select {
	case descResp = <-descChan:
	case err := <-descErrChan:
		return "", fmt.Errorf("failed to get descriptions data: %w", err)
	}

	// Format the results
	return formatLocalResults(poisResp, descResp), nil
}

// getLocationIDs performs the initial search to get location IDs
func getLocationIDs(apiKey string, query string, count int, rateLimiter *ratelimit.RateLimiter) ([]string, error) {
	// Check rate limits
	if err := rateLimiter.CheckLimit(); err != nil {
		return nil, err
	}

	// Build the URL
	baseURL := "https://api.search.brave.com/res/v1/web/search"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("q", query)
	q.Set("search_lang", "en")
	q.Set("result_filter", "locations")
	q.Set("count", strconv.Itoa(count))
	u.RawQuery = q.Encode()

	// Create the request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Subscription-Token", apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Brave API error: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// Create a reader based on content encoding
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Parse the response
	var locationResp LocationSearchResponse
	if err := json.NewDecoder(reader).Decode(&locationResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract location IDs
	var locationIDs []string
	for _, location := range locationResp.Locations.Results {
		if location.ID != "" {
			locationIDs = append(locationIDs, location.ID)
		}
	}

	return locationIDs, nil
}

// getPOIsData gets POI details for the given location IDs
func getPOIsData(apiKey string, ids []string, rateLimiter *ratelimit.RateLimiter) (POIsResponse, error) {
	// Check rate limits
	if err := rateLimiter.CheckLimit(); err != nil {
		return POIsResponse{}, err
	}

	// Build the URL
	baseURL := "https://api.search.brave.com/res/v1/local/pois"
	u, err := url.Parse(baseURL)
	if err != nil {
		return POIsResponse{}, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters (multiple IDs)
	q := u.Query()
	for _, id := range ids {
		if id != "" {
			q.Add("ids", id)
		}
	}
	u.RawQuery = q.Encode()

	// Create the request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return POIsResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Subscription-Token", apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return POIsResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return POIsResponse{}, fmt.Errorf("Brave API error: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// Create a reader based on content encoding
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return POIsResponse{}, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Parse the response
	var poisResp POIsResponse
	if err := json.NewDecoder(reader).Decode(&poisResp); err != nil {
		return POIsResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return poisResp, nil
}

// getDescriptionsData gets descriptions for the given location IDs
func getDescriptionsData(apiKey string, ids []string, rateLimiter *ratelimit.RateLimiter) (DescriptionsResponse, error) {
	// Check rate limits
	if err := rateLimiter.CheckLimit(); err != nil {
		return DescriptionsResponse{}, err
	}

	// Build the URL
	baseURL := "https://api.search.brave.com/res/v1/local/descriptions"
	u, err := url.Parse(baseURL)
	if err != nil {
		return DescriptionsResponse{}, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters (multiple IDs)
	q := u.Query()
	for _, id := range ids {
		if id != "" {
			q.Add("ids", id)
		}
	}
	u.RawQuery = q.Encode()

	// Create the request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return DescriptionsResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Subscription-Token", apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return DescriptionsResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return DescriptionsResponse{}, fmt.Errorf("Brave API error: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// Create a reader based on content encoding
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return DescriptionsResponse{}, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Parse the response
	var descResp DescriptionsResponse
	if err := json.NewDecoder(reader).Decode(&descResp); err != nil {
		return DescriptionsResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return descResp, nil
}

// formatLocalResults formats the POIs and descriptions into a string
func formatLocalResults(poisResp POIsResponse, descResp DescriptionsResponse) string {
	if len(poisResp.Results) == 0 {
		return "No local results found"
	}

	var results []string
	for _, poi := range poisResp.Results {
		// Format address
		addressParts := []string{
			poi.Address.StreetAddress,
			poi.Address.AddressLocality,
			poi.Address.AddressRegion,
			poi.Address.PostalCode,
		}
		
		// Filter out empty parts
		var filteredAddressParts []string
		for _, part := range addressParts {
			if part != "" {
				filteredAddressParts = append(filteredAddressParts, part)
			}
		}
		
		address := "N/A"
		if len(filteredAddressParts) > 0 {
			address = strings.Join(filteredAddressParts, ", ")
		}

		// Format rating
		rating := "N/A"
		if poi.Rating.RatingValue > 0 {
			rating = fmt.Sprintf("%.1f (%d reviews)", poi.Rating.RatingValue, poi.Rating.RatingCount)
		}

		// Format hours
		hours := "N/A"
		if len(poi.OpeningHours) > 0 {
			hours = strings.Join(poi.OpeningHours, ", ")
		}

		// Format description
		description := "No description available"
		if desc, ok := descResp.Descriptions[poi.ID]; ok && desc != "" {
			description = desc
		}

		// Format result
		result := fmt.Sprintf("Name: %s\nAddress: %s\nPhone: %s\nRating: %s\nPrice Range: %s\nHours: %s\nDescription: %s",
			poi.Name,
			address,
			getNonEmptyString(poi.Phone, "N/A"),
			rating,
			getNonEmptyString(poi.PriceRange, "N/A"),
			hours,
			description)
		
		results = append(results, result)
	}

	return strings.Join(results, "\n---\n")
}

// getNonEmptyString returns the string if it's not empty, or the default value
func getNonEmptyString(s string, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

// LocalSearchTool defines the schema for the brave_local_search tool
var LocalSearchTool = map[string]interface{}{
	"name": "brave_local_search",
	"description": "Searches for local businesses and places using Brave's Local Search API. " +
		"Best for queries related to physical locations, businesses, restaurants, services, etc. " +
		"Returns detailed information including:\n" +
		"- Business names and addresses\n" +
		"- Ratings and review counts\n" +
		"- Phone numbers and opening hours\n" +
		"Use this when the query implies 'near me' or mentions specific locations. " +
		"Automatically falls back to web search if no local results are found.",
	"inputSchema": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Local search query (e.g. 'pizza near Central Park')",
			},
			"count": map[string]interface{}{
				"type":        "number",
				"description": "Number of results (1-20, default 5)",
				"default":     5,
			},
		},
		"required": []string{"query"},
	},
}
