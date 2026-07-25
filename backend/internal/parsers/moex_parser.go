package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MoexMarketData struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
	Value  float64
}

type MoexSecurity struct {
	SecID  string
	Boards []string // Список всех board'ов, на которых торговалась бумага
	ISIN   string
	Name   string
}

type MoexParser struct {
	client  *http.Client
	baseURL string
}

func NewMoexParser() *MoexParser {
	return &MoexParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://iss.moex.com",
	}
}

type moexResponse struct {
	MarketData struct {
		Columns []string        `json:"columns"`
		Data    [][]interface{} `json:"data"`
	} `json:"marketdata"`
	History struct {
		Columns []string        `json:"columns"`
		Data    [][]interface{} `json:"data"`
	} `json:"history"`
}

func (p *MoexParser) GetCurrentPrice(ticker string) (*MoexMarketData, error) {
	return p.GetCurrentPriceWithBoard(ticker, "TQBR")
}

func (p *MoexParser) GetCurrentPriceWithBoard(secID, board string) (*MoexMarketData, error) {
	url := fmt.Sprintf("%s/iss/engines/stock/markets/shares/boards/%s/securities/%s.json", p.baseURL, board, secID)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MOEX data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MOEX API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var moexResp moexResponse
	if err := json.Unmarshal(body, &moexResp); err != nil {
		return nil, fmt.Errorf("failed to parse MOEX response: %w", err)
	}

	if len(moexResp.MarketData.Data) == 0 {
		return nil, fmt.Errorf("no market data found for secid %s on board %s", secID, board)
	}

	data := moexResp.MarketData.Data[0]
	columns := moexResp.MarketData.Columns

	result := &MoexMarketData{
		Date: time.Now(),
	}

	for i, col := range columns {
		if i >= len(data) {
			break
		}
		switch col {
		case "LAST":
			if val, ok := data[i].(float64); ok {
				result.Close = val
			}
		case "OPEN":
			if val, ok := data[i].(float64); ok {
				result.Open = val
			}
		case "HIGH":
			if val, ok := data[i].(float64); ok {
				result.High = val
			}
		case "LOW":
			if val, ok := data[i].(float64); ok {
				result.Low = val
			}
		case "VOLUME":
			if val, ok := data[i].(float64); ok {
				result.Volume = int64(val)
			}
		case "VALUE":
			if val, ok := data[i].(float64); ok {
				result.Value = val
			}
		}
	}

	return result, nil
}

func (p *MoexParser) SearchSecurity(query string) (*MoexSecurity, error) {
	// Шаг 1: Поиск бумаги через search endpoint
	searchURL := fmt.Sprintf("%s/iss/securities.json?q=%s", p.baseURL, query)

	resp, err := p.client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search MOEX security: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MOEX search API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search response: %w", err)
	}

	var searchResp struct {
		Securities struct {
			Columns []string        `json:"columns"`
			Data    [][]interface{} `json:"data"`
		} `json:"securities"`
	}

	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	if len(searchResp.Securities.Data) == 0 {
		return nil, fmt.Errorf("no securities found for query %s", query)
	}

	// Найти индексы колонок
	columns := searchResp.Securities.Columns
	secidIdx := -1
	isinIdx := -1
	shortnameIdx := -1

	for i, col := range columns {
		switch col {
		case "secid":
			secidIdx = i
		case "isin":
			isinIdx = i
		case "shortname":
			shortnameIdx = i
		}
	}

	if secidIdx == -1 {
		return nil, fmt.Errorf("secid column not found in search response")
	}

	// Найти подходящую бумагу
	var foundSecID, foundISIN, foundName string
	for _, row := range searchResp.Securities.Data {
		if len(row) <= secidIdx {
			continue
		}

		secid, _ := row[secidIdx].(string)
		isin := ""
		if isinIdx >= 0 && isinIdx < len(row) {
			isin, _ = row[isinIdx].(string)
		}

		// Совпадение по ISIN или secid (тикеру)
		if secid == query || isin == query {
			foundSecID = secid
			foundISIN = isin
			if shortnameIdx >= 0 && shortnameIdx < len(row) {
				foundName, _ = row[shortnameIdx].(string)
			}
			break
		}
	}

	if foundSecID == "" {
		return nil, fmt.Errorf("no matching security found for query %s", query)
	}

	// Шаг 2: Получить все boards для найденной бумаги
	boardsURL := fmt.Sprintf("%s/iss/securities/%s.json", p.baseURL, foundSecID)

	resp2, err := p.client.Get(boardsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch boards info: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MOEX boards API returned status %d", resp2.StatusCode)
	}

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read boards response: %w", err)
	}

	var boardsResp struct {
		Boards struct {
			Columns []string        `json:"columns"`
			Data    [][]interface{} `json:"data"`
		} `json:"boards"`
	}

	if err := json.Unmarshal(body2, &boardsResp); err != nil {
		return nil, fmt.Errorf("failed to parse boards response: %w", err)
	}

	// Извлечь все board'ы с market = "shares"
	var boards []string
	boardIdx := -1
	marketIdx := -1

	for i, col := range boardsResp.Boards.Columns {
		switch col {
		case "boardid":
			boardIdx = i
		case "market":
			marketIdx = i
		}
	}

	for _, row := range boardsResp.Boards.Data {
		if boardIdx >= 0 && boardIdx < len(row) {
			boardID, _ := row[boardIdx].(string)
			market := ""

			if marketIdx >= 0 && marketIdx < len(row) {
				market, _ = row[marketIdx].(string)
			}

		// Берём board'ы с market = "shares" или "sharesndm" (основные торговые board'ы и ОТС)
		if (market == "shares" || market == "sharesndm") && boardID != "" {
				boards = append(boards, boardID)
			}
		}
	}

	if len(boards) == 0 {
		return nil, fmt.Errorf("no trading boards found for security %s", foundSecID)
	}

	return &MoexSecurity{
		SecID:  foundSecID,
		Boards: boards,
		ISIN:   foundISIN,
		Name:   foundName,
	}, nil
}

