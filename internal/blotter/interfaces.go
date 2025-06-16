package blotter

import "encoding/csv"

type TradeAdder interface {
	AddTrade(trade Trade) error
}

type TradeRemover interface {
	RemoveTrade(tradeID string) error
}

type TradeGetter interface {
	GetTrades() []Trade
	GetTradeByID(tradeID string) (*Trade, error)
	GetTradesByTicker(ticker string) ([]Trade, error)
	GetAllTickers() ([]string, error)
}

type TradeExporter interface {
	ExportToCSVBytes() ([]byte, error)
}

type TradeImporter interface {
	ImportFromCSVFile(filepath string) error
	ImportFromCSVReader(reader *csv.Reader) error
}
