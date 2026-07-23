package parsers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvestfundsParser_SearchFund(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body>
			<a href="/funds/123-test-fund">Test Fund</a>
			<a href="/funds/456-other">Other Fund</a>
		</body></html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	parser := &InvestfundsParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	url, err := parser.SearchFund("test")

	assert.NoError(t, err)
	assert.Contains(t, url, "/funds/123-test-fund")
}

func TestInvestfundsParser_SearchFund_AbsoluteURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body>
			<a href="https://investfunds.ru/funds/123-test">Test Fund</a>
		</body></html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	parser := &InvestfundsParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	url, err := parser.SearchFund("test")

	assert.NoError(t, err)
	assert.Equal(t, "https://investfunds.ru/funds/123-test", url)
}

func TestInvestfundsParser_SearchFund_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body><p>No results</p></body></html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	parser := &InvestfundsParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	url, err := parser.SearchFund("nonexistent")

	assert.Error(t, err)
	assert.Empty(t, url)
	assert.Contains(t, err.Error(), "not found")
}

func TestInvestfundsParser_SearchFund_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	parser := &InvestfundsParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	url, err := parser.SearchFund("test")

	assert.Error(t, err)
	assert.Empty(t, url)
}

func TestInvestfundsParser_GetFundData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body>
			<table class="table_part">
				<tbody>
					<tr><td>15.01.2024</td><td>1,050.50</td><td>5,000,000</td></tr>
					<tr><td>14.01.2024</td><td>1,045.00</td><td>4,950,000</td></tr>
				</tbody>
			</table>
			<table class="dividends_table">
				<tbody>
					<tr><td>1</td><td>15.01.2024</td><td>3</td><td>80.00</td><td>8.5</td></tr>
					<tr><td>2</td><td>15.12.2023</td><td>4</td><td>75.50</td><td>8.0</td></tr>
				</tbody>
			</table>
		</body></html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	parser := &InvestfundsParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetFundData(server.URL + "/funds/123")

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1050.50, data.NAV)
	assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), data.NAVDate)
	assert.Equal(t, 5000000.0, data.TotalAssets)
	assert.Len(t, data.NAVHistory, 2)
	assert.Len(t, data.PayoutHistory, 2)

	assert.Equal(t, 80.0, data.PayoutHistory[0].Amount)
	assert.Equal(t, 8.5, data.PayoutHistory[0].YieldPercent)
}

func TestInvestfundsParser_GetFundData_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	parser := &InvestfundsParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetFundData(server.URL + "/funds/123")

	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestInvestfundsParser_GetFundData_EmptyPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body><p>No data</p></body></html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	parser := &InvestfundsParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetFundData(server.URL + "/funds/123")

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 0.0, data.NAV)
	assert.Len(t, data.NAVHistory, 0)
	assert.Len(t, data.PayoutHistory, 0)
}

func TestParseRussianNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1,050.50", 1050.50},
		{"1,000", 1000.0},
		{"500.75", 500.75},
		{"1,234,567.89", 1234567.89},
		{"100 ₽", 100.0},
		{"200 руб", 200.0},
		{"300 руб.", 300.0},
		{"", 0.0},
		{"invalid", 0.0},
	}

	for _, tt := range tests {
		result := parseRussianNumber(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestParseRussianDate(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"15.01.2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"01.12.2023", time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)},
		{"", time.Time{}},
		{"invalid", time.Time{}},
	}

	for _, tt := range tests {
		result := parseRussianDate(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestNewInvestfundsParser(t *testing.T) {
	parser := NewInvestfundsParser()
	assert.NotNil(t, parser)
	assert.NotNil(t, parser.client)
	assert.Equal(t, 30*time.Second, parser.client.Timeout)
}