func (p *MoexParser) GetPriceHistory(ticker string) ([]MoexMarketData, error) {
	return p.GetPriceHistoryWithBoard(ticker, "TQBR")
}

func (p *MoexParser) GetPriceHistoryWithBoard(secID, board string) ([]MoexMarketData, error) {
	var allData []MoexMarketData
	start := 0
	limit := 100

	for {
		url := fmt.Sprintf("%s/iss/history/engines/stock/markets/shares/boards/%s/securities/%s.json?start=%d&limit=%d", p.baseURL, board, secID, start, limit)

		resp, err := p.client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch MOEX history: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("MOEX API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		var moexResp moexResponse
		if err := json.Unmarshal(body, &moexResp); err != nil {
			return nil, fmt.Errorf("failed to parse MOEX response: %w", err)
		}

		if len(moexResp.History.Data) == 0 {
			break
		}

		columns := moexResp.History.Columns
		for _, row := range moexResp.History.Data {
			data := &MoexMarketData{}

			for i, col := range columns {
				if i >= len(row) {
					break
				}
				switch col {
				case "TRADEDATE":
					if val, ok := row[i].(string); ok {
						if t, err := time.Parse("2006-01-02", val); err == nil {
							data.Date = t
						}
					}
				case "CLOSE":
					if val, ok := row[i].(float64); ok && val > 0 {
						data.Close = val
					}
				case "LEGALCLOSEPRICE":
					// Fallback: используем LEGALCLOSEPRICE если CLOSE равен 0
					if val, ok := row[i].(float64); ok && val > 0 && data.Close == 0 {
						data.Close = val
					}
				case "WAPRICE":
					// Fallback: используем WAPRICE если CLOSE и LEGALCLOSEPRICE равны 0
					if val, ok := row[i].(float64); ok && val > 0 && data.Close == 0 {
						data.Close = val
					}
				case "OPEN":
					if val, ok := row[i].(float64); ok {
						data.Open = val
					}
				case "HIGH":
					if val, ok := row[i].(float64); ok {
						data.High = val
					}
				case "LOW":
					if val, ok := row[i].(float64); ok {
						data.Low = val
					}
				case "VOLUME":
					if val, ok := row[i].(float64); ok {
						data.Volume = int64(val)
					}
				case "VALUE":
					if val, ok := row[i].(float64); ok {
						data.Value = val
					}
				}
			}

			if !data.Date.IsZero() && data.Close > 0 {
				allData = append(allData, *data)
			}
		}

		if len(moexResp.History.Data) < limit {
			break
		}

		start += limit
		time.Sleep(100 * time.Millisecond)
	}

	return allData, nil
}
