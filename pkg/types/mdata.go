package types

type StockData struct {
	Ticker    string
	Price     float64
	Currency  string
	Timestamp int64
}

type DividendsMetadata struct {
	Ticker         string
	ExDate         string
	Amount         float64
	Interest       float64 // SSB and Bonds only
	AvgInterest    float64 // SSB and Bonds only
	WithholdingTax float64 // in decimal, not percentage
}

// DataSource defines the interface for different data source engines
type DataSource interface {
	GetStockPrice(symbol string) (*StockData, error)
	GetDividendsMetadata(symbol string) ([]DividendsMetadata, error)
	GetHistoricalData(symbol string, fromDate, toDate int64) ([]*StockData, error)
}
