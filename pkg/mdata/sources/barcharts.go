package sources

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

const (
	barchartsBaseURL      = "https://www.barchart.com/proxies/core-api/v1/historical/get"
	barchartsDefaultLimit = 520
)

type BarchartsSource struct {
	client *http.Client
	logger *logging.Logger
	jar    *cookiejar.Jar

	xsrfToken      string
	laravelToken   string
	tokenFailures  int
	lastTokenFetch time.Time
}

// NewBarcharts creates a new Barcharts futures data source
func NewBarcharts() *BarchartsSource {
	jar, _ := cookiejar.New(nil)
	return &BarchartsSource{
		client: &http.Client{Timeout: 15 * time.Second, Jar: jar},
		logger: logging.GetLogger(),
		jar:    jar,
	}
}

type BarchartsHistoricalRecord struct {
	Ticker        string
	TradeDate     string
	OpenPrice     float64
	HighPrice     float64
	LowPrice      float64
	LastPrice     float64
	PriceChange   float64
	PercentChange float64
	Volume        int64
	OpenInterest  int64
	Timestamp     int64
}

type barchartsHistoricalResponse struct {
	Count int `json:"count"`
	Total int `json:"total"`
	Data  []struct {
		Raw struct {
			TradeTime     string  `json:"tradeTime"`
			OpenPrice     float64 `json:"openPrice"`
			HighPrice     float64 `json:"highPrice"`
			LowPrice      float64 `json:"lowPrice"`
			LastPrice     float64 `json:"lastPrice"`
			PriceChange   float64 `json:"priceChange"`
			PercentChange float64 `json:"percentChange"`
			Volume        int64   `json:"volume"`
			OpenInterest  int64   `json:"openInterest"`
		} `json:"raw"`
	} `json:"data"`
}

// GetHistoricalData fetches futures historical data from Barcharts.
func (src *BarchartsSource) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*BarchartsHistoricalRecord, error) {
	if err := src.ensureTokens(); err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("symbol", ticker)
	query.Set("fields", "tradeTime.format(m/d/Y),openPrice,highPrice,lowPrice,lastPrice,priceChange,percentChange,volume,openInterest,symbolCode,symbolType")
	query.Set("type", "eod")
	query.Set("orderBy", "tradeTime")
	query.Set("orderDir", "desc")
	query.Set("limit", fmt.Sprintf("%d", barchartsDefaultLimit))
	query.Set("meta", "field.shortName,field.type,field.description")
	query.Set("raw", "1")

	requestURL := fmt.Sprintf("%s?%s", barchartsBaseURL, query.Encode())
	resp, err := src.doBarchartsRequest(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		if src.tokenFailures >= 1 {
			return nil, fmt.Errorf("barcharts unauthorized after token refresh")
		}
		src.tokenFailures++
		if err := src.refreshTokens(); err != nil {
			return nil, err
		}
		resp.Body.Close()
		resp, err = src.doBarchartsRequest(requestURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("barcharts unauthorized after token refresh")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("barcharts returned status code: %d", resp.StatusCode)
	}

	src.tokenFailures = 0

	bodyBytes, err := readBarchartsBody(resp)
	if err != nil {
		return nil, err
	}

	var payload barchartsHistoricalResponse
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode barcharts response: %w", err)
	}

	results := make([]*BarchartsHistoricalRecord, 0, len(payload.Data))
	for _, entry := range payload.Data {
		tradeDate := entry.Raw.TradeTime
		parsedTime, err := time.Parse("2006-01-02", tradeDate)
		if err != nil {
			src.logger.Errorf("failed to parse barcharts trade date %s: %v", tradeDate, err)
			continue
		}

		timestamp := parsedTime.Unix()
		if (fromDate > 0 && timestamp < fromDate) || (toDate > 0 && timestamp > toDate) {
			continue
		}

		results = append(results, &BarchartsHistoricalRecord{
			Ticker:        ticker,
			TradeDate:     tradeDate,
			OpenPrice:     entry.Raw.OpenPrice,
			HighPrice:     entry.Raw.HighPrice,
			LowPrice:      entry.Raw.LowPrice,
			LastPrice:     entry.Raw.LastPrice,
			PriceChange:   entry.Raw.PriceChange,
			PercentChange: entry.Raw.PercentChange,
			Volume:        entry.Raw.Volume,
			OpenInterest:  entry.Raw.OpenInterest,
			Timestamp:     timestamp,
		})
	}

	return results, nil
}

func (src *BarchartsSource) doBarchartsRequest(requestURL string) (*http.Response, error) {
	req, err := common.NewBrowserLikeRequest("GET", requestURL)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("x-xsrf-token", src.xsrfToken)
	req.AddCookie(&http.Cookie{Name: "laravel_token", Value: src.laravelToken})

	resp, err := common.DoWithRetry(src.client, req, 2, 2500*time.Millisecond, false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch barcharts data: %w", err)
	}

	return resp, nil
}

func (src *BarchartsSource) ensureTokens() error {
	if src.xsrfToken != "" && src.laravelToken != "" {
		return nil
	}
	return src.refreshTokens()
}

func (src *BarchartsSource) refreshTokens() error {
	if src.jar == nil {
		jar, _ := cookiejar.New(nil)
		src.jar = jar
		src.client.Jar = jar
	}

	homeURL, _ := url.Parse("https://www.barchart.com/")
	req, err := common.NewBrowserLikeRequest("GET", homeURL.String())
	if err != nil {
		return err
	}
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := common.DoWithRetry(src.client, req, 2, 2000*time.Millisecond, false)
	if err != nil {
		return fmt.Errorf("failed to fetch barcharts homepage: %w", err)
	}
	resp.Body.Close()

	cookies := src.jar.Cookies(homeURL)
	var xsrfToken string
	var laravelToken string
	for _, cookie := range cookies {
		switch cookie.Name {
		case "XSRF-TOKEN":
			decoded, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				xsrfToken = cookie.Value
			} else {
				xsrfToken = decoded
			}
		case "laravel_token":
			laravelToken = cookie.Value
		}
	}

	if xsrfToken == "" || laravelToken == "" {
		return fmt.Errorf("failed to retrieve barcharts tokens")
	}

	src.xsrfToken = xsrfToken
	src.laravelToken = laravelToken
	src.lastTokenFetch = time.Now()
	return nil
}

func readBarchartsBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read barcharts response body: %w", err)
	}

	return bodyBytes, nil
}
