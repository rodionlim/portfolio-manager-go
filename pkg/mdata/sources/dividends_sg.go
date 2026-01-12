package sources

import (
	"fmt"
	"math"
	"net/http"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"sort"
	"strconv"
	"strings"
	"time"

	"slices"

	"github.com/PuerkitoBio/goquery"
	"github.com/patrickmn/go-cache"
)

type DividendsSg struct {
	BaseDividendSource
}

func NewDividendsSg(db dal.Database) *DividendsSg {
	return &DividendsSg{
		BaseDividendSource: BaseDividendSource{
			db:    db,
			cache: cache.New(24*time.Hour, 1*time.Hour),
		},
	}
}

// GetHistoricalData implements types.DataSource.
func (src *DividendsSg) GetHistoricalData(ticker string, fromDate int64, toDate int64) ([]*types.AssetData, bool, error) {
	panic("unimplemented")
}

// GetAssetPrice implements types.DataSource.
func (src *DividendsSg) GetAssetPrice(ticker string) (*types.AssetData, error) {
	logger := logging.GetLogger()

	// Check cache first
	if cachedData, found := src.cache.Get(ticker); found {
		logger.Info("Returning cached price data for ticker:", ticker)
		return cachedData.(*types.AssetData), nil
	}

	url := fmt.Sprintf("https://www.dividends.sg/view/%s", ticker)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch prices: status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var price float64
	var priceFound bool

	// Find h4 elements containing span with price
	doc.Find("h4").Each(func(i int, s *goquery.Selection) {
		// If we've already found the price, skip
		if priceFound {
			return
		}

		// Look for a span inside this h4
		s.Find("span.badge").Each(func(j int, span *goquery.Selection) {
			priceText := strings.TrimSpace(span.Text())

			// Try to parse as float
			p, err := strconv.ParseFloat(priceText, 64)
			if err == nil {
				price = p
				priceFound = true
			}
		})

		// Fallback: If span.badge not found, check if currency code exists in the h4 text
		if !priceFound {
			h4Text := s.Text()
			currencies := []string{"SGD", "USD", "HKD", "EUR", "GBP", "JPY", "MYR", "AUD", "CAD"}

			for _, currency := range currencies {
				if strings.Contains(h4Text, currency) {
					// Extract price after currency code
					// Example: "TEMASEK S$500M 1.8% B 261124\t(TEMB)\tSGD 1.013\t\n\t\u00a0\n\t +0.79% +0.01"
					parts := strings.Split(h4Text, currency)
					if len(parts) > 1 {
						// Get the part after currency code
						afterCurrency := strings.TrimSpace(parts[1])
						// Split by whitespace and get first token
						tokens := strings.Fields(afterCurrency)
						if len(tokens) > 0 {
							// Try to parse the first token as price
							p, err := strconv.ParseFloat(tokens[0], 64)
							if err == nil {
								price = p
								priceFound = true
								break
							}
						}
					}
				}
			}
		}
	})

	if !priceFound {
		return nil, fmt.Errorf("could not find price for %s", ticker)
	}

	// Create asset data object
	assetData := &types.AssetData{
		Ticker:    ticker,
		Price:     price,
		Currency:  "SGD",
		Timestamp: time.Now().Unix(),
	}

	// Store in cache
	src.cache.Set(ticker, assetData, cache.DefaultExpiration)

	return assetData, nil
}

func (src *DividendsSg) GetDividendsMetadata(ticker string, withholdingTax float64) ([]types.DividendsMetadata, error) {
	// Fetch new dividends from dividends.sg, then merge it with custom dividends if any
	logger := logging.GetLogger()

	// Check cache first
	if cachedData, found := src.cache.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker)); found {
		logger.Info("Returning cached dividends data for ticker:", ticker)
		return cachedData.([]types.DividendsMetadata), nil
	}

	officialDividendsMetadata, _ := src.GetSingleDividendsMetadataWithType(ticker, false)

	// Escape hatch for custom dividends with invalid dividend history on dividends.sg
	specialTickers := []string{"FCOT.SI"}
	if slices.Contains(specialTickers, ticker) {
		return src.GetSingleDividendsMetadataWithType(ticker, true)
	}

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

	isBond := false
	doc.Find("table.table-bordered tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row, and determine if product is bond or equity
		if i == 0 {
			cells := s.Find("th")
			if cells.Length() == 4 {
				isBond = true
			}
			return
		}

		// Extract date and amount (equity)
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

		if isBond {
			// e.g. https://www.dividends.sg/view/TEMB, ex-date and particulars column
			amountIdx = 3
			dateIdx = 1
		}

		// Amount
		amountStr := cells.Eq(amountIdx).Text()
		if amountStr == "-" {
			logger.Warn("skipping non-dividend event")
			amountStr = "0"
		}

		// Parse amount, removing "SGD" prefix if present, also parsing for % for bonds
		var amount float64
		if strings.Contains(amountStr, "%") {
			// check that amount string starts with Rate:, all other cases are not real dividends
			if !strings.HasPrefix(amountStr, "Rate:") {
				return
			}
			amountStr = strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(amountStr, "Rate: ")), "%", "")
			amount, err = strconv.ParseFloat(amountStr, 64)
			if err != nil {
				return
			}
			// here we make the assumption that the bond pays 2 times a year
			// TODO: this might not be true and needs a rework, thankfully, semiannual bonds are the most common
			amount = amount / 100 / 2
		} else {
			// handle USD denominated singapore stocks
			amountStr = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(amountStr, "SGD"), "USD"))
			amount, err = strconv.ParseFloat(amountStr, 64)
			if err != nil {
				return
			}
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
	var dividends []types.DividendsMetadata
	for date, amount := range dividendMap {
		dividends = append(dividends, types.DividendsMetadata{
			Ticker:         ticker,
			ExDate:         date,
			Amount:         math.Round(amount*10000) / 10000, // Round to 4 decimal places
			WithholdingTax: withholdingTax})                  // sg dividends have no withholding tax
	}

	// Sort dividends by date string (works because format is yyyy-mm-dd)
	sort.Slice(dividends, func(i, j int) bool {
		return dividends[i].ExDate < dividends[j].ExDate
	})

	if src.db != nil {
		// Store in database if we have new data
		if len(dividends) > len(officialDividendsMetadata) {
			logger.Infof("New dividends for ticker %s, storing into database", ticker)
			src.StoreDividendsMetadata(ticker, dividends, false)
		}
	}

	return src.GetSingleDividendsMetadata(ticker)
}
