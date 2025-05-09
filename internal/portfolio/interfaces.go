package portfolio

// PortfolioGetter defines the interface for getting portfolio data
type PortfolioGetter interface {
	// GetAllPositions returns all positions in the portfolio
	GetAllPositions() ([]*Position, error)
}
