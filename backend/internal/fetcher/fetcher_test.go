package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFetcher(t *testing.T) {
	f := NewFetcher("/tmp/test-docs")
	assert.NotNil(t, f)
	assert.Equal(t, "/tmp/test-docs", f.documentsDir)
	assert.NotNil(t, f.client)
}

func TestFetcher_Fetch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	f := NewFetcher("")
	result, err := f.Fetch(context.Background(), server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, server.URL, result.URL)
	assert.Equal(t, 200, result.StatusCode)
	assert.Contains(t, string(result.Content), "Test content")
	assert.NotEmpty(t, result.Hash)
	assert.False(t, result.IsPDF)
}

func TestFetcher_Fetch_PDF(t *testing.T) {
	tmpDir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4 test pdf content"))
	}))
	defer server.Close()

	f := NewFetcher(tmpDir)
	result, err := f.Fetch(context.Background(), server.URL+"/test.pdf")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsPDF)
	assert.NotEmpty(t, result.FilePath)
	assert.FileExists(t, result.FilePath)

	content, err := os.ReadFile(result.FilePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "%PDF")
}

func TestFetcher_Fetch_InvalidURL(t *testing.T) {
	f := NewFetcher("")
	result, err := f.Fetch(context.Background(), "not-a-url")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFetcher_Fetch_NonHTTPScheme(t *testing.T) {
	f := NewFetcher("")
	result, err := f.Fetch(context.Background(), "ftp://example.com/file")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported scheme")
}

func TestFetcher_Fetch_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	f := NewFetcher("")
	result, err := f.Fetch(context.Background(), server.URL)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "non-OK status 404")
}

func TestFetcher_FetchHTML_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>HTML content</body></html>"))
	}))
	defer server.Close()

	f := NewFetcher("")
	html, err := f.FetchHTML(context.Background(), server.URL)

	assert.NoError(t, err)
	assert.Contains(t, html, "HTML content")
}

func TestFetcher_FetchHTML_ReturnsPDF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4"))
	}))
	defer server.Close()

	f := NewFetcher("")
	html, err := f.FetchHTML(context.Background(), server.URL+"/test.pdf")

	assert.Error(t, err)
	assert.Empty(t, html)
	assert.Contains(t, err.Error(), "expected HTML but got PDF")
}

func TestFetcher_FetchPDF_Success(t *testing.T) {
	tmpDir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4 test pdf"))
	}))
	defer server.Close()

	f := NewFetcher(tmpDir)
	result, err := f.FetchPDF(context.Background(), server.URL+"/test.pdf")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsPDF)
}

func TestFetcher_FetchPDF_ReturnsHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html>HTML</html>"))
	}))
	defer server.Close()

	f := NewFetcher("")
	result, err := f.FetchPDF(context.Background(), server.URL)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, err.Error(), "expected PDF but got HTML")
}

func TestFetcher_EnsureDocumentsDir(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test-docs")

	f := NewFetcher(testDir)
	err := f.EnsureDocumentsDir()

	assert.NoError(t, err)
	assert.DirExists(t, testDir)
}

func TestFetcher_EnsureDocumentsDir_EmptyPath(t *testing.T) {
	f := NewFetcher("")
	err := f.EnsureDocumentsDir()

	assert.NoError(t, err)
}
