package parsers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMoexParser_GetCurrentPriceWithBoard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := moexResponse{}
		resp.MarketData.Columns = []string{"LAST", "OPEN", "HIGH", "LOW", "VOLUME", "VALUE"}
		resp.MarketData.Data = [][]interface{}{
			{1000.5, 995.0, 1010.0, 990.0, 50000.0, 50000000.0},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetCurrentPriceWithBoard("TEST", "TQBR")

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1000.5, data.Close)
	assert.Equal(t, 995.0, data.Open)
	assert.Equal(t, 1010.0, data.High)
	assert.Equal(t, 990.0, data.Low)
	assert.Equal(t, int64(50000), data.Volume)
	assert.Equal(t, 50000000.0, data.Value)
}

func TestMoexParser_GetCurrentPriceWithBoard_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetCurrentPriceWithBoard("TEST", "TQBR")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "status 500")
}

func TestMoexParser_GetCurrentPriceWithBoard_EmptyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := moexResponse{}
		resp.MarketData.Columns = []string{"LAST"}
		resp.MarketData.Data = [][]interface{}{}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetCurrentPriceWithBoard("TEST", "TQBR")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "no market data")
}

func TestMoexParser_SearchSecurity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"description": map[string]interface{}{
				"sec_id": "TEST01",
				"name":   "Test Fund",
				"isin":   "RU000TEST01",
			},
			"boards": map[string]interface{}{
				"columns": []string{"boardid", "market"},
				"data": [][]interface{}{
					{"TQBR", "shares"},
					{"TQIF", "shares"},
					{"SMAL", "bonds"},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	security, err := parser.SearchSecurity("RU000TEST01")

	assert.NoError(t, err)
	assert.NotNil(t, security)
	assert.Equal(t, "RU000TEST01", security.SecID)
	assert.Equal(t, "RU000TEST01", security.ISIN)
	assert.Equal(t, "Test Fund", security.Name)
	assert.Contains(t, security.Boards, "TQBR")
	assert.Contains(t, security.Boards, "TQIF")
	assert.NotContains(t, security.Boards, "SMAL")
}

func TestMoexParser_SearchSecurity_NoBoards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"description": map[string]interface{}{
				"sec_id": "TEST01",
				"name":   "Test Fund",
				"isin":   "RU000TEST01",
			},
			"boards": map[string]interface{}{
				"columns": []string{"boardid", "market"},
				"data":    [][]interface{}{},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	security, err := parser.SearchSecurity("RU000TEST01")

	assert.Error(t, err)
	assert.Nil(t, security)
	assert.Contains(t, err.Error(), "no trading boards")
}

func TestMoexParser_GetPriceHistoryWithBoard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := moexResponse{}
		resp.History.Columns = []string{"TRADEDATE", "CLOSE", "OPEN", "HIGH", "LOW", "VOLUME", "VALUE"}
		resp.History.Data = [][]interface{}{
			{"2024-01-15", 1000.0, 995.0, 1010.0, 990.0, 50000.0, 50000000.0},
			{"2024-01-16", 1005.0, 1000.0, 1015.0, 995.0, 55000.0, 55000000.0},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	history, err := parser.GetPriceHistoryWithBoard("TEST", "TQBR")

	assert.NoError(t, err)
	assert.Len(t, history, 2)

	assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), history[0].Date)
	assert.Equal(t, 1000.0, history[0].Close)
	assert.Equal(t, 995.0, history[0].Open)

	assert.Equal(t, time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), history[1].Date)
	assert.Equal(t, 1005.0, history[1].Close)
}

func TestMoexParser_GetPriceHistoryWithBoard_FallbackPrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := moexResponse{}
		resp.History.Columns = []string{"TRADEDATE", "CLOSE", "LEGALCLOSEPRICE", "WAPRICE"}
		resp.History.Data = [][]interface{}{
			{"2024-01-15", 0.0, 1000.0, 0.0},
			{"2024-01-16", 0.0, 0.0, 1005.0},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	history, err := parser.GetPriceHistoryWithBoard("TEST", "TQBR")

	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, 1000.0, history[0].Close)
	assert.Equal(t, 1005.0, history[1].Close)
}

func TestMoexParser_GetPriceHistoryWithBoard_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := moexResponse{}
		resp.History.Columns = []string{"TRADEDATE", "CLOSE"}

		if callCount == 1 {
			data := make([][]interface{}, 100)
			for i := 0; i < 100; i++ {
				data[i] = []interface{}{"2024-01-15", 1000.0}
			}
			resp.History.Data = data
		} else {
			resp.History.Data = [][]interface{}{
				{"2024-01-16", 1005.0},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	history, err := parser.GetPriceHistoryWithBoard("TEST", "TQBR")

	assert.NoError(t, err)
	assert.Len(t, history, 101)
	assert.Equal(t, 2, callCount)
}

func TestMoexParser_GetCurrentPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "TQBR")
		resp := moexResponse{}
		resp.MarketData.Columns = []string{"LAST"}
		resp.MarketData.Data = [][]interface{}{{1000.0}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetCurrentPrice("TEST")

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1000.0, data.Close)
}

func TestMoexParser_GetPriceHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "TQBR")
		resp := moexResponse{}
		resp.History.Columns = []string{"TRADEDATE", "CLOSE"}
		resp.History.Data = [][]interface{}{{"2024-01-15", 1000.0}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	parser := &MoexParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	history, err := parser.GetPriceHistory("TEST")

	assert.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestNewMoexParser(t *testing.T) {
	parser := NewMoexParser()
	assert.NotNil(t, parser)
	assert.NotNil(t, parser.client)
	assert.Equal(t, 30*time.Second, parser.client.Timeout)
}
