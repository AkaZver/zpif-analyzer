package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type ProxyConfig struct {
	Enabled  bool
	URL      string
	Username string
	Password string
}

func NewClient(apiKey, baseURL, model string, proxy *ProxyConfig) *Client {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	transport := &http.Transport{}

	if proxy != nil && proxy.Enabled && proxy.URL != "" {
		proxyURL, err := url.Parse(proxy.URL)
		if err == nil {
			if proxy.Username != "" {
				proxyURL.User = url.UserPassword(proxy.Username, proxy.Password)
			}
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client: &http.Client{
			Timeout:   180 * time.Second,
			Transport: transport,
		},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *Client) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("LLM api key is empty")
	}

	reqBody := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.3,
	}
	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("LLM request: model=%s, prompt_tokens_estimate=%d", c.model, len(bodyJSON)/4)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return nil, fmt.Errorf("LLM request timeout: %w", err)
		}
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
			return nil, fmt.Errorf("LLM connection error: %w", err)
		}
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		respStr := string(respBody)
		if len(respStr) > 500 {
			respStr = respStr[:500] + "..."
		}
		
		switch resp.StatusCode {
		case 400:
			return nil, fmt.Errorf("LLM error: invalid request (400). Response: %s", respStr)
		case 401:
			return nil, fmt.Errorf("LLM error: invalid API key (401)")
		case 403:
			return nil, fmt.Errorf("LLM error: access forbidden (403)")
		case 404:
			return nil, fmt.Errorf("LLM error: model not found (404). Check model name: %s", c.model)
		case 429:
			return nil, fmt.Errorf("LLM error: rate limit exceeded (429). Try again later")
		case 500, 502, 503, 504:
			return nil, fmt.Errorf("LLM server error (%d). Try again later", resp.StatusCode)
		default:
			return nil, fmt.Errorf("LLM error: status %d. Response: %s", resp.StatusCode, respStr)
		}
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode LLM response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	log.Printf("LLM response: model=%s, tokens=%d (prompt=%d, completion=%d)", 
		c.model, chatResp.Usage.TotalTokens, chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens)

	return &chatResp, nil
}

func (c *Client) ChatSimple(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}
	resp, err := c.Chat(ctx, messages)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *Client) TestConnection(ctx context.Context) error {
	resp, err := c.ChatSimple(ctx, "You are a test assistant. Reply with 'OK'.", "Check connection")
	if err != nil {
		return err
	}
	if resp == "" {
		return fmt.Errorf("empty response from LLM")
	}
	return nil
}

func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("models list error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	return models, nil
}

func (c *Client) GetModel() string {
	return c.model
}

func ExtractJSON(s string) string {
	s = strings.TrimPrefix(s, "\xef\xbb\xbf")
	s = strings.TrimSpace(s)

	if idx := strings.Index(s, "```"); idx != -1 {
		rest := s[idx+3:]
		if nl := strings.IndexByte(rest, '\n'); nl != -1 {
			rest = rest[nl+1:]
		}
		if end := strings.Index(rest, "```"); end != -1 {
			candidate := strings.TrimSpace(rest[:end])
			if strings.HasPrefix(candidate, "{") {
				s = candidate
			}
		}
	}

	replacer := strings.NewReplacer(
		"\u201c", "'",
		"\u201d", "'",
		"\u2018", "'",
		"\u2019", "'",
	)
	s = replacer.Replace(s)

	start := strings.IndexByte(s, '{')
	if start == -1 {
		return s
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		ch := s[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}

	end := strings.LastIndexByte(s, '}')
	if end > start {
		return s[start : end+1]
	}
	return s
}

func SanitizeJSON(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	inString := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escaped {
			result.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' && inString {
			result.WriteByte(ch)
			escaped = true
			continue
		}

		if ch == '"' {
			if !inString {
				inString = true
				result.WriteByte(ch)
			} else {
				rest := strings.TrimSpace(s[i+1:])
				if len(rest) == 0 || rest[0] == ':' || rest[0] == ',' || rest[0] == '}' || rest[0] == ']' {
					inString = false
					result.WriteByte(ch)
				} else {
					result.WriteByte('\'')
				}
			}
			continue
		}

		result.WriteByte(ch)
	}

	return result.String()
}