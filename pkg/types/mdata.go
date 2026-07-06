package types

const (
	DividendSourceOfficial = "Official"
	DividendSourceCustom   = "Custom"
)

type AssetData struct {
	Ticker    string  `json:"ticker"`
	Price     float64 `json:"price"`
	AdjClose  float64 `json:"adj_close"`
	Currency  string  `json:"currency"`
	Timestamp int64   `json:"timestamp"`
}

type DividendsMetadata struct {
	Ticker         string
	ExDate         string
	Amount         float64
	Interest       float64 // SSB, TBills and Bonds only, in percentage
	AvgInterest    float64 // SSB, TBills and Bonds only, in percentage
	WithholdingTax float64 // in decimal, not percentage
	Source         string
}

type InterestRates struct {
	Date     string  `json:"date"`
	Rate     float64 `json:"rate"`
	Tenor    string  `json:"tenor"`
	Country  string  `json:"country"`
	RateType string  `json:"rate_type"`
}

type USAIndustryOverview struct {
	ID            string   `json:"id"`
	Industry      string   `json:"industry"`
	MarketCap     *float64 `json:"market_cap"`
	Currency      string   `json:"currency"`
	DividendYield *float64 `json:"dividend_yield"`
	Change        *float64 `json:"change"`
	Volume        *float64 `json:"volume"`
	Sector        string   `json:"sector"`
	Stocks        *int     `json:"stocks"`
}

type USAIndustryPerformance struct {
	ID          string   `json:"id"`
	Industry    string   `json:"industry"`
	Change      *float64 `json:"change"`
	OneWeek     *float64 `json:"one_week"`
	OneMonth    *float64 `json:"one_month"`
	ThreeMonths *float64 `json:"three_months"`
	SixMonths   *float64 `json:"six_months"`
	YTD         *float64 `json:"ytd"`
	OneYear     *float64 `json:"one_year"`
	FiveYears   *float64 `json:"five_years"`
	TenYears    *float64 `json:"ten_years"`
	AllTime     *float64 `json:"all_time"`
}

type USAIndustryStockOverview struct {
	ID                     string   `json:"id"`
	Ticker                 string   `json:"ticker"`
	Company                string   `json:"company"`
	Exchange               string   `json:"exchange"`
	MarketCap              *float64 `json:"market_cap"`
	Currency               string   `json:"currency"`
	Price                  *float64 `json:"price"`
	Change                 *float64 `json:"change"`
	Volume                 *float64 `json:"volume"`
	RelativeVolume         *float64 `json:"relative_volume"`
	PriceToEarnings        *float64 `json:"price_to_earnings"`
	EPSDilutedTTM          *float64 `json:"eps_diluted_ttm"`
	EPSDilutedGrowthYoYTTM *float64 `json:"eps_diluted_growth_yoy_ttm"`
	DividendYieldTTM       *float64 `json:"dividend_yield_ttm"`
	Sector                 string   `json:"sector"`
	AnalystRating          string   `json:"analyst_rating"`
}

type USAIndustryStockPerformance struct {
	ID                 string   `json:"id"`
	Ticker             string   `json:"ticker"`
	Company            string   `json:"company"`
	Exchange           string   `json:"exchange"`
	Currency           string   `json:"currency"`
	Price              *float64 `json:"price"`
	Change             *float64 `json:"change"`
	OneWeek            *float64 `json:"one_week"`
	OneMonth           *float64 `json:"one_month"`
	ThreeMonths        *float64 `json:"three_months"`
	SixMonths          *float64 `json:"six_months"`
	YTD                *float64 `json:"ytd"`
	OneYear            *float64 `json:"one_year"`
	FiveYears          *float64 `json:"five_years"`
	TenYears           *float64 `json:"ten_years"`
	AllTime            *float64 `json:"all_time"`
	VolatilityOneWeek  *float64 `json:"volatility_one_week"`
	VolatilityOneMonth *float64 `json:"volatility_one_month"`
}

type USAStockUnusualVolumeOverview struct {
	ID                     string   `json:"id"`
	Ticker                 string   `json:"ticker"`
	Company                string   `json:"company"`
	Exchange               string   `json:"exchange"`
	RelativeVolume         *float64 `json:"relative_volume"`
	Currency               string   `json:"currency"`
	Price                  *float64 `json:"price"`
	Change                 *float64 `json:"change"`
	Volume                 *float64 `json:"volume"`
	MarketCap              *float64 `json:"market_cap"`
	MarketCapCurrency      string   `json:"market_cap_currency"`
	PriceToEarnings        *float64 `json:"price_to_earnings"`
	EPSDilutedTTM          *float64 `json:"eps_diluted_ttm"`
	EPSDilutedCurrency     string   `json:"eps_diluted_currency"`
	EPSDilutedGrowthYoYTTM *float64 `json:"eps_diluted_growth_yoy_ttm"`
	DividendYieldTTM       *float64 `json:"dividend_yield_ttm"`
	Sector                 string   `json:"sector"`
	AnalystRating          string   `json:"analyst_rating"`
}

