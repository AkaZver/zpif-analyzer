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

func (p *MoexParser) SearchSecurity(isin string) (*MoexSecurity, error) {
	// Получаем информацию о бумаге
	url := fmt.Sprintf("%s/iss/securities/%s.json", p.baseURL, isin)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MOEX security info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MOEX API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Парсим ответ
	var secInfo struct {
		Description struct {
			SecID string `json:"sec_id"`
			Name  string `json:"name"`
			ISIN  string `json:"isin"`
		} `json:"description"`
		Boards struct {
			Columns []string        `json:"columns"`
			Data    [][]interface{} `json:"data"`
		} `json:"boards"`
	}

	if err := json.Unmarshal(body, &secInfo); err != nil {
		return nil, fmt.Errorf("failed to parse MOEX security info: %w", err)
	}

	// Извлекаем все board'ы, где торговалась бумага (market = "shares")
	var boards []string
	boardIdx := -1
	marketIdx := -1

	for i, col := range secInfo.Boards.Columns {
		switch col {
		case "boardid":
			boardIdx = i
		case "market":
			marketIdx = i
		}
	}

	for _, row := range secInfo.Boards.Data {
		if boardIdx >= 0 && boardIdx < len(row) {
			boardID, _ := row[boardIdx].(string)
			market := ""

			if marketIdx >= 0 && marketIdx < len(row) {
				market, _ = row[marketIdx].(string)
			}

			// Берём только board'ы с market = "shares" (основные торговые board'ы)
			if market == "shares" && boardID != "" {
				boards = append(boards, boardID)
			}
		}
	}

	if len(boards) == 0 {
		return nil, fmt.Errorf("no trading boards found for ISIN %s", isin)
	}

	return &MoexSecurity{
		SecID:  isin,
		Boards: boards,
		ISIN:   isin,
		Name:   secInfo.Description.Name,
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
