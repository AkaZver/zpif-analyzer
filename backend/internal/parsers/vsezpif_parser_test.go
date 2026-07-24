package parsers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVsezpifParser_GetFundDataByISIN(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		funds := []VsezpifFund{
			{
				ID:                   1,
				Name:                 "Test Fund",
				ISIN:                 "RU000TEST01",
				Price:                1000.0,
				PaymentsPerYear:      12,
				PaymentBeforeTax:     80.0,
				ForQualifiedInvestors: 0,
				Description:          "Складской комплекс",
				ManagementCompany:    "Test UK",
				ObjectsInFund:        "3 объекта",
				Renters:              "Ozon, Wildberries",
				UKCommission:         "1.5% от СЧА",
				PaymentsForLastYear:  960.0,
				FundLifetime:         "31.12.2030",
			},
			{
				ID:   2,
				Name: "Other Fund",
				ISIN: "RU000OTHER",
			},
		}
		json.NewEncoder(w).Encode(funds)
	}))
	defer server.Close()

	parser := &VsezpifParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetFundDataByISIN("RU000TEST01")

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 3, data.NumberOfProperties)
	assert.Equal(t, "Ozon, Wildberries", data.MainTenants)
	assert.Equal(t, "склады", data.RealEstateSegment)
	assert.NotNil(t, data.FundEndDate)
	assert.Equal(t, time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC), *data.FundEndDate)
	assert.Equal(t, 1.5, data.ManagementFeePct)
	assert.Equal(t, "sca", data.ManagementFeeType)
	assert.Equal(t, 960.0, data.AnnualPayoutRub)
	assert.Equal(t, 12, data.PaymentsPerYear)
}

func TestVsezpifParser_GetFundDataByISIN_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		funds := []VsezpifFund{
			{ID: 1, ISIN: "RU000OTHER"},
		}
		json.NewEncoder(w).Encode(funds)
	}))
	defer server.Close()

	parser := &VsezpifParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetFundDataByISIN("RU000TEST01")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "not found")
}

func TestVsezpifParser_SearchByISIN(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		funds := []VsezpifFund{
			{
				ID:                   42,
				Name:                 "Test Fund",
				ISIN:                 "RU000TEST01",
				Description:          "Складской комплекс",
				ObjectsInFund:        "3 объекта",
				Renters:              "Ozon",
				UKCommission:         "1.5% от СЧА",
				PaymentsForLastYear:  960.0,
				FundLifetime:         "31.12.2030",
			},
		}
		json.NewEncoder(w).Encode(funds)
	}))
	defer server.Close()

	parser := &VsezpifParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	fundURL, data, err := parser.SearchByISIN("RU000TEST01")

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, server.URL+"/?route=fund&id=42", fundURL)
	assert.Equal(t, 3, data.NumberOfProperties)
	assert.Equal(t, "Ozon", data.MainTenants)
}

func TestVsezpifParser_SearchByISIN_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		funds := []VsezpifFund{
			{ID: 1, ISIN: "RU000OTHER"},
		}
		json.NewEncoder(w).Encode(funds)
	}))
	defer server.Close()

	parser := &VsezpifParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	fundURL, data, err := parser.SearchByISIN("RU000TEST01")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Empty(t, fundURL)
	assert.Contains(t, err.Error(), "not found")
}

func TestVsezpifParser_GetFundDataByISIN_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	parser := &VsezpifParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetFundDataByISIN("RU000TEST01")

	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestVsezpifParser_GetFundDataByISIN_StringNumericFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonData := `[{"id":1,"name":"Test Fund","isin":"RU000TEST01","price":"1000.50","payments_per_year":12,"payment_before_tax":"80.00","for_qualified_investors":0,"description":"Складской комплекс","management_company":"Test UK","payment_stability":"стабильно","payment_indexation":"индексация","calculated_nav":"1050.25","discount_percent":"5.5","fund_lifetime":"31.12.2030","nav":"1100.00","payments_for_last_year":"960.00","avg_trade_volume":"1000000","renters":"Ozon, Wildberries","objects_in_fund":"3 объекта","debt_to_nav":"нет","market_maker":"да","slug":"test-fund","uk_commission":"1.5% от СЧА","yield_before_tax_year":"8.5","payment_after_tax":"75.00","yield_month":"0.7","yield_year":"8.4","yield_last_year":"7.9","full_yield_last_year":"8.1"}]`
		w.Write([]byte(jsonData))
	}))
	defer server.Close()

	parser := &VsezpifParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetFundDataByISIN("RU000TEST01")

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 3, data.NumberOfProperties)
	assert.Equal(t, "Ozon, Wildberries", data.MainTenants)
	assert.Equal(t, "склады", data.RealEstateSegment)
	assert.Equal(t, 960.0, data.AnnualPayoutRub)
}

