package sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"time"
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
				IssueDate   string  `json:"issue_date"`
				CutoffPrice float64 `json:"cutoff_price"`
				CutoffYield float64 `json:"cutoff_yield"`
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
	}

	return dividends, nil
}
