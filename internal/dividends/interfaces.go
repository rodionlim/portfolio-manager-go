package dividends

// YearlyDividend represents dividend statistics for a specific year
type YearlyDividend struct {
	Year   int     `json:"year"`
	Amount float64 `json:"amount"`
}

// DividendMetrics represents overall dividend performance metrics
type DividendMetrics struct {
	CAGR                  float64          `json:"cagr"`
	AverageGrowth         float64          `json:"averageGrowth"`
	YearlyDividends       []YearlyDividend `json:"yearlyDividends"`
	TotalDividends        float64          `json:"totalDividends"`
	AverageYearlyDividend float64          `json:"averageYearlyDividend"`
}

// Manager defines the interface for managing dividends
type Manager interface {
	// CalculateDividendsForAllTickers returns a map of all dividends indexed by ticker across entire portfolio
	CalculateDividendsForAllTickers() (map[string][]Dividends, error)

	// CalculateDividendsForSingleTicker calculates dividends for a single ticker
	CalculateDividendsForSingleTicker(ticker string) ([]Dividends, error)

	// CalculateDividendsForSingleBook calculates dividends for all tickers within a specific book
	CalculateDividendsForSingleBook(book string) (map[string][]Dividends, error)
}
