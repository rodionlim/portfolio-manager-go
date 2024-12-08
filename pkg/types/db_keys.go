package types

// Define unexported types for database keys
type dbKey string

// Define database keys
const (
	HeadSequenceKey dbKey = "HEAD_SEQUENCE_NUM"

	TradeKeyPrefix    dbKey = "TRADE:"
	PositionKeyPrefix dbKey = "POSITION:"
)
