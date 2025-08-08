package portfolio

type Position struct {
	Ticker        string
	Book          string
	Ccy           string
	AssetClass    string
	AssetSubClass string
	Qty           float64
	Mv            float64
	PnL           float64
	Dividends     float64
	AvgPx         float64
	Px            float64
	TotalPaid     float64
	FxRate        float64
}

// PortfolioReader defines the interface for getting portfolio data
type PortfolioReader interface {
	// GetAllPositions returns all positions in the portfolio
	GetAllPositions() ([]*Position, error)
}

type PortfolioWriter interface {
	// DeletePosition deletes a specific position by book and ticker
	DeletePosition(book, ticker string) error
}

type PortfolioReaderWriter interface {
	PortfolioReader
	PortfolioWriter
}
