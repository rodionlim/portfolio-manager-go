package types

// Define unexported types for database keys
type dbKey string

// Define database keys
const (
	HeadSequenceBlotterKey   dbKey = "BLOTTER_HEAD_SEQUENCE_NUM"
	HeadSequencePortfolioKey dbKey = "PORTFOLIO_HEAD_SEQUENCE_NUM"

	TradeKeyPrefix             dbKey = "TRADE"
	PositionKeyPrefix          dbKey = "POSITION"
	ReferenceDataKeyPrefix     dbKey = "REFDATA"
	DividendsKeyPrefix         dbKey = "DIVIDENDS"
	DividendsCustomKeyPrefix   dbKey = "DIVIDENDS_CUSTOM"
	HistoricalMetricsKeyPrefix dbKey = "METRICS" // Key prefix for historical metrics
	AnalyticsSummaryKeyPrefix  dbKey = "ANALYTICS_SUMMARY"
	MigrationKeyPrefix         dbKey = "MIGRATION"     // Key prefix for migration tracking
	ScheduledJobKeyPrefix      dbKey = "SCHEDULED_JOB" // Key prefix for registered scheduled jobs
)
