package sources

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type DividendsSg struct{}

func NewDividendsSg() *DividendsSg {
	return &DividendsSg{}
}

func (d *DividendsSg) FetchDividends(ticker string) (map[string]float64, error) {
	url := fmt.Sprintf("https://www.dividends.sg/view/%s", ticker)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dividends: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	dividends := make(map[string]float64)

	// Find the table with dividend data (it's the first table with bordered class)
	doc.Find("table.table-bordered tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if i == 0 {
			return
		}

		// Extract date and amount
		cells := s.Find("td")
		if cells.Length() < 4 { // Ensure we have enough cells
			return
		}

		// Amount is in column 4
		amountStr := cells.Eq(3).Text()
		if amountStr == "-" {
			return
		}

		// Parse amount, removing "SGD" prefix if present
		amountStr = strings.TrimSpace(strings.TrimPrefix(amountStr, "SGD"))
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return
		}

		// Date is in column 5
		dateStr := cells.Eq(4).Text()
		if dateStr == "-" {
			return
		}

		// Aggregate multiple dividends on the same date
		if existing, ok := dividends[dateStr]; ok {
			dividends[dateStr] = existing + amount
		} else {
			dividends[dateStr] = amount
		}
	})

	return dividends, nil
}
