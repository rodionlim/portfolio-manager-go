package sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"strconv"
	"strings"
	"time"

	"portfolio-manager/pkg/common"

	"github.com/PuerkitoBio/goquery"
	"github.com/patrickmn/go-cache"
)

type yahooFinance struct {
	BaseDividendSource
	client *http.Client
	logger *logging.Logger
}

// NewYahooFinance creates a new Yahoo Finance data source
func NewYahooFinance(db dal.Database) types.DataSource {
	return &yahooFinance{
		BaseDividendSource: BaseDividendSource{
			db:    db,
			cache: cache.New(5*time.Minute, 10*time.Minute),
		},
		client: &http.Client{
			Timeout: 15 * time.Second, // Increased timeout for better reliability
		},
		logger: logging.GetLogger(),
	}
}

// GetDividends implements types.DataSource.
func (src *yahooFinance) GetDividendsMetadata(ticker string, withholdingTax float64) ([]types.DividendsMetadata, error) {
	// Check cache first
	if cachedData, found := src.cache.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker)); found {
		src.logger.Info("Returning cached dividends data for ticker:", ticker)
		return cachedData.([]types.DividendsMetadata), nil
	}

	// Fetch dividends from Yahoo Finance
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/history/?period1=511108200&period2=%d&filter=div", ticker, time.Now().Unix())

	req, err := common.NewBrowserLikeRequest("GET", url)
	if err != nil {
		return nil, err
	}

	resp, err := src.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo finance API returned status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var dividends []types.DividendsMetadata
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		date := s.Find("td").Eq(0).Text()
		dividend := s.Find("td").Eq(1).Text()
		if dividend != "" && strings.Contains(dividend, "Dividend") {
			amount, err := strconv.ParseFloat(strings.TrimSpace(strings.Replace(dividend, "Dividend", "", -1)), 64)
			if err != nil {
				src.logger.Errorf("failed to parse dividend amount: %v", err)
				return
			}
			date, err = common.ConvertDateFormat(date, "Jan 2, 2006", "2006-01-02")
			if err != nil {
				src.logger.Errorf("failed to convert dividend date: %v", err)
				return
			}
			dividends = append(dividends, types.DividendsMetadata{
				Ticker:         ticker,
				ExDate:         date,
				Amount:         amount,
				WithholdingTax: withholdingTax,
			})
		}
	})

	// Store in database if we have new data
	if src.db != nil {
		existingDividends, _ := src.getSingleDividendsMetadata(ticker, false)
		if len(dividends) > len(existingDividends) {
			src.logger.Infof("New dividends for ticker %s, storing into database", ticker)
			dividends, err = src.StoreDividendsMetadata(ticker, dividends, false)
		}
	}

	return dividends, err
}

func (src *yahooFinance) GetAssetPrice(ticker string) (*types.AssetData, error) {
	if cachedData, found := src.cache.Get(ticker); found {
		src.logger.Infof("Returning cached data for ticker: %s", ticker)
		return cachedData.(*types.AssetData), nil
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", ticker)
	src.logger.Debugf("Fetching asset price from: %s", url)

	req, err := common.NewBrowserLikeRequest("GET", url)
	if err != nil {
		return nil, err
	}

	resp, err := src.client.Do(req)
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
	stockData := &types.AssetData{
		Ticker:    result.Meta.Symbol,
		Price:     result.Meta.Price,
		Currency:  result.Meta.Currency,
		Timestamp: time.Now().Unix(),
	}

	src.cache.Set(ticker, stockData, cache.DefaultExpiration)

	return stockData, nil
}

func (src *yahooFinance) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.AssetData, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d",
		ticker, fromDate, toDate)

	src.logger.Debugf("Fetching historical data from: %s", url)

	// Apply rate limiting before making the request
	common.GetYahooRateLimiter().Wait()

	req, err := common.NewBrowserLikeRequest("GET", url)
	if err != nil {
		return nil, err
	}

	resp, err := src.client.Do(req)
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
	data := make([]*types.AssetData, len(result.Timestamp))

	for i := range result.Timestamp {
		data[i] = &types.AssetData{
			Ticker:    result.Meta.Symbol,
			Price:     result.Indicators.Quote[0].Close[i],
			Currency:  result.Meta.Currency,
			Timestamp: result.Timestamp[i],
		}
	}

	return data, nil
}
