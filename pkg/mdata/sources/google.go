package sources

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"portfolio-manager/pkg/types"
	"strconv"
	"time"

	"golang.org/x/net/html"
)

type googleFinance struct {
	client *http.Client
}

// NewGoogleFinance creates a new Google Finance data source
func NewGoogleFinance() types.DataSource {
	return &googleFinance{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *googleFinance) GetStockPrice(ticker string) (*types.StockData, error) {
	// Google Finance URL (note: this might need adjustments as Google doesn't provide an official API)
	url := fmt.Sprintf("https://www.google.com/finance/quote/%s", ticker)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to mimic browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google finance returned status code: %d", resp.StatusCode)
	}

	// Parse HTML
	price, err := extractPrice(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract price: %w", err)
	}

	return &types.StockData{
		Ticker:    ticker,
		Price:     price,
		Currency:  "USD",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (g *googleFinance) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.StockData, error) {
	return nil, errors.New("historical data not yet implemented for google data source")
}

// extractPrice helper function to extract price from HTML response
func extractPrice(r io.Reader) (float64, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return 0, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var price string
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "div" {
			for _, attr := range node.Attr {
				if attr.Key == "data-last-price" {
					price = attr.Val
					return
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			crawler(c)
		}
	}
	crawler(doc)

	if price == "" {
		return 0, fmt.Errorf("price not found in HTML")
	}

	return strconv.ParseFloat(price, 64)
}