type USAStockPreMarketMostActiveOverview struct {
	ID                   string   `json:"id"`
	Ticker               string   `json:"ticker"`
	Company              string   `json:"company"`
	Exchange             string   `json:"exchange"`
	PreMarketVolume      *float64 `json:"pre_market_volume"`
	PreMarketClose       *float64 `json:"pre_market_close"`
	PreMarketCurrency    string   `json:"pre_market_currency"`
	PreMarketChangeAbs   *float64 `json:"pre_market_change_abs"`
	PreMarketChange      *float64 `json:"pre_market_change"`
	PreMarketGap         *float64 `json:"pre_market_gap"`
	Currency             string   `json:"currency"`
	Price                *float64 `json:"price"`
	Change               *float64 `json:"change"`
	Volume               *float64 `json:"volume"`
	MarketCap            *float64 `json:"market_cap"`
	MarketCapCurrency    string   `json:"market_cap_currency"`
	MarketCapPerformance *float64 `json:"market_cap_performance"`
}

type ETFFundFlowOverview struct {
	ID                    string   `json:"id"`
	Ticker                string   `json:"ticker"`
	Fund                  string   `json:"fund"`
	Exchange              string   `json:"exchange"`
	FundFlowsOneYear      *float64 `json:"fund_flows_one_year"`
	FundFlowsCurrency     string   `json:"fund_flows_currency"`
	Price                 *float64 `json:"price"`
	Currency              string   `json:"currency"`
	Change                *float64 `json:"change"`
	Volume                *float64 `json:"volume"`
	VolumeCurrency        string   `json:"volume_currency"`
	RelativeVolume        *float64 `json:"relative_volume"`
	AssetsUnderManagement *float64 `json:"assets_under_management"`
	AUMCurrency           string   `json:"aum_currency"`
	NAVTotalReturn3Y      *float64 `json:"nav_total_return_3y"`
	ExpenseRatio          *float64 `json:"expense_ratio"`
	AssetClass            string   `json:"asset_class"`
	Focus                 string   `json:"focus"`
}

type ETFFundFlowPerformance struct {
	ID          string   `json:"id"`
	Ticker      string   `json:"ticker"`
	Fund        string   `json:"fund"`
	Exchange    string   `json:"exchange"`
	Currency    string   `json:"currency"`
	Price       *float64 `json:"price"`
	Change      *float64 `json:"change"`
	OneWeek     *float64 `json:"one_week"`
	OneMonth    *float64 `json:"one_month"`
	ThreeMonths *float64 `json:"three_months"`
	SixMonths   *float64 `json:"six_months"`
	YTD         *float64 `json:"ytd"`
	OneYear     *float64 `json:"one_year"`
	FiveYears   *float64 `json:"five_years"`
	TenYears    *float64 `json:"ten_years"`
	AllTime     *float64 `json:"all_time"`
}

type ETFFundFlows struct {
	ID          string   `json:"id"`
	Ticker      string   `json:"ticker"`
	Fund        string   `json:"fund"`
	Exchange    string   `json:"exchange"`
	Currency    string   `json:"currency"`
	OneMonth    *float64 `json:"one_month"`
	ThreeMonths *float64 `json:"three_months"`
	OneYear     *float64 `json:"one_year"`
	ThreeYears  *float64 `json:"three_years"`
	YTD         *float64 `json:"ytd"`
}

type ETFSectorOverview struct {
	ID                    string   `json:"id"`
	Ticker                string   `json:"ticker"`
	Fund                  string   `json:"fund"`
	Exchange              string   `json:"exchange"`
	AssetsUnderManagement *float64 `json:"assets_under_management"`
	AUMCurrency           string   `json:"aum_currency"`
	Price                 *float64 `json:"price"`
	Currency              string   `json:"currency"`
	Change                *float64 `json:"change"`
	Volume                *float64 `json:"volume"`
	VolumeCurrency        string   `json:"volume_currency"`
	RelativeVolume        *float64 `json:"relative_volume"`
	NAVTotalReturn3Y      *float64 `json:"nav_total_return_3y"`
	ExpenseRatio          *float64 `json:"expense_ratio"`
	AssetClass            string   `json:"asset_class"`
	Focus                 string   `json:"focus"`
}

type PercentageMetadata struct {
	Unit   string   `json:"unit"`
	Fields []string `json:"fields"`
	Note   string   `json:"note"`
}

type MonetaryMetadata struct {
	CurrencyFields map[string]string `json:"currency_fields"`
	Fields         []string          `json:"fields"`
	Note           string            `json:"note"`
}

