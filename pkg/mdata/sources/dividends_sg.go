package sources

import (
	"fmt"
	"math"
	"net/http"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type DividendsSg struct{}

func NewDividendsSg() *DividendsSg {
	return &DividendsSg{}
}

// GetHistoricalData implements types.DataSource.
func (d *DividendsSg) GetHistoricalData(symbol string, fromDate int64, toDate int64) ([]*types.StockData, error) {
	panic("unimplemented")
}

// GetStockPrice implements types.DataSource.
func (d *DividendsSg) GetStockPrice(symbol string) (*types.StockData, error) {
	panic("unimplemented")
}

func (d *DividendsSg) GetDividends(ticker string) ([]types.Dividend, error) {
	logger := logging.GetLogger()

	url := fmt.Sprintf("https://www.dividends.sg/view/%s", ticker)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dividends: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch dividends: status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Use map to aggregate dividends by date
	dividendMap := make(map[string]float64)

	doc.Find("table.table-bordered tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if i == 0 {
			return
		}

		// Extract date and amount
		cells := s.Find("td")
		clen := cells.Length()
		amountIdx := 3
		dateIdx := 4
		if clen < 4 { // Ensure we have enough cells
			return
		} else if clen == 4 {
			amountIdx = 0
			dateIdx = 1
		}

		// Amount
		amountStr := cells.Eq(amountIdx).Text()
		if amountStr == "-" {
			logger.Warn("skipping non-dividend event")
			amountStr = "0"
		}

		// Parse amount, removing "SGD" prefix if present
		amountStr = strings.TrimSpace(strings.TrimPrefix(amountStr, "SGD"))
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return
		}

		// Date (ex-date)
		dateStr := cells.Eq(dateIdx).Text()
		if dateStr == "-" || dateStr == "" {
			logger.Warn("empty date for dividend, please check if the website has changed")
			return
		}

		// Add amount to existing date or create new entry
		dividendMap[dateStr] += amount
	})

	// Convert map to sorted slice
	var dividends []types.Dividend
	for date, amount := range dividendMap {
		dividends = append(dividends, types.Dividend{Date: date, Amount: math.Round(amount*1000) / 1000})
	}

	// Sort dividends by date string (works because format is yyyy-mm-dd)
	sort.Slice(dividends, func(i, j int) bool {
		return dividends[i].Date < dividends[j].Date
	})

	return dividends, nil
}
