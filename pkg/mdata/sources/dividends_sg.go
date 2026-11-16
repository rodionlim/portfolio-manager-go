package sources

import (
	"fmt"
	"math"
	"net/http"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
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

	price, err := parseDividendsSgPrice(doc, ticker)
	if err != nil {
		return nil, err
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

func parseDividendsSgPrice(doc *goquery.Document, ticker string) (float64, error) {
	quote := doc.Find(".company-quote-line > span").First()
	price, err := strconv.ParseFloat(strings.TrimSpace(quote.Find("strong").First().Text()), 64)
	if err != nil || price <= 0 || math.IsNaN(price) || math.IsInf(price, 0) {
		return 0, fmt.Errorf("could not find price for %s in company quote", ticker)
	}
	return price, nil
}

func (src *DividendsSg) GetDividendsMetadata(ticker string, withholdingTax float64) ([]types.DividendsMetadata, error) {
	// Fetch new dividends from dividends.sg, then merge it with custom dividends if any
	logger := logging.GetLogger()

	// Check cache first
	if cachedData, found := src.cache.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker)); found {
		logger.Info("Returning cached dividends data for ticker:", ticker)
		return cachedData.([]types.DividendsMetadata), nil
	}

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

	dividends, err := parseDividendsSgMetadata(doc, ticker, withholdingTax)
	if err != nil {
		return nil, err
	}
	dividends = src.withDividendSource(dividends, types.DividendSourceOfficial)

	if src.db != nil {
		return src.upsertOfficialDividendsMetadata(ticker, dividends)
	}

	src.cache.Set(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), dividends, cache.DefaultExpiration)

	return dividends, nil
}
