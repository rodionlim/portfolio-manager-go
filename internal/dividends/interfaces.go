package dividends

// Manager defines the interface for managing dividends
type Manager interface {
	// CalculateDividendsForAllTickers returns a map of all dividends indexed by ticker
	CalculateDividendsForAllTickers() (map[string][]Dividends, error)

	// CalculateDividendsForSingleTicker calculates dividends for a single ticker
	CalculateDividendsForSingleTicker(ticker string) ([]Dividends, error)
}
