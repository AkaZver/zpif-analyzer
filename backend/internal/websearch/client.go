package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SearchResult struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

type Client interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
}

type SerpAPIClient struct {
	apiKey string
	client *http.Client
}

func NewSerpAPIClient(apiKey string) *SerpAPIClient {
	return &SerpAPIClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *SerpAPIClient) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("serpapi api key is empty")
	}

	endpoint := "https://serpapi.com/search.json?" + url.Values{
		"q":        {query},
		"api_key":  {c.apiKey},
		"engine":   {"google"},
		"num":      {"10"},
		"gl":       {"ru"},
		"hl":       {"ru"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("serpapi error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var data struct {
		OrganicResults []struct {
			Link    string `json:"link"`
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode serpapi response: %w", err)
	}

	results := make([]SearchResult, 0, len(data.OrganicResults))
	for _, r := range data.OrganicResults {
		if r.Link == "" {
			continue
		}
		results = append(results, SearchResult{
			URL:     r.Link,
			Title:   r.Title,
			Snippet: r.Snippet,
		})
	}

	return results, nil
}

type ExaClient struct {
	apiKey string
	client *http.Client
}

func NewExaClient(apiKey string) *ExaClient {
	return &ExaClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *ExaClient) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("exa api key is empty")
	}

	body, _ := json.Marshal(map[string]interface{}{
		"query":        query,
		"numResults":   10,
		"startPublishedDate": "2020-01-01",
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.exa.ai/search", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("exa error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var data struct {
		Results []struct {
			URL     string `json:"url"`
			Title   string `json:"title"`
			Text    string `json:"text"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode exa response: %w", err)
	}

	results := make([]SearchResult, 0, len(data.Results))
	for _, r := range data.Results {
		if r.URL == "" {
			continue
		}
		snippet := r.Text
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		results = append(results, SearchResult{
			URL:     r.URL,
			Title:   r.Title,
			Snippet: snippet,
		})
	}

	return results, nil
}

func NewClient(provider string, apiKey string) (Client, error) {
	switch strings.ToLower(provider) {
	case "serpapi":
		return NewSerpAPIClient(apiKey), nil
	case "exa":
		return NewExaClient(apiKey), nil
	default:
		return nil, fmt.Errorf("unknown web search provider: %s", provider)
	}
}