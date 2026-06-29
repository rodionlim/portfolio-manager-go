package types

// MarketRotationMethodologyVersion changes whenever scoring or classification semantics change.
const MarketRotationMethodologyVersion = "1.5.0"

// ETFContribution explains which ETF materially affected a sector's aggregate flow.
type ETFContribution struct {
	Ticker         string  `json:"ticker"`
	Fund           string  `json:"fund"`
	Classification string  `json:"classification"`
	OneMonthFlow   float64 `json:"one_month_flow_usd"`
}

// RotationPerformance contains the return shape used by the deterministic scorer.
type RotationPerformance struct {
	OneDay            float64 `json:"one_day_percent"`
	OneWeek           float64 `json:"one_week_percent"`
	OneMonth          float64 `json:"one_month_percent"`
	ThreeMonths       float64 `json:"three_months_percent"`
	SixMonths         float64 `json:"six_months_percent"`
	OneYear           float64 `json:"one_year_percent"`
	PriorTwoMonthPace float64 `json:"prior_two_month_monthly_pace_percent"`
	Acceleration      float64 `json:"one_month_acceleration_percent"`
}

// SectorFundFlowSummary is the compact, precomputed input for the LLM's 1M flow narrative.
type SectorFundFlowSummary struct {
	Sector                              string              `json:"sector"`
	ETFCount                            int                 `json:"etf_count"`
	BroadETFCount                       int                 `json:"broad_etf_count"`
	SubsectorETFCount                   int                 `json:"subsector_etf_count"`
	PositiveFlowETFCount                int                 `json:"positive_flow_etf_count"`
	PositiveFlowBreadthPercent          float64             `json:"positive_flow_breadth_percent"`
	CombinedAUMUSD                      float64             `json:"combined_aum_usd"`
	OneMonthFlowUSD                     float64             `json:"one_month_flow_usd"`
	ThreeMonthFlowUSD                   float64             `json:"three_month_flow_usd"`
	OneMonthFlowPercentAUM              float64             `json:"one_month_flow_percent_aum"`
	ThreeMonthFlowPercentAUM            float64             `json:"three_month_flow_percent_aum"`
	PriorTwoMonthMonthlyFlowUSD         float64             `json:"prior_two_month_monthly_flow_usd"`
	PriorTwoMonthMonthlyPercentAUM      float64             `json:"prior_two_month_monthly_percent_aum"`
	FlowAccelerationPercentAUM          float64             `json:"flow_acceleration_percent_aum"`
	BroadOneMonthFlowUSD                float64             `json:"broad_one_month_flow_usd"`
	SubsectorOneMonthFlowUSD            float64             `json:"subsector_one_month_flow_usd"`
	BroadFlowAccelerationPercentAUM     float64             `json:"broad_flow_acceleration_percent_aum"`
	SubsectorFlowAccelerationPercentAUM float64             `json:"subsector_flow_acceleration_percent_aum"`
	BroadSubsectorRelationship          string              `json:"broad_subsector_relationship"`
	FlowTrend                           string              `json:"flow_trend"`
	HistoryTransition                   string              `json:"history_transition"`
	BroadPerformance                    RotationPerformance `json:"broad_performance"`
	PerformanceFlowAlignment            string              `json:"performance_flow_alignment"`
	LargestContributors                 []ETFContribution   `json:"largest_contributors"`
	LargestDetractors                   []ETFContribution   `json:"largest_detractors"`
}

// SectorRotationSignal describes a broad-sector setup.
type SectorRotationSignal struct {
	Sector      string              `json:"sector"`
	Setup       string              `json:"setup"`
	Score       float64             `json:"score"`
	Transition  string              `json:"transition"`
	FlowTrend   string              `json:"flow_trend"`
	Performance RotationPerformance `json:"performance"`
}

