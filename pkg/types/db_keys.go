package types

// Define unexported types for database keys
type dbKey string

// Define database keys
const (
	HeadSequenceBlotterKey   dbKey = "BLOTTER_HEAD_SEQUENCE_NUM"
	HeadSequencePortfolioKey dbKey = "PORTFOLIO_HEAD_SEQUENCE_NUM"

	TradeKeyPrefix         dbKey = "TRADE"
	PositionKeyPrefix      dbKey = "POSITION"
	ReferenceDataKeyPrefix dbKey = "REFDATA"
	DividendsKeyPrefix     dbKey = "DIVIDENDS"
)
