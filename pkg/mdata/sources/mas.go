package sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Mas struct {
	client *http.Client
	db     dal.Database
	url    string
	logger *logging.Logger
}

func NewMas(db dal.Database) *Mas {
	return &Mas{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		db:     db,
		url:    "https://eservices.mas.gov.sg/statistics/api/v1/bondsandbills/m/listauctionbondsandbills?rows=1",
		logger: logging.GetLogger(),
	}
}

// GetHistoricalData implements types.DataSource. MAS Bills are always traded at par value.
func (src *Mas) GetHistoricalData(ticker string, fromDate int64, toDate int64) ([]*types.AssetData, error) {
	// SSB is always traded at par value
	var historicalData []*types.AssetData
	startDate := time.Unix(fromDate, 0)
	endDate := time.Unix(toDate, 0)

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			historicalData = append(historicalData, &types.AssetData{
				Ticker:    ticker,
				Price:     100.0,
				Currency:  "SGD",
				Timestamp: d.Unix(),
			})
		}
	}

	return historicalData, nil
}

// GetAssetPrice implements types.DataSource. MAS Bills are always traded at par value.
func (src *Mas) GetAssetPrice(ticker string) (*types.AssetData, error) {
	return &types.AssetData{
		Ticker:    ticker,
		Price:     100.0,
		Currency:  "SGD",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (src *Mas) GetDividendsMetadata(ticker string, withholdingTax float64) ([]types.DividendsMetadata, error) {
	if !common.IsSgTBill(ticker) {
		return nil, fmt.Errorf("invalid sg tbill ticker: %s", ticker)
	}

	// fetch from db, if exist, then don't need to hit the actual data source
	if src.db != nil {
		var dividends []types.DividendsMetadata
		src.db.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), &dividends)
		if len(dividends) > 0 {
			src.logger.Infof("Found coupons for ticker %s in database", ticker)
			return dividends, nil
		}
	}

	url := fmt.Sprintf("%s&filters=issue_code:%s", src.url, ticker)
	req, err := common.NewHttpRequestWithUserAgent("GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := src.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sg tbill interest rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch coupon payments: status code %d", resp.StatusCode)
	}

	// Parse response and create StockData
	var response struct {
		Result struct {
			Records []struct {
				IssueDate    string  `json:"issue_date"`
				CutoffPrice  float64 `json:"cutoff_price"`
				CutoffYield  float64 `json:"cutoff_yield"`
				MaturityDate string  `json:"maturity_date"`
			} `json:"records"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Result.Records) == 0 {
		return nil, fmt.Errorf("no data found for ticker: %s", ticker)
	}

	result := response.Result.Records[0]
	dividends := []types.DividendsMetadata{{
		Ticker:         ticker,
		ExDate:         result.IssueDate,
		Amount:         100 - result.CutoffPrice,
		Interest:       result.CutoffYield, // interest in percentage
		AvgInterest:    result.CutoffYield, // interest in percentage
		WithholdingTax: withholdingTax,
	}}

	// For issuance that are not found in leveldb, store it into level db
	if src.db != nil {
		err := src.db.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), &dividends)
		if err != nil {
			src.logger.Infof("New coupons for ticker %s, storing into database", ticker)
			src.db.Put(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), dividends)
		}

		// store ref data into the db to update maturity date
		tickerRef := rdata.TickerReference{
			ID:            ticker,
			Name:          ticker,
			Domicile:      "SG",
			Ccy:           "SGD",
			AssetClass:    rdata.AssetClassBonds,
			AssetSubClass: rdata.AssetSubClassGovies,
			MaturityDate:  result.MaturityDate,
		}
		src.logger.Infof("Updating maturity date for ticker %s to %s", ticker, result.MaturityDate)
		src.db.Put(fmt.Sprintf("%s:%s", types.ReferenceDataKeyPrefix, ticker), tickerRef)
	}

	return dividends, nil
}

// StoreDividendsMetadata stores either custom or official dividends metadata for a given ticker
func (src *Mas) StoreDividendsMetadata(ticker string, dividends []types.DividendsMetadata, isCustom bool) ([]types.DividendsMetadata, error) {
	panic("unimplemented")
}

// FetchBenchmarkInterestRates fetches benchmark interest rates from MAS for Singapore
// It scrapes the SORA (Singapore Overnight Rate Average) rates from the MAS website
// using ASP.NET form submission with viewstate management.
//
// Best practices for ASP.NET scraping:
// 1. Two-step process: GET initial page to extract viewstate, then POST with viewstate
// 2. Extract required fields: __VIEWSTATE, __VIEWSTATEGENERATOR, __EVENTVALIDATION
// 3. Maintain session: Use the same HTTP client for both requests
// 4. Parse HTML response: Use goquery to extract structured data
//
// Parameters:
// - country: Only "SG" (Singapore) is supported
// - points: Number of recent records to fetch (estimated as days, filtered post-fetch)
//
// Returns:
// - Array of InterestRates with SORA overnight rates
// - Error if country is unsupported or scraping fails
func (src *Mas) FetchBenchmarkInterestRates(country string, points int) ([]types.InterestRates, error) {
	if country != "SG" {
		return nil, fmt.Errorf("unsupported country: %s", country)
	}

	// Check database first for existing data
	dbKey := fmt.Sprintf("%s:%s", types.InterestRatesKeyPrefix, country)
	var cachedRates []types.InterestRates

	if src.db != nil {
		err := src.db.Get(dbKey, &cachedRates)
		if err == nil && len(cachedRates) > 0 {
			// Check if cached data is recent enough (contains data from current month/year)
			isRecentEnough := false
			currentMonthYear := time.Now().Format("Jan 2006")
			
			for _, rate := range cachedRates {
				// Parse the date to check if it's from the current month and year
				if rateTime, err := time.Parse("02 Jan 2006", rate.Date); err == nil {
					if rateTime.Format("Jan 2006") == currentMonthYear {
						isRecentEnough = true
						break
					}
				}
			}
			
			// Only use cached data if it's recent enough and has sufficient records
			if isRecentEnough {
				filteredRates := filterRecentRates(cachedRates, points)
				if len(filteredRates) >= points {
					src.logger.Infof("Found %d recent interest rates for %s in database", len(filteredRates), country)
					return filteredRates, nil
				}
			} else {
				src.logger.Infof("Cached data for %s is not recent enough, fetching fresh data", country)
			}
		}
	}

	baseURL := "https://eservices.mas.gov.sg/statistics/dir/DomesticInterestRates.aspx"

	// Step 1: GET the initial page to extract viewstate (with retry)
	initialReq, err := common.NewBrowserLikeRequest("GET", baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create initial request: %w", err)
	}

	initialResp, err := common.DoWithRetry(src.client, initialReq, 3, 500*time.Millisecond, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch initial page: %w", err)
	}
	defer initialResp.Body.Close()

	// Parse the HTML to extract viewstate and other required fields
	doc, err := goquery.NewDocumentFromReader(initialResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initial page: %w", err)
	}

	viewstate, _ := doc.Find("input[name='__VIEWSTATE']").Attr("value")
	viewstateGenerator, _ := doc.Find("input[name='__VIEWSTATEGENERATOR']").Attr("value")
	eventValidation, _ := doc.Find("input[name='__EVENTVALIDATION']").Attr("value")

	if viewstate == "" || viewstateGenerator == "" || eventValidation == "" {
		return nil, fmt.Errorf("failed to extract required viewstate fields")
	}

	// Find the SORA checkbox by looking for exact label match
	soraCheckboxName := ""
	doc.Find("input[type='checkbox']").Each(func(i int, checkbox *goquery.Selection) {
		checkboxId, exists := checkbox.Attr("id")
		if !exists {
			return
		}

		// Find the corresponding label
		labelSelector := fmt.Sprintf("label[for='%s']", checkboxId)
		label := doc.Find(labelSelector)
		if label.Length() > 0 {
			labelText := strings.TrimSpace(label.Text())
			// Look for exact match "SORA" to avoid "1-month Compounded SORA" etc.
			if labelText == "SORA" {
				if name, exists := checkbox.Attr("name"); exists {
					soraCheckboxName = name
				}
			}
		}
	})

	if soraCheckboxName == "" {
		return nil, fmt.Errorf("could not find SORA checkbox in form")
	}

	// Step 2: Prepare POST data - fetch wider range to ensure we have enough data
	currentYear := time.Now().Year()
	endYear := currentYear
	startYear := currentYear
	endMonth := int(time.Now().Month())

	// Estimate months needed based on points (assuming ~22 business days per month)
	estimatedMonths := (points / 22) + 3 // Add extra buffer for current month data availability
	if estimatedMonths < 2 {
		estimatedMonths = 2 // Minimum 2 months to ensure we get data
	}

	startMonth := endMonth - estimatedMonths + 1
	if startMonth <= 0 {
		startYear--
		startMonth = 12 + startMonth
		// Handle edge case where startMonth is still <= 0
		for startMonth <= 0 {
			startYear--
			startMonth += 12
		}
	}

	// Log the date range for debugging
	src.logger.Infof("Fetching MAS data from %d-%02d to %d-%02d (estimated months: %d)", 
		startYear, startMonth, endYear, endMonth, estimatedMonths)

	formData := url.Values{}
	formData.Set("__EVENTTARGET", "")
	formData.Set("__EVENTARGUMENT", "")
	formData.Set("__VIEWSTATE", viewstate)
	formData.Set("__VIEWSTATEGENERATOR", viewstateGenerator)
	formData.Set("__EVENTVALIDATION", eventValidation)
	formData.Set("ctl00$ContentPlaceHolder1$StartYearDropDownList", strconv.Itoa(startYear))
	formData.Set("ctl00$ContentPlaceHolder1$EndYearDropDownList", strconv.Itoa(endYear))
	formData.Set("ctl00$ContentPlaceHolder1$StartMonthDropDownList", strconv.Itoa(startMonth))
	formData.Set("ctl00$ContentPlaceHolder1$EndMonthDropDownList", strconv.Itoa(endMonth))
	formData.Set("ctl00$ContentPlaceHolder1$Button1", "Display")
	formData.Set(soraCheckboxName, "on") // Use dynamically found SORA checkbox

	// Step 3: POST the form data
	postReq, err := http.NewRequest("POST", baseURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	// Set POST headers
	postReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	postReq.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	postReq.Header.Set("Cache-Control", "max-age=0")
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.Header.Set("Origin", "https://eservices.mas.gov.sg")
	postReq.Header.Set("Referer", baseURL)
	postReq.Header.Set("Upgrade-Insecure-Requests", "1")
	postReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")

	postResp, err := common.DoWithRetry(src.client, postReq, 3, 500*time.Millisecond, true)
	if err != nil {
		return nil, fmt.Errorf("failed to POST form data: %w", err)
	}
	defer postResp.Body.Close()

	// Step 4: Parse the response HTML to extract interest rates
	responseDoc, err := goquery.NewDocumentFromReader(postResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var interestRates []types.InterestRates

	// Find the data table and extract interest rates
	responseDoc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			if j <= 1 { // Skip header row and empty row
				return
			}

			cells := row.Find("td")
			if cells.Length() >= 5 {
				// Extract the date components and rate
				yearStr := strings.TrimSpace(cells.Eq(0).Text())
				monthStr := strings.TrimSpace(cells.Eq(1).Text())
				dayStr := strings.TrimSpace(cells.Eq(2).Text())
				publicationDateStr := strings.TrimSpace(cells.Eq(3).Text())
				rateStr := strings.TrimSpace(cells.Eq(4).Text())

				// Parse rate
				rate, err := strconv.ParseFloat(rateStr, 64)
				if err != nil {
					src.logger.Warnf("Failed to parse rate %s: %v", rateStr, err)
					return
				}

				// Use publication date if available, otherwise construct from year/month/day
				dateStr := publicationDateStr
				if dateStr == "" && yearStr != "" && monthStr != "" && dayStr != "" {
					dateStr = fmt.Sprintf("%s %s %s", dayStr, monthStr, yearStr)
				}

				if dateStr != "" && rate > 0 {
					interestRates = append(interestRates, types.InterestRates{
						Date:     dateStr,
						Rate:     rate,
						Tenor:    "O/N", // Overnight rate
						Country:  "SG",
						RateType: "SORA", // Singapore Overnight Rate Average
					})
				}
			}
		})
	})

	if len(interestRates) == 0 {
		return nil, fmt.Errorf("no interest rates found in response")
	}

	// Log the date range of fetched data for debugging
	if len(interestRates) > 0 {
		src.logger.Infof("Fetched data from %s to %s (%d records)", 
			interestRates[len(interestRates)-1].Date, interestRates[0].Date, len(interestRates))
	}

	// Store the complete series in database
	if src.db != nil {
		// Merge with existing data to avoid duplicates
		allRates := mergeInterestRates(cachedRates, interestRates)
		src.logger.Infof("Storing %d interest rates for %s in database", len(allRates), country)
		src.db.Put(dbKey, allRates)
	}

	// Filter to return only the requested number of recent records
	filteredRates := filterRecentRates(interestRates, points)

	src.logger.Infof("Successfully fetched %d interest rates for %s", len(filteredRates), country)
	return filteredRates, nil
}

// filterRecentRates filters interest rates to return the most recent 'points' records
func filterRecentRates(rates []types.InterestRates, points int) []types.InterestRates {
	if len(rates) <= points {
		return rates
	}

	// Data is in chronological order (oldest first), so take the last N records
	// to get the most recent data
	return rates[len(rates)-points:]
}

// mergeInterestRates merges existing and new interest rates, removing duplicates
func mergeInterestRates(existing, new []types.InterestRates) []types.InterestRates {
	// Create a map to track existing dates
	existingDates := make(map[string]bool)
	for _, rate := range existing {
		existingDates[rate.Date] = true
	}

	// Add new rates that don't exist
	merged := make([]types.InterestRates, len(existing))
	copy(merged, existing)

	for _, rate := range new {
		if !existingDates[rate.Date] {
			merged = append(merged, rate)
		}
	}

	return merged
}
