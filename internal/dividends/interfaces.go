package dividends

// Manager defines the interface for managing dividends
type Manager interface {
	// CalculateDividendsForAllTickers returns a map of all dividends indexed by ticker across entire portfolio
	CalculateDividendsForAllTickers() (map[string][]Dividends, error)

	// CalculateDividendsForSingleTicker calculates dividends for a single ticker
	CalculateDividendsForSingleTicker(ticker string) ([]Dividends, error)

	// CalculateDividendsForSingleBook calculates dividends for all tickers within a specific book
	CalculateDividendsForSingleBook(book string) (map[string][]Dividends, error)
}
