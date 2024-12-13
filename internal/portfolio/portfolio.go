package portfolio

import (
	"fmt"
	"sync"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/pkg/event"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
)

type Position struct {
	Ticker        string
	Trader        string
	Ccy           string
	AssetClass    string
	AssetSubClass string
	Qty           float64
	Mv            float64
	PnL           float64
	Dividends     float64
	AvgPx         float64
}

type Portfolio struct {
	positions     map[string]map[string]*Position // map[trader]map[ticker]*Position
	currentSeqNum int                             // used as a pointer to point to the last blotter trade that was processed
	db            dal.Database
	mdata         *mdata.Manager
	rdata         rdata.ReferenceManager
	dividendsMgr  *dividends.DividendsManager
	mu            sync.Mutex
}

func NewPortfolio(db dal.Database, mdata *mdata.Manager, rdata rdata.ReferenceManager, dividendsSvc *dividends.DividendsManager) *Portfolio {
	var currentSeqNum int
	err := db.Get(string(types.HeadSequencePortfolioKey), currentSeqNum)
	if err != nil {
		currentSeqNum = -1
	}

	return &Portfolio{
		positions:     make(map[string]map[string]*Position),
		currentSeqNum: currentSeqNum,
		mdata:         mdata,
		rdata:         rdata,
		dividendsMgr:  dividendsSvc,
		db:            db,
	}
}

// LoadPositions loads the positions from the database.
func (p *Portfolio) LoadPositions() error {
	positionKeys, err := p.db.GetAllKeysWithPrefix(string(types.PositionKeyPrefix))
	if err != nil {
		return err
	}

	for _, key := range positionKeys {
		var position Position
		err := p.db.Get(key, &position)
		if err != nil {
			return err
		}
		err = p.updatePositionFromDb(&position)
		if err != nil {
			return err
		}
	}

	logging.GetLogger().Info("Loaded positions from database")

	return nil
}

// GetMdataManager returns the market data manager.
func (p *Portfolio) GetMdataManager() *mdata.Manager {
	return p.mdata
}

// GetRdataManager returns the reference data manager.
func (p *Portfolio) GetRdataManager() rdata.ReferenceManager {
	return p.rdata
}

// GetDividendsManager returns the dividends manager.
func (p *Portfolio) GetDividendsManager() *dividends.DividendsManager {
	return p.dividendsMgr
}

// SubscribeToBlotter subscribes to the blotter service and listens for new trade events.
func (p *Portfolio) SubscribeToBlotter(blotterSvc *blotter.TradeBlotter) {
	// Check if the currentSeqNum is less than the current sequence number of the blotter, i
	// if it is, replay the trades from the blotter starting from the currentSeqNum
	logger := logging.GetLogger()

	blotterSeqNum := blotterSvc.GetCurrentSeqNum()
	if p.currentSeqNum < blotterSeqNum {
		blotterSvc.GetTradesBySeqNumRangeWithCallback(p.currentSeqNum+1, blotterSeqNum, func(trade blotter.Trade) { p.updatePosition(&trade) })
	}

	blotterSvc.Subscribe(blotter.NewTradeEvent, event.NewEventHandler(func(e event.Event) {
		trade := e.Data.(blotter.NewTradeEventPayload).Trade
		logger.Info("Received new trade event. tradeID:", trade.TradeID)
		p.updatePosition(&trade)
	}))

	logger.Info("Subscribed to blotter service")
}

func (p *Portfolio) updatePositionFromDb(position *Position) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	trader := position.Trader
	ticker := position.Ticker

	if _, ok := p.positions[trader]; !ok {
		p.positions[trader] = make(map[string]*Position)
	}

	if _, ok := p.positions[trader][ticker]; !ok {
		p.positions[trader][ticker] = &Position{Ticker: ticker, Trader: trader}
	}

	positionToUpdate := p.positions[trader][ticker]
	positionToUpdate.Qty = position.Qty
	positionToUpdate.Mv = position.Mv
	positionToUpdate.PnL = position.PnL
	positionToUpdate.Dividends = position.Dividends
	positionToUpdate.AvgPx = position.AvgPx

	return nil
}

func (p *Portfolio) updatePosition(trade *blotter.Trade) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	trader := trade.Trader
	ticker := trade.Ticker

	qty := trade.Quantity
	if trade.Side == blotter.TradeSideSell {
		qty = qty * -1
	}

	if _, ok := p.positions[trader]; !ok {
		p.positions[trader] = make(map[string]*Position)
	}

	if _, ok := p.positions[trader][ticker]; !ok {
		p.positions[trader][ticker] = &Position{Ticker: ticker, Trader: trader}
	}

	position := p.positions[trader][ticker]
	totalPaid := position.AvgPx*position.Qty + trade.Price*qty
	position.Qty += qty

	if position.Qty == 0 {
		position.AvgPx = 0
	} else {
		position.AvgPx = totalPaid / position.Qty
	}

	// Write position to the database
	positionKey := generatePositionKey(trade)
	err := p.db.Put(positionKey, position)
	if err != nil {
		return err
	}

	if trade.SeqNum > p.currentSeqNum {
		p.saveSeqNumToDAL(trade.SeqNum)
	}

	return nil
}

func (p *Portfolio) GetPosition(trader, ticker string) (*Position, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if tickers, ok := p.positions[trader]; ok {
		if position, ok := tickers[ticker]; ok {
			err := p.enrichPosition(position)
			return position, err
		}
	}
	return nil, fmt.Errorf("position not found for trader %s and ticker %s", trader, ticker)
}

func (p *Portfolio) GetPositions(trader string) ([]*Position, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var positions []*Position
	if tickers, ok := p.positions[trader]; ok {
		for _, position := range tickers {
			positions = append(positions, position)
		}
	}
	err := p.enrichPositions(positions)
	return positions, err
}

func (p *Portfolio) GetAllPositions() ([]*Position, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var positions []*Position
	for _, traders := range p.positions {
		for _, position := range traders {
			positions = append(positions, position)
		}
	}
	err := p.enrichPositions(positions)
	return positions, err
}

func (p *Portfolio) enrichPositions(positions []*Position) error {
	for _, position := range positions {
		if err := p.enrichPosition(position); err != nil {
			return err
		}
	}
	return nil
}

func (p *Portfolio) enrichPosition(position *Position) error {
	tickerRef, err := p.rdata.GetTicker(position.Ticker)
	if err != nil {
		return err
	}
	position.Ccy = tickerRef.Ccy
	position.AssetClass = tickerRef.AssetClass
	position.AssetSubClass = tickerRef.AssetSubClass
	return nil
}

// saveSeqNumToDAL saves the current sequence number to the DAL database.
func (p *Portfolio) saveSeqNumToDAL(seqNum int) {
	// Implement the logic to save seqNum to the DAL database
	p.db.Put(string(types.HeadSequencePortfolioKey), seqNum)
}

// generatePositionKey generates a unique key for the position.
func generatePositionKey(trade *blotter.Trade) string {
	return fmt.Sprintf("%s:%s:%s", types.PositionKeyPrefix, trade.Trader, trade.Ticker)
}
