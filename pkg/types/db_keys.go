package types

// Define unexported types for database keys
type dbKey string

// Define database keys
const (
	HeadSequenceBlotterKey   dbKey = "BLOTTER_HEAD_SEQUENCE_NUM"
	HeadSequencePortfolioKey dbKey = "PORTFOLIO_HEAD_SEQUENCE_NUM"

	TradeKeyPrefix             dbKey = "TRADE"
	ConfirmationKeyPrefix      dbKey = "CONFIRMATION"
	PositionKeyPrefix          dbKey = "POSITION"
	ReferenceDataKeyPrefix     dbKey = "REFDATA"
	DividendsKeyPrefix         dbKey = "DIVIDENDS"
	DividendsCustomKeyPrefix   dbKey = "DIVIDENDS_CUSTOM"
	InterestRatesKeyPrefix     dbKey = "INTEREST_RATES" // Key prefix for interest rates data
	HistoricalMetricsKeyPrefix dbKey = "METRICS"        // Key prefix for historical metrics
	AnalyticsSummaryKeyPrefix  dbKey = "ANALYTICS_SUMMARY"
	MigrationKeyPrefix         dbKey = "MIGRATION" // Key prefix for migration tracking

	ScheduledJobKeyPrefix           dbKey = "SCHEDULED_JOB"            // Key prefix for registered scheduled jobs
	CustomMetricsJobKeyPrefix       dbKey = "METRICS_BY_BOOK"          // Key prefix for custom metrics jobs
	CachedPrevDailyPricesKeyPrefix  dbKey = "CACHE_PREV_DAILY_PRICES"  // Key prefix for cached previous daily prices
	CachedPrev2DailyPricesKeyPrefix dbKey = "CACHE_PREV2_DAILY_PRICES" // Key prefix for cached t-2 daily prices

	HistoricalAssetConfigKey     dbKey = "HISTORICAL_ASSET_CONFIG" // Key for historical asset configuration
	HistoricalAssetDataKeyPrefix dbKey = "HISTORICAL_ASSET_DATA"   // Key prefix for historical asset data
)
