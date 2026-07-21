package fetcher

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Fetcher struct {
	client      *http.Client
	documentsDir string
}

func NewFetcher(documentsDir string) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 60 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
		documentsDir: documentsDir,
	}
}

type FetchResult struct {
	URL        string
	StatusCode int
	Content    []byte
	Hash       string
	FilePath   string
	IsPDF      bool
}

func (f *Fetcher) Fetch(ctx context.Context, targetURL string) (*FetchResult, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/pdf,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.8")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK status %d for %s", resp.StatusCode, targetURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	isPDF := strings.Contains(contentType, "application/pdf") || strings.HasSuffix(strings.ToLower(parsed.Path), ".pdf")

	hash := fmt.Sprintf("%x", sha256.Sum256(body))

	result := &FetchResult{
		URL:        targetURL,
		StatusCode: resp.StatusCode,
		Content:    body,
		Hash:       hash,
		IsPDF:      isPDF,
	}

	if isPDF && f.documentsDir != "" {
		fileName := fmt.Sprintf("%s.pdf", hash[:16])
		filePath := filepath.Join(f.documentsDir, fileName)
		if err := os.WriteFile(filePath, body, 0644); err != nil {
			return result, fmt.Errorf("failed to write PDF: %w", err)
		}
		result.FilePath = filePath
	}

	return result, nil
}

func (f *Fetcher) FetchHTML(ctx context.Context, targetURL string) (string, error) {
	result, err := f.Fetch(ctx, targetURL)
	if err != nil {
		return "", err
	}
	if result.IsPDF {
		return "", fmt.Errorf("expected HTML but got PDF")
	}
	return string(result.Content), nil
}

func (f *Fetcher) FetchPDF(ctx context.Context, targetURL string) (*FetchResult, error) {
	result, err := f.Fetch(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	if !result.IsPDF {
		return result, fmt.Errorf("expected PDF but got HTML")
	}
	return result, nil
}

func (f *Fetcher) EnsureDocumentsDir() error {
	if f.documentsDir == "" {
		return nil
	}
	return os.MkdirAll(f.documentsDir, 0755)
}