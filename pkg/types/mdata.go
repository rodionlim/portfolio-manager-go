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

// DataSource defines the interface for different data source engines
type DataSource interface {
	GetAssetPrice(ticker string) (*AssetData, error)
	GetDividendsMetadata(ticker string, witholdingTax float64) ([]DividendsMetadata, error)
	StoreDividendsMetadata(ticker string, dividends []DividendsMetadata, isCustom bool) ([]DividendsMetadata, error)
	GetHistoricalData(ticker string, fromDate, toDate int64) ([]*AssetData, error)
}
