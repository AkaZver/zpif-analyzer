package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zpif-analyzer/backend/internal/fetcher"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"github.com/zpif-analyzer/backend/internal/websearch"
)

type MockFetcherManager struct {
	mock.Mock
}

func (m *MockFetcherManager) FetchHTML(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

func (m *MockFetcherManager) FetchPDF(ctx context.Context, url string) (*fetcher.FetchResult, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fetcher.FetchResult), args.Error(1)
}

type MockSearchClient struct {
	mock.Mock
}

func (m *MockSearchClient) Search(ctx context.Context, query string) ([]websearch.SearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]websearch.SearchResult), args.Error(1)
}

func TestNewDiscoverer(t *testing.T) {
	llmClient := NewClient("test-key", "", "gpt-4o-mini")
	searchClient := &MockSearchClient{}
	fetcherMgr := &MockFetcherManager{}
	documentRepo := &repositories.DocumentRepository{}
	fundRepo := &repositories.FundRepository{}

	discoverer := NewDiscoverer(llmClient, searchClient, fetcherMgr, documentRepo, fundRepo)

	assert.NotNil(t, discoverer)
	assert.Equal(t, llmClient, discoverer.llmClient)
	assert.Equal(t, 5, discoverer.maxPages)
	assert.Equal(t, 3, discoverer.concurrency)
}

func TestDiscoverer_GetStatus_Idle(t *testing.T) {
	llmClient := NewClient("test-key", "", "gpt-4o-mini")
	discoverer := NewDiscoverer(llmClient, nil, nil, nil, nil)

	status := discoverer.GetStatus(123)

	assert.NotNil(t, status)
	assert.Equal(t, uint(123), status.FundID)
	assert.Equal(t, "idle", status.Status)
	assert.Equal(t, 0, status.URLsFound)
	assert.Equal(t, 0, status.Downloaded)
	assert.Equal(t, 0, status.Errors)
}

func TestExtractJSON_ValidArray(t *testing.T) {
	input := `Some text before [{"url":"http://example.com","type":"report","title":"Test"}] and after`
	result := extractJSON(input)

	assert.Equal(t, `[{"url":"http://example.com","type":"report","title":"Test"}]`, result)
}

func TestExtractJSON_NoArray(t *testing.T) {
	input := `No JSON array here`
	result := extractJSON(input)

	assert.Empty(t, result)
}

func TestExtractJSON_IncompleteArray(t *testing.T) {
	input := `[{"url":"http://example.com"`
	result := extractJSON(input)

	assert.Empty(t, result)
}

func TestParseDocumentList_ValidJSON(t *testing.T) {
	rawResponse := `[{"url":"http://example.com/doc.pdf","type":"report","title":"Annual Report"}]`
	docs, err := parseDocumentList(rawResponse)

	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "http://example.com/doc.pdf", docs[0].URL)
	assert.Equal(t, "report", docs[0].Type)
	assert.Equal(t, "Annual Report", docs[0].Title)
}

func TestParseDocumentList_InvalidJSON(t *testing.T) {
	rawResponse := `not valid json`
	docs, err := parseDocumentList(rawResponse)

	assert.Error(t, err)
	assert.Nil(t, docs)
	assert.Contains(t, err.Error(), "no valid JSON in response")
}

func TestResolveURL_AbsoluteURL(t *testing.T) {
	result := resolveURL("http://example.com/page", "https://other.com/doc.pdf")
	assert.Equal(t, "https://other.com/doc.pdf", result)
}

func TestResolveURL_RelativeURL(t *testing.T) {
	result := resolveURL("http://example.com/page", "/docs/file.pdf")
	assert.Equal(t, "http://example.com/docs/file.pdf", result)
}

func TestResolveURL_RelativePath(t *testing.T) {
	result := resolveURL("http://example.com/page", "docs/file.pdf")
	assert.Equal(t, "http://example.com/docs/file.pdf", result)
}

func TestResolveURL_EmptyRef(t *testing.T) {
	result := resolveURL("http://example.com/page", "")
	assert.Empty(t, result)
}

func TestTruncateString_Short(t *testing.T) {
	result := truncateString("short", 10)
	assert.Equal(t, "short", result)
}

func TestTruncateString_Long(t *testing.T) {
	result := truncateString("this is a very long string", 10)
	assert.Equal(t, "this is a ...", result)
}

func TestDiscoverer_Discover_NoSearchResults(t *testing.T) {
	llmClient := NewClient("test-key", "", "gpt-4o-mini")
	searchClient := &MockSearchClient{}
	fetcherMgr := &MockFetcherManager{}
	documentRepo := &repositories.DocumentRepository{}
	fundRepo := &repositories.FundRepository{}

	discoverer := NewDiscoverer(llmClient, searchClient, fetcherMgr, documentRepo, fundRepo)

	fund := &models.Fund{
		ID:   1,
		Name: "Test Fund",
		ISIN: "RU000TEST001",
	}

	searchClient.On("Search", mock.Anything, mock.Anything).Return([]websearch.SearchResult{}, nil)

	status, err := discoverer.Discover(context.Background(), fund)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "completed", status.Status)
	assert.Equal(t, 0, status.URLsFound)
	assert.Equal(t, 0, status.Downloaded)
}
