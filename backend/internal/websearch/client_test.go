package websearch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_SerpAPI(t *testing.T) {
	client, err := NewClient("serpapi", "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, &SerpAPIClient{}, client)
}

func TestNewClient_Exa(t *testing.T) {
	client, err := NewClient("exa", "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, &ExaClient{}, client)
}

func TestNewClient_Unknown(t *testing.T) {
	client, err := NewClient("unknown", "test-key")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "unknown web search provider")
}

func TestSerpAPIClient_Search_EmptyAPIKey(t *testing.T) {
	client := NewSerpAPIClient("")
	results, err := client.Search(context.Background(), "test query")
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "serpapi api key is empty")
}

func TestExaClient_Search_EmptyAPIKey(t *testing.T) {
	client := NewExaClient("")
	results, err := client.Search(context.Background(), "test query")
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "exa api key is empty")
}

func TestNewSerpAPIClient(t *testing.T) {
	client := NewSerpAPIClient("test-key")
	assert.NotNil(t, client)
	assert.Equal(t, "test-key", client.apiKey)
	assert.NotNil(t, client.client)
}

func TestNewExaClient(t *testing.T) {
	client := NewExaClient("test-key")
	assert.NotNil(t, client)
	assert.Equal(t, "test-key", client.apiKey)
	assert.NotNil(t, client.client)
}
