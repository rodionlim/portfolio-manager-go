package metrics

// CashFlowType defines the type of cash flow
type CashFlowType string

const (
	// CashFlowTypeBuy represents a buy trade cash flow
	CashFlowTypeBuy CashFlowType = "buy"
	// CashFlowTypeSell represents a sell trade cash flow
	CashFlowTypeSell CashFlowType = "sell"
	// CashFlowTypeDividend represents a dividend cash flow
	CashFlowTypeDividend CashFlowType = "dividend"
	// CashFlowTypePortfolioValue represents the current portfolio value cash flow
	CashFlowTypePortfolioValue CashFlowType = "final value"
)

// CashFlow represents a single cash flow entry for IRR calculation
type CashFlow struct {
	Date        string       `json:"date"`
	Cash        float64      `json:"cash"`
	Ticker      string       `json:"ticker"`
	Description CashFlowType `json:"description"`
}

// MetricsResult represents the result of an IRR calculation, and other portfolio metrics
type MetricsResult struct {
	IRR            float64 `json:"irr"`
	PricePaid      float64 `json:"pricePaid"`      // Buy - Sell
	MV             float64 `json:"mv"`             // Portfolio market value
	TotalDividends float64 `json:"totalDividends"` // Total dividends
}

// MetricResultsWithCashFlows includes the cashflows used for portfolio metric calculations
type MetricResultsWithCashFlows struct {
	Metrics   MetricsResult `json:"metrics"`
	CashFlows []CashFlow    `json:"cashFlows"`
	Label     string        `json:"label"` // Optional label for the metrics, e.g. book name, empty for entire portfolio
}

// MetricsCalculator defines the interface for portfolio metrics calculations
type MetricsCalculator interface {
	// CalculateIRR computes the Internal Rate of Return (XIRR) for the portfolio
	CalculateIRR() (MetricsResult, error)
}

// MetricsServicer defines the interface for the metrics service
type MetricsServicer interface {
	// CalculatePortfolioMetrics computes the XIRR for the portfolio using all trades, dividends, and current market value as final cash flow
	// It also stores other metrics such as price paid, market value of portfolio and total dividends
	// If book_filter is specified, it filters trades by the given book
	CalculatePortfolioMetrics(book_filter string) (MetricResultsWithCashFlows, error)
}
