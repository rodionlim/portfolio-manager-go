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

// PortfolioGetter defines the interface for getting portfolio data
type PortfolioGetter interface {
	// GetAllPositions returns all positions in the portfolio
	GetAllPositions() ([]*Position, error)
}
