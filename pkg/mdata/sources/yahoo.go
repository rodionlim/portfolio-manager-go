package sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"time"

	"github.com/patrickmn/go-cache"
)

type yahooFinance struct {
	client *http.Client
	cache  *cache.Cache
	logger *logging.Logger
}

// NewYahooFinance creates a new Yahoo Finance data source
func NewYahooFinance() types.DataSource {
	return &yahooFinance{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:  cache.New(5*time.Minute, 10*time.Minute),
		logger: logging.GetLogger(),
	}
}

// GetDividends implements types.DataSource.
func (src *yahooFinance) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	panic("unimplemented")
}

func (src *yahooFinance) GetStockPrice(ticker string) (*types.StockData, error) {
	if cachedData, found := src.cache.Get(ticker); found {
		src.logger.Infof("Returning cached data for ticker: %s", ticker)
		return cachedData.(*types.StockData), nil
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", ticker)

	resp, err := src.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo finance API returned status code: %d", resp.StatusCode)
	}

	// Parse response and create StockData
	var response struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency string  `json:"currency"`
					Symbol   string  `json:"symbol"`
					Price    float64 `json:"regularMarketPrice"`
				} `json:"meta"`
			} `json:"result"`
		} `json:"chart"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", ticker)
	}

	result := response.Chart.Result[0]
	stockData := &types.StockData{
		Ticker:    result.Meta.Symbol,
		Price:     result.Meta.Price,
		Currency:  result.Meta.Currency,
		Timestamp: time.Now().Unix(),
	}

	src.cache.Set(ticker, stockData, cache.DefaultExpiration)

	return stockData, nil
}

func (src *yahooFinance) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.StockData, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d",
		ticker, fromDate, toDate)

	resp, err := src.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo finance API returned status code: %d", resp.StatusCode)
	}

	// Parse response and create StockData array (Commented out most fields for brevity)
	var response struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency string `json:"currency"`
					Symbol   string `json:"symbol"`
					// ExchangeName         string   `json:"exchangeName"`
					// FullExchangeName     string   `json:"fullExchangeName"`
					// InstrumentType       string   `json:"instrumentType"`
					// FirstTradeDate       int64    `json:"firstTradeDate"`
					// RegularMarketTime    int64    `json:"regularMarketTime"`
					// HasPrePostMarketData bool     `json:"hasPrePostMarketData"`
					// GMTOffset            int      `json:"gmtoffset"`
					// Timezone             string   `json:"timezone"`
					// ExchangeTimezoneName string   `json:"exchangeTimezoneName"`
					// RegularMarketPrice   float64  `json:"regularMarketPrice"`
					// FiftyTwoWeekHigh     float64  `json:"fiftyTwoWeekHigh"`
					// FiftyTwoWeekLow      float64  `json:"fiftyTwoWeekLow"`
					// RegularMarketDayHigh float64  `json:"regularMarketDayHigh"`
					// RegularMarketDayLow  float64  `json:"regularMarketDayLow"`
					// RegularMarketVolume  int64    `json:"regularMarketVolume"`
					// LongName             string   `json:"longName"`
					// ShortName            string   `json:"shortName"`
					// ChartPreviousClose   float64  `json:"chartPreviousClose"`
					// PriceHint            int      `json:"priceHint"`
					// DataGranularity      string   `json:"dataGranularity"`
					// Range                string   `json:"range"`
					// ValidRanges          []string `json:"validRanges"`
				} `json:"meta"`
				Timestamp  []int64 `json:"timestamp"`
				Indicators struct {
					Quote []struct {
						Open   []float64 `json:"open"`
						High   []float64 `json:"high"`
						Low    []float64 `json:"low"`
						Close  []float64 `json:"close"`
						Volume []int64   `json:"volume"`
					} `json:"quote"`
					Adjclose []struct {
						Adjclose []float64 `json:"adjclose"`
					} `json:"adjclose"`
				} `json:"indicators"`
			} `json:"result"`
			Error interface{} `json:"error"`
		} `json:"chart"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no historical data found for ticker: %s", ticker)
	}

	result := response.Chart.Result[0]
	data := make([]*types.StockData, len(result.Timestamp))

	for i := range result.Timestamp {
		data[i] = &types.StockData{
			Ticker:    result.Meta.Symbol,
			Price:     result.Indicators.Quote[0].Close[i],
			Currency:  result.Meta.Currency,
			Timestamp: result.Timestamp[i],
		}
	}

	return data, nil
}
