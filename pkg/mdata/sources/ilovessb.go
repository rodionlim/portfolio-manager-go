package sources

import (
	"fmt"
	"net/http"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ILoveSsb struct {
	db     dal.Database
	url    string
	logger *logging.Logger
}

type SsbData struct {
	CouponDates          []string
	InterestRates        []float64
	AverageReturnPerYear []float64
}

func NewILoveSsb(db dal.Database) *ILoveSsb {
	return &ILoveSsb{
		db:     db,
		url:    "https://www.ilovessb.com/historical-rates",
		logger: logging.GetLogger(),
	}
}

// GetHistoricalData implements types.DataSource. SSB is always traded at par value.
func (src *ILoveSsb) GetHistoricalData(ticker string, fromDate int64, toDate int64) ([]*types.AssetData, error) {
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

// GetAssetPrice implements types.DataSource. SSB is always traded at par value.
func (src *ILoveSsb) GetAssetPrice(ticker string) (*types.AssetData, error) {
	return &types.AssetData{
		Ticker:    ticker,
		Price:     100.0,
		Currency:  "SGD",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (src *ILoveSsb) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	if !common.IsSSB(ticker) {
		return nil, fmt.Errorf("invalid SSB ticker: %s", ticker)
	}

	// fetch from db, if exist, then don't need to hit the actual data source
	if src.db != nil {
		var dividends []types.DividendsMetadata
		src.db.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), &dividends)
		if len(dividends) > 0 {
			return dividends, nil
		}
	}

	resp, err := http.Get(src.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ssb interest rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch coupon payments: status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Use map to aggregate data by issuance name
	ssbDataMap := make(map[string]*SsbData)
	totalCoupons := 20

	doc.Find("div.container").Each(func(i int, s *goquery.Selection) {
		// Extract issuance name, e.g. SBJAN25
		issuanceName := s.Find("h2").Text()
		if issuanceName == "" || len(issuanceName) != 7 {
			return
		}

		// Find the next sibling table
		table := s.NextAllFiltered("div.container.mb-2").Find("table.table-bordered")
		if table.Length() == 0 {
			return
		}

		// Initialize SsbData struct
		ssbData := &SsbData{
			CouponDates:          make([]string, totalCoupons),
			InterestRates:        make([]float64, totalCoupons),
			AverageReturnPerYear: make([]float64, totalCoupons),
		}

		// Extract interest rates and average return per year
		table.Find("tbody tr").Each(func(j int, tr *goquery.Selection) {
			tr.Find("td").Each(func(k int, td *goquery.Selection) {
				valueStr := td.Text()
				value, err := strconv.ParseFloat(valueStr, 64)
				if err != nil {
					return
				}

				if j == 0 {
					ssbData.InterestRates[2*k] = value
					ssbData.InterestRates[2*k+1] = value
				} else if j == 1 {
					ssbData.AverageReturnPerYear[2*k] = value
					ssbData.AverageReturnPerYear[2*k+1] = value
				}
			})
		})

		// Calculate coupon dates
		monthMap := map[string]string{
			"JAN": "01",
			"FEB": "02",
			"MAR": "03",
			"APR": "04",
			"MAY": "05",
			"JUN": "06",
			"JUL": "07",
			"AUG": "08",
			"SEP": "09",
			"OCT": "10",
			"NOV": "11",
			"DEC": "12",
		}
		month := monthMap[issuanceName[2:5]]
		year := "20" + issuanceName[5:7]
		issuanceDate, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-01", year, month))
		if err != nil {
			return
		}
		for i := 0; i < 20; i++ {
			couponDate := issuanceDate.AddDate(0, 6*(i+1), 0)
			ssbData.CouponDates[i] = couponDate.Format("2006-01-02")
		}

		ssbDataMap[issuanceName] = ssbData
	})

	// Convert map to slice of DividendsMetadata
	var dividends []types.DividendsMetadata
	var results []types.DividendsMetadata
	for issuanceName, data := range ssbDataMap {
		for i := 0; i < totalCoupons; i++ {
			dividends = append(dividends, types.DividendsMetadata{
				Ticker:      issuanceName,
				ExDate:      data.CouponDates[i],
				Amount:      data.InterestRates[i],        // interest per $100 notional
				Interest:    data.InterestRates[i],        // interest in percentage
				AvgInterest: data.AverageReturnPerYear[i], // average interest in percentage
			})
		}
		if ticker == issuanceName {
			results = dividends
		}
		// For issuance that are not found in leveldb, store it into level db
		if src.db != nil {
			err := src.db.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, issuanceName), &dividends)
			if err != nil {
				src.logger.Infof("New coupons for ticker %s, storing into database", ticker)
				src.db.Put(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, issuanceName), dividends)
			}
		}
		// zero out dividends slice for next iteration
		dividends = dividends[:0]
	}

	return results, nil
}
