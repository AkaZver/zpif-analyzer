package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zpif-analyzer/backend/internal/fetcher"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"github.com/zpif-analyzer/backend/internal/websearch"
)

type Discoverer struct {
	llmClient    *Client
	searchClient websearch.Client
	fetcherMgr   FetcherManager
	documentRepo *repositories.DocumentRepository
	fundRepo     *repositories.FundRepository
	maxPages     int
	concurrency  int
}

type FetcherManager interface {
	FetchHTML(ctx context.Context, url string) (string, error)
	FetchPDF(ctx context.Context, url string) (*fetcher.FetchResult, error)
}

func NewDiscoverer(
	llmClient *Client,
	searchClient websearch.Client,
	fetcherMgr FetcherManager,
	documentRepo *repositories.DocumentRepository,
	fundRepo *repositories.FundRepository,
) *Discoverer {
	return &Discoverer{
		llmClient:    llmClient,
		searchClient: searchClient,
		fetcherMgr:   fetcherMgr,
		documentRepo: documentRepo,
		fundRepo:     fundRepo,
		maxPages:     5,
		concurrency:  3,
	}
}

type DiscoveredDocument struct {
	URL   string `json:"url"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type DiscoveryStatus struct {
	FundID     uint
	Status     string
	URLsFound  int
	Downloaded int
	Errors     int
	StartedAt  time.Time
	FinishedAt time.Time
}

var discoveryStorage sync.Map

func (d *Discoverer) Discover(ctx context.Context, fund *models.Fund) (*DiscoveryStatus, error) {
	status := &DiscoveryStatus{
		FundID:    fund.ID,
		Status:    "in_progress",
		StartedAt: time.Now(),
	}
	discoveryStorage.Store(fund.ID, status)
	defer func() {
		status.FinishedAt = time.Now()
		if status.Status == "in_progress" {
			status.Status = "completed"
		}
		discoveryStorage.Store(fund.ID, status)
	}()

	query := fmt.Sprintf("%q OR %q документы отчёт оценщика КИД ПДУ", fund.Name, fund.ISIN)
	results, err := d.searchClient.Search(ctx, query)
	if err != nil {
		status.Status = "error"
		return status, fmt.Errorf("web search failed: %w", err)
	}
	status.URLsFound = len(results)
	if len(results) == 0 {
		status.Status = "completed"
		return status, nil
	}

	pagesLimit := len(results)
	if pagesLimit > d.maxPages {
		pagesLimit = d.maxPages
	}

	type pageResult struct {
		url     string
		content string
	}
	pageCh := make(chan pageResult, pagesLimit)
	sem := make(chan struct{}, d.concurrency)
	var wg sync.WaitGroup

	for i := 0; i < pagesLimit; i++ {
		wg.Add(1)
		go func(r websearch.SearchResult) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			html, err := d.fetcherMgr.FetchHTML(ctx, r.URL)
			if err != nil {
				return
			}
			pageCh <- pageResult{url: r.URL, content: html}
		}(results[i])
	}
	go func() {
		wg.Wait()
		close(pageCh)
	}()

	for page := range pageCh {
		docs, err := d.extractDocumentURLs(ctx, page.content, page.url)
		if err != nil {
			continue
		}
		for _, doc := range docs {
			if err := d.downloadAndStore(ctx, fund, doc.URL, doc.Type, doc.Title); err != nil {
				status.Errors++
				continue
			}
			status.Downloaded++
		}
	}

	status.Status = "completed"
	return status, nil
}

func (d *Discoverer) extractDocumentURLs(ctx context.Context, html, sourceURL string) ([]DiscoveredDocument, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	htmlSnippet := truncateString(doc.Text(), 8000)
	cleanedHTML := removeNoiseHTML(html, doc)
	llmHTML := truncateString(cleanedHTML, 12000)
	llmHTML += "\n\nSOURCE_URL=" + sourceURL

	llmResponse, err := d.llmClient.ChatSimple(ctx, DiscoverPrompt, llmHTML)
	if err != nil {
		return nil, fmt.Errorf("LLM discovery failed: %w", err)
	}

	docs, err := parseDocumentList(llmResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	filtered := make([]DiscoveredDocument, 0, len(docs))
	for _, doc := range docs {
		absolute := resolveURL(sourceURL, doc.URL)
		if absolute == "" {
			continue
		}
		doc.URL = absolute
		filtered = append(filtered, doc)
	}

	_ = htmlSnippet
	return filtered, nil
}

func (d *Discoverer) downloadAndStore(ctx context.Context, fund *models.Fund, url string, docType, title string) error {
	resp, err := d.fetcherMgr.FetchPDF(ctx, url)
	if err != nil {
		return err
	}

	existing, _ := d.documentRepo.GetByHash(resp.Hash)
	if existing != nil {
		return nil
	}

	now := time.Now()
	document := &models.FundDocument{
		FundID:       fund.ID,
		FileName:     title + ".pdf",
		FilePath:     resp.FilePath,
		DocumentType: docType,
		ContentHash:  resp.Hash,
		Source:       "auto",
		SourceURL:    url,
		UploadDate:   now,
		Status:       "downloaded",
	}
	return d.documentRepo.Create(document)
}

func (d *Discoverer) GetStatus(fundID uint) *DiscoveryStatus {
	val, ok := discoveryStorage.Load(fundID)
	if !ok {
		return &DiscoveryStatus{FundID: fundID, Status: "idle"}
	}
	return val.(*DiscoveryStatus)
}

func parseDocumentList(rawResponse string) ([]DiscoveredDocument, error) {
	cleaned := extractJSON(rawResponse)
	if cleaned == "" {
		return nil, fmt.Errorf("no valid JSON in response")
	}
	var docs []DiscoveredDocument
	if err := json.Unmarshal([]byte(cleaned), &docs); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return docs, nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "[")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(s, "]")
	if end == -1 || end <= start {
		return ""
	}
	return s[start : end+1]
}

func resolveURL(base, ref string) string {
	if ref == "" {
		return ""
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	parts := strings.Split(base, "/")
	if len(parts) < 3 {
		return ref
	}
	origin := parts[0] + "//" + parts[2]
	if strings.HasPrefix(ref, "/") {
		return origin + ref
	}
	return origin + "/" + ref
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func removeNoiseHTML(raw string, doc *goquery.Document) string {
	doc.Find("script,style,noscript,iframe,svg").Remove()
	html, err := doc.Find("body").Html()
	if err != nil || html == "" {
		return raw
	}
	return html
}