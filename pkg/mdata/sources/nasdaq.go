package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"sort"
	"time"

	"github.com/patrickmn/go-cache"
)

type NasdaqSource struct {
	BaseDividendSource
}

func NewNasdaq(db dal.Database) *NasdaqSource {
	return &NasdaqSource{
		BaseDividendSource: BaseDividendSource{
			db:    db,
			cache: cache.New(24*time.Hour, 1*time.Hour),
		},
	}
}

// nasdaqDividendResponse represents the response from Nasdaq API
type nasdaqDividendResponse struct {
	Data struct {
		Dividends struct {
			Rows []struct {
				ExOrEffDate     string `json:"exOrEffDate"`
				Type            string `json:"type"`
				Amount          string `json:"amount"`
				DeclarationDate string `json:"declarationDate"`
				RecordDate      string `json:"recordDate"`
				PaymentDate     string `json:"paymentDate"`
				Currency        string `json:"currency"`
			} `json:"rows"`
		} `json:"dividends"`
	} `json:"data"`
}

// GetAssetPrice implements types.DataSource.
func (src *NasdaqSource) GetAssetPrice(ticker string) (*types.AssetData, error) {
	panic("unimplemented")
}

// GetHistoricalData implements types.DataSource.
func (src *NasdaqSource) GetHistoricalData(ticker string, fromDate int64, toDate int64) ([]*types.AssetData, error) {
	panic("unimplemented")
}

// GetDividendsMetadata fetches dividend data from Nasdaq API
func (src *NasdaqSource) GetDividendsMetadata(ticker string, withholdingTax float64) ([]types.DividendsMetadata, error) {
	logger := logging.GetLogger()

	// Check cache first
	if cachedData, found := src.cache.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker)); found {
		logger.Info("Returning cached dividends data for ticker:", ticker)
		return cachedData.([]types.DividendsMetadata), nil
	}

	officialDividendsMetadata, _ := src.getSingleDividendsMetadata(ticker, false)

	url := fmt.Sprintf("https://api.nasdaq.com/api/quote/%s/dividends?assetclass=stocks", ticker)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dividends: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch dividends: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var dividendResp nasdaqDividendResponse
	if err := json.Unmarshal(body, &dividendResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Use map to aggregate dividends by date
	dividendMap := make(map[string]float64)

	for _, row := range dividendResp.Data.Dividends.Rows {
		// Only process cash dividends
		if row.Type != "Cash" {
			continue
		}

		// Parse amount, removing "$" prefix if present
		amountStr := row.Amount
		if len(amountStr) > 0 && amountStr[0] == '$' {
			amountStr = amountStr[1:]
		}

		var amount float64
		_, err := fmt.Sscanf(amountStr, "%f", &amount)
		if err != nil {
			logger.Warnf("failed to parse amount %s: %v", row.Amount, err)
			continue
		}

		// Convert date from MM/DD/YYYY to YYYY-MM-DD
		exDate := convertDateFormat(row.ExOrEffDate)
		if exDate == "" {
			logger.Warnf("invalid date format: %s", row.ExOrEffDate)
			continue
		}

		// Add amount to existing date or create new entry
		dividendMap[exDate] += amount
	}

	// Convert map to sorted slice
	var dividends []types.DividendsMetadata
	for date, amount := range dividendMap {
		dividends = append(dividends, types.DividendsMetadata{
			Ticker:         ticker,
			ExDate:         date,
			Amount:         math.Round(amount*10000) / 10000, // Round to 4 decimal places
			WithholdingTax: withholdingTax,
		})
	}

	// Sort dividends by date
	sort.Slice(dividends, func(i, j int) bool {
		return dividends[i].ExDate < dividends[j].ExDate
	})

	if src.db != nil {
		// Store in database if we have new data
		if len(dividends) > len(officialDividendsMetadata) {
			logger.Infof("New dividends for ticker %s, storing into database", ticker)
			dividends, err = src.StoreDividendsMetadata(ticker, dividends, false)
		}
	}

	// Store in cache
	src.cache.Set(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), dividends, cache.DefaultExpiration)

	return dividends, err
}

// convertDateFormat converts MM/DD/YYYY to YYYY-MM-DD
func convertDateFormat(dateStr string) string {
	t, err := time.Parse("01/02/2006", dateStr)
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02")
}