// SubsectorRotationSignal describes an independently confirmed industry/subsector ETF move.
type SubsectorRotationSignal struct {
	Ticker                   string              `json:"ticker"`
	Fund                     string              `json:"fund"`
	Sector                   string              `json:"sector"`
	MappedIndustries         []string            `json:"mapped_industries,omitempty"`
	Setup                    string              `json:"setup"`
	Score                    float64             `json:"score"`
	Transition               string              `json:"transition"`
	OneMonthFlowUSD          float64             `json:"one_month_flow_usd"`
	FlowPercentAUM           float64             `json:"one_month_flow_percent_aum"`
	ThreeMonthFlowPercentAUM float64             `json:"three_month_flow_percent_aum"`
	FlowAcceleration         float64             `json:"flow_acceleration_percent_aum"`
	Performance              RotationPerformance `json:"performance"`
}

// IndustryRotationSignal is an industry selected for stock drill-down.
type IndustryRotationSignal struct {
	ID          string              `json:"id"`
	Industry    string              `json:"industry"`
	StockSector string              `json:"stock_sector"`
	ETFSector   string              `json:"etf_sector"`
	Lane        string              `json:"lane"`
	Setup       string              `json:"setup"`
	Score       float64             `json:"score"`
	Transition  string              `json:"transition"`
	Performance RotationPerformance `json:"performance"`
}

// StockRotationCandidate is a deterministic candidate, not an LLM-generated recommendation.
type StockRotationCandidate struct {
	Ticker            string              `json:"ticker"`
	Company           string              `json:"company"`
	Industry          string              `json:"industry"`
	ETFSector         string              `json:"etf_sector"`
	Lane              string              `json:"lane"`
	Setup             string              `json:"setup"`
	Score             float64             `json:"score"`
	Transition        string              `json:"transition"`
	MarketCapUSD      float64             `json:"market_cap_usd"`
	DailyTurnoverUSD  float64             `json:"daily_turnover_usd"`
	MonthlyVolatility float64             `json:"monthly_volatility_percent"`
	EPSGrowthYoY      *float64            `json:"eps_growth_yoy_percent,omitempty"`
	AnalystRating     string              `json:"analyst_rating,omitempty"`
	Performance       RotationPerformance `json:"performance"`
	Invalidation      string              `json:"invalidation"`
}

// RotationLostConfirmation identifies a signal present in the prior session but absent now.
type RotationLostConfirmation struct {
	Kind          string  `json:"kind"`
	Key           string  `json:"key"`
	PreviousScore float64 `json:"previous_score"`
}

// MarketRotationDataQuality makes exclusions and partial coverage visible to the LLM and user.
type MarketRotationDataQuality struct {
	Source                string         `json:"source"`
	SourceURL             string         `json:"source_url"`
	InputETFCount         int            `json:"input_etf_count"`
	EligibleETFCount      int            `json:"eligible_etf_count"`
	ExcludedETFCount      int            `json:"excluded_etf_count"`
	ExclusionsByReason    map[string]int `json:"exclusions_by_reason"`
	UnmappedIndustryCount int            `json:"unmapped_industry_count"`
	Stale                 bool           `json:"stale"`
	Persisted             bool           `json:"persisted"`
	Warnings              []string       `json:"warnings,omitempty"`
}

// MarketRotationBrief is intentionally compact so the LLM narrates rather than recomputes.
type MarketRotationBrief struct {
	AsOf                     string                     `json:"as_of"`
	USSessionDate            string                     `json:"us_session_date"`
	MethodologyVersion       string                     `json:"methodology_version"`
	SectorFundFlows          []SectorFundFlowSummary    `json:"sector_fund_flows"`
	BroadSectorRotations     []SectorRotationSignal     `json:"broad_sector_rotations"`
	SubsectorRotations       []SubsectorRotationSignal  `json:"subsector_rotations"`
	PerformanceLedIndustries []IndustryRotationSignal   `json:"performance_led_industries"`
	StockCandidates          []StockRotationCandidate   `json:"stock_candidates"`
	LostConfirmations        []RotationLostConfirmation `json:"lost_confirmations,omitempty"`
	RejectedStockCounts      map[string]int             `json:"rejected_stock_counts"`
	DataQuality              MarketRotationDataQuality  `json:"data_quality"`
	BriefInstructions        []string                   `json:"brief_instructions"`
}