func TestVsezpifParser_GetFundDataByURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fund := VsezpifFund{
			ID:                123,
			Name:              "Test Fund",
			ISIN:              "RU000TEST01",
			ObjectsInFund:     "5 складов",
			Renters:           "Amazon",
			Description:       "Офисный центр",
			UKCommission:      "2% от дохода",
			PaymentsForLastYear: 1200.0,
		}
		json.NewEncoder(w).Encode(fund)
	}))
	defer server.Close()

	parser := &VsezpifParser{
		client:  server.Client(),
		baseURL: server.URL,
	}

	data, err := parser.GetFundDataByURL(server.URL + "?route=api&action=get_fund&id=123")

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 5, data.NumberOfProperties)
	assert.Equal(t, "Amazon", data.MainTenants)
	assert.Equal(t, "офисы", data.RealEstateSegment)
	assert.Equal(t, 2.0, data.ManagementFeePct)
	assert.Equal(t, "income", data.ManagementFeeType)
}

func TestVsezpifParser_GetFundDataByURL_InvalidURL(t *testing.T) {
	parser := NewVsezpifParser()

	data, err := parser.GetFundDataByURL("https://example.com/invalid")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "failed to extract fund ID")
}

func TestVsezpifParser_ExtractFundIDFromURL(t *testing.T) {
	parser := NewVsezpifParser()

	id, err := parser.extractFundIDFromURL("https://vsezpif.ru/?route=api&action=get_fund&id=123")
	assert.NoError(t, err)
	assert.Equal(t, 123, id)

	id, err = parser.extractFundIDFromURL("https://vsezpif.ru/fund?id=456&other=param")
	assert.NoError(t, err)
	assert.Equal(t, 456, id)

	_, err = parser.extractFundIDFromURL("https://vsezpif.ru/fund")
	assert.Error(t, err)
}

func TestParseNumberOfProperties(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"3 объекта", 3},
		{"5 складов", 5},
		{"10", 10},
		{"", 0},
		{"нет данных", 0},
	}

	for _, tt := range tests {
		result := parseNumberOfProperties(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestParseSegment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Складской комплекс", "склады"},
		{"Бизнес-центр", "офисы"},
		{"Офисное здание", "офисы"},
		{"Торговый центр", "ТЦ"},
		{"Дата-центр", "ЦОД"},
		{"Жилой комплекс", "жильё"},
		{"Склад и офис", "смешанный"},
		{"Неизвестный тип", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := parseSegment(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestParseFundEndDate(t *testing.T) {
	tests := []struct {
		input    string
		expected *time.Time
	}{
		{"31.12.2030", func() *time.Time { t := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC); return &t }()},
		{"2030-12-31", func() *time.Time { t := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC); return &t }()},
		{"", nil},
		{"Нет данных", nil},
		{"invalid", nil},
	}

	for _, tt := range tests {
		result := parseFundEndDate(tt.input)
		if tt.expected == nil {
			assert.Nil(t, result, "input: %s", tt.input)
		} else {
			assert.NotNil(t, result, "input: %s", tt.input)
			assert.Equal(t, *tt.expected, *result, "input: %s", tt.input)
		}
	}
}

func TestParseManagementFee(t *testing.T) {
	tests := []struct {
		input       string
		expectedPct float64
		expectedText string
		expectedType string
	}{
		{"1.5% от СЧА", 1.5, "1.5% от СЧА", "sca"},
		{"2% от дохода", 2.0, "2% от дохода", "income"},
		{"2,5% от СЧА", 2.5, "2,5% от СЧА", "sca"},
		{"", 0, "", ""},
		{"Нет данных", 0, "", ""},
	}

	for _, tt := range tests {
		pct, text, feeType := parseManagementFee(tt.input)
		assert.Equal(t, tt.expectedPct, pct, "input: %s", tt.input)
		assert.Equal(t, tt.expectedText, text, "input: %s", tt.input)
		assert.Equal(t, tt.expectedType, feeType, "input: %s", tt.input)
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
	}{
		{100.5, 100.5},
		{"200.5", 200.5},
		{"300,5", 300.5},
		{"", 0.0},
		{nil, 0.0},
		{json.Number("400.5"), 400.5},
	}

	for _, tt := range tests {
		result := parseFloat(tt.input)
		assert.Equal(t, tt.expected, result, "input: %v", tt.input)
	}
}

func TestNewVsezpifParser(t *testing.T) {
	parser := NewVsezpifParser()
	assert.NotNil(t, parser)
	assert.NotNil(t, parser.client)
	assert.Equal(t, 30*time.Second, parser.client.Timeout)
}
