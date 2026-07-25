package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var mskLocation = time.FixedZone("MSK", 3*3600)

type InvestfundsData struct {
	NAV           float64
	NAVDate       time.Time
	TotalAssets   float64
	NAVHistory    []NAVData
	PayoutHistory []Payout
}

type NAVData struct {
	Date      time.Time
	UnitPrice float64 // Цена пая (колонка "Пай")
	NAV       float64 // РСП (то же, что и UnitPrice для ЗПИФ)
	SCA       float64 // СЧА (колонка "СЧА")
}

type Payout struct {
	Date         time.Time
	Amount       float64
	YieldPercent float64
}

type investfundsSearchResult struct {
	FundID   string `json:"fund_id"`
	FundName string `json:"fund_name"`
	ISIN     string `json:"isin"`
	FundsURL string `json:"funds_url"`
}

type investfundsSearchResponse struct {
	Total          int                       `json:"total"`
	CurrentResults []investfundsSearchResult `json:"currentResults"`
	VerifyHash     string                    `json:"verifyHash"`
}

type investfundsChartResponse []struct {
	Data    [][]float64 `json:"data"`
	Name    string      `json:"name"`
	Tooltip struct {
		ValueDecimals int    `json:"valueDecimals"`
		XDateFormat   string `json:"xDateFormat"`
	} `json:"tooltip"`
}

type InvestfundsParser struct {
	client  *http.Client
	baseURL string
}

func NewInvestfundsParser() *InvestfundsParser {
	return &InvestfundsParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://investfunds.ru",
	}
}

func (p *InvestfundsParser) SearchFund(query string) (string, error) {
	searchURL := fmt.Sprintf("%s/funds/index.php?searchString=%s&verifyHash=check", p.baseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to search fund: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("investfunds search returned status %d", resp.StatusCode)
	}

	var searchResp investfundsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to parse search response: %w", err)
	}

	if searchResp.Total == 0 || len(searchResp.CurrentResults) == 0 {
		return "", fmt.Errorf("fund not found on investfunds.ru for query: %s", query)
	}

	isISIN := isISINFormat(query)
	queryLower := strings.ToLower(query)

	for _, result := range searchResp.CurrentResults {
		if isISIN {
			if strings.EqualFold(result.ISIN, query) {
				return p.baseURL + result.FundsURL, nil
			}
		} else {
			resultNameLower := strings.ToLower(result.FundName)
			if strings.Contains(resultNameLower, queryLower) || strings.Contains(queryLower, resultNameLower) {
				return p.baseURL + result.FundsURL, nil
			}
		}
	}

	return "", fmt.Errorf("fund not found on investfunds.ru for query: %s", query)
}

