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

// WebResult represents a single web search result
type WebResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// WebSearchResponse represents the response from the Brave web search API
type WebSearchResponse struct {
	Web struct {
		Results []WebResult `json:"results"`
	} `json:"web"`
}

// WebSearch performs a web search using the Brave Search API
func WebSearch(
	apiKey string,
	query string,
	count int,
	offset int,
	rateLimiter *ratelimit.RateLimiter,
) (string, error) {
	// Check rate limits
	if err := rateLimiter.CheckLimit(); err != nil {
		return "", err
	}

	// Ensure count is within API limits
	if count <= 0 {
		count = 10 // Default value
	} else if count > 20 {
		count = 20 // API maximum
	}

	// Build the URL
	baseURL := "https://api.search.brave.com/res/v1/web/search"
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("q", query)
	q.Set("count", strconv.Itoa(count))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()

	// Create the request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip") // Explicitly accept gzip encoding
	req.Header.Set("X-Subscription-Token", apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Brave API error: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// Create a reader based on content encoding
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Parse the response
	var searchResp WebSearchResponse
	if err := json.NewDecoder(reader).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Format the results
	var results []string
	for _, result := range searchResp.Web.Results {
		formattedResult := fmt.Sprintf("Title: %s\nDescription: %s\nURL: %s",
			result.Title,
			result.Description,
			result.URL)
		results = append(results, formattedResult)
	}

	return strings.Join(results, "\n\n"), nil
}

// WebSearchTool defines the schema for the brave_web_search tool
var WebSearchTool = map[string]interface{}{
	"name": "brave_web_search",
	"description": "Performs a web search using the Brave Search API, ideal for general queries, news, articles, and online content. " +
		"Use this for broad information gathering, recent events, or when you need diverse web sources. " +
		"Supports pagination, content filtering, and freshness controls. " +
		"Maximum 20 results per request, with offset for pagination. ",
	"inputSchema": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query (max 400 chars, 50 words)",
			},
			"count": map[string]interface{}{
				"type":        "number",
				"description": "Number of results (1-20, default 10)",
				"default":     10,
			},
			"offset": map[string]interface{}{
				"type":        "number",
				"description": "Pagination offset (max 9, default 0)",
				"default":     0,
			},
		},
		"required": []string{"query"},
	},
}