type USAIndustryOverviewResponse struct {
	PercentageValues PercentageMetadata    `json:"percentage_values"`
	Industries       []USAIndustryOverview `json:"industries"`
}

type USAIndustryPerformanceResponse struct {
	PercentageValues PercentageMetadata       `json:"percentage_values"`
	Industries       []USAIndustryPerformance `json:"industries"`
}

type USAIndustryStocksOverviewResponse struct {
	Industry         string                     `json:"industry"`
	PercentageValues PercentageMetadata         `json:"percentage_values"`
	Stocks           []USAIndustryStockOverview `json:"stocks"`
}

type USAIndustryStocksPerformanceResponse struct {
	Industry         string                        `json:"industry"`
	PercentageValues PercentageMetadata            `json:"percentage_values"`
	Stocks           []USAIndustryStockPerformance `json:"stocks"`
}

type USAStockUnusualVolumeOverviewResponse struct {
	Screen           string                          `json:"screen"`
	PercentageValues PercentageMetadata              `json:"percentage_values"`
	MonetaryValues   MonetaryMetadata                `json:"monetary_values"`
	Stocks           []USAStockUnusualVolumeOverview `json:"stocks"`
}

type USAStockPreMarketMostActiveOverviewResponse struct {
	Screen           string                                `json:"screen"`
	PercentageValues PercentageMetadata                    `json:"percentage_values"`
	MonetaryValues   MonetaryMetadata                      `json:"monetary_values"`
	Stocks           []USAStockPreMarketMostActiveOverview `json:"stocks"`
}

type ETFFundFlowOverviewResponse struct {
	Screen           string                `json:"screen"`
	PercentageValues PercentageMetadata    `json:"percentage_values"`
	MonetaryValues   MonetaryMetadata      `json:"monetary_values"`
	ETFs             []ETFFundFlowOverview `json:"etfs"`
}

type ETFFundFlowPerformanceResponse struct {
	Screen           string                   `json:"screen"`
	PercentageValues PercentageMetadata       `json:"percentage_values"`
	ETFs             []ETFFundFlowPerformance `json:"etfs"`
}

type ETFFundFlowsResponse struct {
	Screen         string           `json:"screen"`
	MonetaryValues MonetaryMetadata `json:"monetary_values"`
	ETFs           []ETFFundFlows   `json:"etfs"`
}

type ETFSectorOverviewResponse struct {
	Screen           string              `json:"screen"`
	PercentageValues PercentageMetadata  `json:"percentage_values"`
	MonetaryValues   MonetaryMetadata    `json:"monetary_values"`
	ETFs             []ETFSectorOverview `json:"etfs"`
}

// ScreenerSource defines a source for aggregate market screening data.
type ScreenerSource interface {
	FetchUSAIndustryPerformance() ([]USAIndustryPerformance, error)
	FetchUSAIndustryOverview() ([]USAIndustryOverview, error)
	FetchUSAIndustryStocksOverview(industry string) ([]USAIndustryStockOverview, error)
	FetchUSAIndustryStocksPerformance(industry string) ([]USAIndustryStockPerformance, error)
	FetchUSAStockUnusualVolumeOverview() ([]USAStockUnusualVolumeOverview, error)
	FetchUSAStockPreMarketMostActiveOverview() ([]USAStockPreMarketMostActiveOverview, error)
	FetchETFLargestInflowsOverview() ([]ETFFundFlowOverview, error)
	FetchETFLargestInflowsPerformance() ([]ETFFundFlowPerformance, error)
	FetchETFLargestInflowsFundFlows() ([]ETFFundFlows, error)
	FetchETFLargestOutflowsOverview() ([]ETFFundFlowOverview, error)
	FetchETFLargestOutflowsPerformance() ([]ETFFundFlowPerformance, error)
	FetchETFLargestOutflowsFundFlows() ([]ETFFundFlows, error)
	FetchETFSectorOverview() ([]ETFSectorOverview, error)
	FetchETFSectorPerformance() ([]ETFFundFlowPerformance, error)
	FetchETFSectorFundFlows() ([]ETFFundFlows, error)
}

// DataSource defines the interface for different data source engines
type DataSource interface {
	GetAssetPrice(ticker string) (*AssetData, error)
	GetDividendsMetadata(ticker string, witholdingTax float64) ([]DividendsMetadata, error)
	StoreDividendsMetadata(ticker string, dividends []DividendsMetadata, isCustom bool) ([]DividendsMetadata, error)
	DeleteDividendsMetadata(ticker string, isCustom bool) error
	GetHistoricalData(ticker string, fromDate, toDate int64) ([]*AssetData, bool, error)
	FetchBenchmarkInterestRates(country string, points int) ([]InterestRates, error)
}