func isISINFormat(s string) bool {
	isinRegex := regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}\d$`)
	return isinRegex.MatchString(s)
}

func extractFundID(fundURL string) string {
	re := regexp.MustCompile(`/funds/(\d+)`)
	matches := re.FindStringSubmatch(fundURL)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func (p *InvestfundsParser) getChartData(fundURL, fundID, dataKey string) ([][]float64, error) {
	chartURL := fmt.Sprintf("%s?action=chartData&data_key=%s&date_from=01.01.2000&ids[]=%s&currencyId=1",
		fundURL, dataKey, fundID)

	req, err := http.NewRequest("GET", chartURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create chart request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chart data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chart API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read chart response: %w", err)
	}

	var chartResp investfundsChartResponse
	if err := json.Unmarshal(body, &chartResp); err != nil {
		return nil, fmt.Errorf("failed to parse chart response: %w", err)
	}

	if len(chartResp) == 0 {
		return nil, fmt.Errorf("empty chart response")
	}

	return chartResp[0].Data, nil
}

func (p *InvestfundsParser) GetFundData(fundURL string) (*InvestfundsData, error) {
	data := &InvestfundsData{}

	fundID := extractFundID(fundURL)
	if fundID != "" {
		payData, payErr := p.getChartData(fundURL, fundID, "pay")
		scaData, scaErr := p.getChartData(fundURL, fundID, "sca")

		if payErr == nil && len(payData) > 0 {
			scaByTimestamp := make(map[int64]float64)
			if scaErr == nil {
				for _, point := range scaData {
					if len(point) >= 2 {
						scaByTimestamp[int64(point[0])] = point[1]
					}
				}
			}

		for _, point := range payData {
			if len(point) >= 2 {
				tsMs := int64(point[0])
				nav := point[1]
				date := time.Unix(tsMs/1000, 0).In(mskLocation)
				date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, mskLocation)

					navData := NAVData{
						Date:      date,
						UnitPrice: nav,
						NAV:       nav,
						SCA:       scaByTimestamp[tsMs],
					}
					data.NAVHistory = append(data.NAVHistory, navData)
				}
			}

			sort.Slice(data.NAVHistory, func(i, j int) bool {
				return data.NAVHistory[i].Date.After(data.NAVHistory[j].Date)
			})

			if len(data.NAVHistory) > 0 {
				data.NAV = data.NAVHistory[0].NAV
				data.NAVDate = data.NAVHistory[0].Date
				data.TotalAssets = data.NAVHistory[0].SCA
			}
		}
	}

	req, err := http.NewRequest("GET", fundURL, nil)
	if err != nil {
		if len(data.NAVHistory) > 0 {
			return data, nil
		}
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		if len(data.NAVHistory) > 0 {
			return data, nil
		}
		return nil, fmt.Errorf("failed to fetch fund page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if len(data.NAVHistory) > 0 {
			return data, nil
		}
		return nil, fmt.Errorf("investfunds returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		if len(data.NAVHistory) > 0 {
			return data, nil
		}
		return nil, fmt.Errorf("failed to parse fund page: %w", err)
	}

	if len(data.NAVHistory) == 0 {
		doc.Find("table.table_part tbody tr").Each(func(i int, s *goquery.Selection) {
			cells := s.Find("td")
			if cells.Length() >= 3 {
				dateText := strings.TrimSpace(cells.Eq(0).Text())
				navText := strings.TrimSpace(cells.Eq(1).Text())
				scaText := strings.TrimSpace(cells.Eq(2).Text())

				date := parseRussianDate(dateText)
				nav := parseRussianNumber(navText)
				sca := parseRussianNumber(scaText)

				if !date.IsZero() && nav > 0 {
					navData := NAVData{
						Date:      date,
						UnitPrice: nav,
						NAV:       nav,
						SCA:       sca,
					}
					data.NAVHistory = append(data.NAVHistory, navData)

					if i == 0 {
						data.NAV = nav
						data.NAVDate = date
						data.TotalAssets = sca
					}
				}
			}
		})
	}

	doc.Find("table.dividends_table tbody tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() >= 4 {
			paymentDateText := strings.TrimSpace(cells.Eq(1).Text())
			amountText := strings.TrimSpace(cells.Eq(3).Text())
			yieldText := ""
			if cells.Length() >= 5 {
				yieldText = strings.TrimSpace(cells.Eq(4).Text())
			}

			paymentDate := parseRussianDate(paymentDateText)
			amount := parseRussianNumber(amountText)
			yieldPercent := parseRussianNumber(yieldText)

			if !paymentDate.IsZero() && amount > 0 {
				payout := Payout{
					Date:         paymentDate,
					Amount:       amount,
					YieldPercent: yieldPercent,
				}
				data.PayoutHistory = append(data.PayoutHistory, payout)
			}
		}
	})

	return data, nil
}

func parseRussianNumber(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\u00A0", "")
	s = strings.ReplaceAll(s, "₽", "")
	s = strings.ReplaceAll(s, "руб", "")
	s = strings.ReplaceAll(s, "руб.", "")
	s = strings.TrimSpace(s)

	re := regexp.MustCompile(`[\d,]+\.?\d*`)
	match := re.FindString(s)
	if match == "" {
		return 0
	}

	match = strings.ReplaceAll(match, ",", "")
	val, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0
	}

	return val
}

func parseRussianDate(s string) time.Time {
	s = strings.TrimSpace(s)
	
	// Try DD.MM.YYYY format
	parts := strings.Split(s, ".")
	if len(parts) == 3 {
		day := parts[0]
		month := parts[1]
		year := parts[2]
		
		dateStr := fmt.Sprintf("%s-%s-%s", year, month, day)
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			return t
		}
	}
	
	return time.Time{}
}
