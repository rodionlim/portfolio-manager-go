package types

type StockData struct {
	Ticker    string
	Price     float64
	Currency  string
	Timestamp int64
}

type Dividend struct {
	Date   string
	Amount float64
}

// DataSource defines the interface for different data source engines
type DataSource interface {
	GetStockPrice(symbol string) (*StockData, error)
	GetDividends(symbol string) ([]Dividend, error)
	GetHistoricalData(symbol string, fromDate, toDate int64) ([]*StockData, error)
}
