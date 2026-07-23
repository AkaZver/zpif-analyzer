package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VsezpifFund struct {
	ID                   int         `json:"id"`
	Name                 string      `json:"name"`
	ISIN                 string      `json:"isin"`
	Price                float64     `json:"price"`
	PaymentsPerYear      int         `json:"payments_per_year"`
	PaymentBeforeTax     interface{} `json:"payment_before_tax"`
	ForQualifiedInvestors int        `json:"for_qualified_investors"`
	Description          string      `json:"description"`
	ManagementCompany    string      `json:"management_company"`
	PaymentStability     string      `json:"payment_stability"`
	PaymentIndexation    string      `json:"payment_indexation"`
	CalculatedNAV        float64     `json:"calculated_nav"`
	DiscountPercent      float64     `json:"discount_percent"`
	FundLifetime         string      `json:"fund_lifetime"`
	NAV                  float64     `json:"nav"`
	PaymentsForLastYear  interface{} `json:"payments_for_last_year"`
	AvgTradeVolume       interface{} `json:"avg_trade_volume"`
	Renters              string      `json:"renters"`
	ObjectsInFund        string      `json:"objects_in_fund"`
	DebtToNAV            string      `json:"debt_to_nav"`
	MarketMaker          string      `json:"market_maker"`
	Slug                 string      `json:"slug"`
	UKCommission         string      `json:"uk_commission"`
	YieldBeforeTaxYear   float64     `json:"yield_before_tax_year"`
	PaymentAfterTax      float64     `json:"payment_after_tax"`
	YieldMonth           float64     `json:"yield_month"`
	YieldYear            float64     `json:"yield_year"`
	YieldLastYear        interface{} `json:"yield_last_year"`
	FullYieldLastYear    interface{} `json:"full_yield_last_year"`
}

type VsezpifData struct {
	NumberOfProperties    int
	MainTenants           string
	RealEstateSegment     string
	FundEndDate           *time.Time
	ManagementFeePct      float64
	ManagementFeeText     string
	ManagementFeeType     string // "sca" или "income"
	AnnualPayoutRub       float64
	PaymentBeforeTaxRub   float64
	PaymentsPerYear       int
}

type VsezpifParser struct {
	client  *http.Client
	baseURL string
}

func NewVsezpifParser() *VsezpifParser {
	return &VsezpifParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://vsezpif.ru",
	}
}

func (p *VsezpifParser) GetFundDataByISIN(isin string) (*VsezpifData, error) {
	fund, err := p.getFundByISIN(isin)
	if err != nil {
		return nil, err
	}
	return p.parseFundData(fund)
}

func (p *VsezpifParser) GetFundDataByURL(fundURL string) (*VsezpifData, error) {
	fundID, err := p.extractFundIDFromURL(fundURL)
	if err != nil {
		return nil, err
	}
	
	fund, err := p.getFundByID(fundID)
	if err != nil {
		return nil, err
	}
	return p.parseFundData(fund)
}

func (p *VsezpifParser) getFundByISIN(isin string) (*VsezpifFund, error) {
	url := p.baseURL + "/?route=api&action=get_funds"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch funds list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vsezpif API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var funds []VsezpifFund
	if err := json.Unmarshal(body, &funds); err != nil {
		return nil, fmt.Errorf("failed to parse funds list: %w", err)
	}

	for _, fund := range funds {
		if fund.ISIN == isin {
			return &fund, nil
		}
	}

	return nil, fmt.Errorf("fund with ISIN %s not found", isin)
}

func (p *VsezpifParser) getFundByID(id int) (*VsezpifFund, error) {
	url := fmt.Sprintf("%s/?route=api&action=get_fund&id=%d", p.baseURL, id)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fund data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vsezpif API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var fund VsezpifFund
	if err := json.Unmarshal(body, &fund); err != nil {
		return nil, fmt.Errorf("failed to parse fund data: %w", err)
	}

	return &fund, nil
}

func (p *VsezpifParser) extractFundIDFromURL(fundURL string) (int, error) {
	re := regexp.MustCompile(`[?&]id=(\d+)`)
	matches := re.FindStringSubmatch(fundURL)
	if len(matches) >= 2 {
		id, err := strconv.Atoi(matches[1])
		if err == nil {
			return id, nil
		}
	}
	return 0, fmt.Errorf("failed to extract fund ID from URL: %s", fundURL)
}

func (p *VsezpifParser) parseFundData(fund *VsezpifFund) (*VsezpifData, error) {
	data := &VsezpifData{}

	data.NumberOfProperties = parseNumberOfProperties(fund.ObjectsInFund)
	data.MainTenants = fund.Renters
	data.RealEstateSegment = parseSegment(fund.Description)
	data.FundEndDate = parseFundEndDate(fund.FundLifetime)
	data.ManagementFeePct, data.ManagementFeeText, data.ManagementFeeType = parseManagementFee(fund.UKCommission)
	data.AnnualPayoutRub = parseFloat(fund.PaymentsForLastYear)
	data.PaymentBeforeTaxRub = parseFloat(fund.PaymentBeforeTax)
	data.PaymentsPerYear = fund.PaymentsPerYear

	return data, nil
}

func parseNumberOfProperties(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		if n, err := strconv.Atoi(matches[1]); err == nil {
			return n
		}
	}
	return 0
}

func parseSegment(description string) string {
	descLower := strings.ToLower(description)
	
	segments := []struct {
		name     string
		keywords []string
	}{
		{"склады", []string{"склад", "логистик"}},
		{"офисы", []string{"офис", "бизнес-центр", "бц"}},
		{"ТЦ", []string{"торгов", "тц", "молл", "магазин"}},
		{"ЦОД", []string{"цод", "дата-центр", "data center"}},
		{"жильё", []string{"квартир", "жил", "апартамент"}},
	}
	
	found := []string{}
	for _, seg := range segments {
		for _, keyword := range seg.keywords {
			if strings.Contains(descLower, keyword) {
				found = append(found, seg.name)
				break
			}
		}
	}
	
	if len(found) == 0 {
		return ""
	}
	if len(found) == 1 {
		return found[0]
	}
	return "смешанный"
}

func parseFundEndDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" || s == "Нет данных" {
		return nil
	}
	
	t, err := time.Parse("02.01.2006", s)
	if err == nil {
		return &t
	}
	
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		return &t
	}
	
	return nil
}

func parseManagementFee(s string) (float64, string, string) {
	s = strings.TrimSpace(s)
	if s == "" || s == "Нет данных" {
		return 0, "", ""
	}
	
	feeType := "income"
	if strings.Contains(strings.ToLower(s), "сча") {
		feeType = "sca"
	}
	
	re := regexp.MustCompile(`(\d+(?:[.,]\d+)?)\s*%`)
	matches := re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		val := strings.ReplaceAll(matches[1], ",", ".")
		if pct, err := strconv.ParseFloat(val, 64); err == nil {
			return pct, s, feeType
		}
	}
	
	return 0, s, feeType
}

func parseFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	
	switch val := v.(type) {
	case float64:
		return val
	case string:
		if val == "" {
			return 0
		}
		val = strings.ReplaceAll(val, ",", ".")
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return f
		}
	}
	
	return 0
}
