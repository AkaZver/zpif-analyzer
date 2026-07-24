package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-key", "https://api.openai.com/v1", "gpt-4o-mini", nil)
	assert.NotNil(t, client)
	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
	assert.Equal(t, "gpt-4o-mini", client.model)
}

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient("test-key", "", "", nil)
	assert.NotNil(t, client)
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
	assert.Equal(t, "gpt-4o-mini", client.model)
}

func TestClient_Chat_EmptyAPIKey(t *testing.T) {
	client := NewClient("", "https://api.openai.com/v1", "gpt-4o-mini", nil)
	messages := []Message{{Role: "user", Content: "test"}}

	resp, err := client.Chat(context.Background(), messages)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "LLM api key is empty")
}

func TestClient_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		response := ChatResponse{
			ID: "chatcmpl-123",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	messages := []Message{{Role: "user", Content: "Hello"}}

	resp, err := client.Chat(context.Background(), messages)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "Hello! How can I help you?", resp.Choices[0].Message.Content)
}

func TestClient_Chat_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			ID:      "chatcmpl-123",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	messages := []Message{{Role: "user", Content: "Hello"}}

	resp, err := client.Chat(context.Background(), messages)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "LLM returned no choices")
}

func TestClient_ChatSimple_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Response content",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	resp, err := client.ChatSimple(context.Background(), "You are helpful", "Hello")

	assert.NoError(t, err)
	assert.Equal(t, "Response content", resp)
}

func TestClient_TestConnection_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "OK",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	err := client.TestConnection(context.Background())

	assert.NoError(t, err)
}

func TestClient_TestConnection_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	err := client.TestConnection(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response from LLM")
}

func TestClient_GetModel(t *testing.T) {
	client := NewClient("test-key", "", "gpt-4-turbo", nil)
	assert.Equal(t, "gpt-4-turbo", client.GetModel())
}

func TestNewClient_WithProxy(t *testing.T) {
	proxy := &ProxyConfig{
		Enabled: true,
		URL:     "http://proxy.example.com:8080",
	}
	client := NewClient("test-key", "https://api.openai.com/v1", "gpt-4o-mini", proxy)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client.Transport)
}

func TestNewClient_WithProxyAuth(t *testing.T) {
	proxy := &ProxyConfig{
		Enabled:  true,
		URL:      "http://proxy.example.com:8080",
		Username: "user",
		Password: "pass",
	}
	client := NewClient("test-key", "https://api.openai.com/v1", "gpt-4o-mini", proxy)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client.Transport)
}

func TestNewClient_ProxyDisabled(t *testing.T) {
	proxy := &ProxyConfig{
		Enabled: false,
		URL:     "http://proxy.example.com:8080",
	}
	client := NewClient("test-key", "https://api.openai.com/v1", "gpt-4o-mini", proxy)
	assert.NotNil(t, client)
}

func TestNewClient_ProxyInvalidURL(t *testing.T) {
	proxy := &ProxyConfig{
		Enabled: true,
		URL:     "://invalid",
	}
	client := NewClient("test-key", "https://api.openai.com/v1", "gpt-4o-mini", proxy)
	assert.NotNil(t, client)
}
