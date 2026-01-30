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

// BenchmarkCost defines broker cost for benchmark trades.
// The effective cost is max(pct * notional, absolute).
type BenchmarkCost struct {
	Pct      float64 `json:"pct"`
	Absolute float64 `json:"absolute"`
}

// BenchmarkMode defines how benchmark trades are simulated.
type BenchmarkMode string

const (
	BenchmarkModeBuyAtStart   BenchmarkMode = "buy_at_start"
	BenchmarkModeMatchTrades  BenchmarkMode = "match_trades"
	BenchmarkModeMatchMonthCF BenchmarkMode = "match_month_cf"
)

// BenchmarkTickerWeight defines benchmark ticker weights.
type BenchmarkTickerWeight struct {
	Ticker string  `json:"ticker"`
	Weight float64 `json:"weight"`
}

// BenchmarkRequest defines parameters for benchmarking.
type BenchmarkRequest struct {
	BookFilter       string                  `json:"book_filter"`
	BenchmarkCost    BenchmarkCost           `json:"benchmark_cost"`
	Mode             BenchmarkMode           `json:"mode"`
	Notional         float64                 `json:"notional,omitempty"`
	BenchmarkTickers []BenchmarkTickerWeight `json:"benchmark_tickers"`
}

// BenchmarkMetrics represents benchmark results.
type BenchmarkMetrics struct {
	IRR       float64 `json:"irr"`
	PricePaid float64 `json:"pricePaid"`
	MV        float64 `json:"mv"`
	Fees      float64 `json:"fees"`
}

// BenchmarkComparisonResult compares portfolio vs benchmark IRR.
type BenchmarkComparisonResult struct {
	PortfolioMetrics   MetricsResult    `json:"portfolio_metrics"`
	BenchmarkMetrics   BenchmarkMetrics `json:"benchmark_metrics"`
	PortfolioIRR       float64          `json:"portfolio_irr"`
	BenchmarkIRR       float64          `json:"benchmark_irr"`
	IRRDifference      float64          `json:"irr_difference"`
	Winner             string           `json:"winner"` // portfolio | benchmark | tie
	BenchmarkCashFlows []CashFlow       `json:"benchmark_cash_flows"`
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
	// BenchmarkPortfolioPerformance compares portfolio IRR against a benchmark
	BenchmarkPortfolioPerformance(req BenchmarkRequest) (BenchmarkComparisonResult, error)
}
