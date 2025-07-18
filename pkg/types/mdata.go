package types

type AssetData struct {
	Ticker    string
	Price     float64
	Currency  string
	Timestamp int64
}

type DividendsMetadata struct {
	Ticker         string
	ExDate         string
	Amount         float64
	Interest       float64 // SSB, TBills and Bonds only, in percentage
	AvgInterest    float64 // SSB, TBills and Bonds only, in percentage
	WithholdingTax float64 // in decimal, not percentage
}

type InterestRates struct {
	Date     string  `json:"date"`
	Rate     float64 `json:"rate"`
	Tenor    string  `json:"tenor"`
	Country  string  `json:"country"`
	RateType string  `json:"rate_type"`
}

// DataSource defines the interface for different data source engines
type DataSource interface {
	GetAssetPrice(ticker string) (*AssetData, error)
	GetDividendsMetadata(ticker string, witholdingTax float64) ([]DividendsMetadata, error)
	StoreDividendsMetadata(ticker string, dividends []DividendsMetadata, isCustom bool) ([]DividendsMetadata, error)
	GetHistoricalData(ticker string, fromDate, toDate int64) ([]*AssetData, error)
	FetchBenchmarkInterestRates(country string, points int) ([]InterestRates, error)
}
