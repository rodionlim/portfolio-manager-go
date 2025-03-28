package portfolio

import (
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/pkg/common"
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
	Px            float64
	TotalPaid     float64
	FxRate        float64
}

type Portfolio struct {
	positions       map[string]map[string]*Position // map[trader]map[ticker]*Position
	currentSeqNum   int                             // used as a pointer to point to the last blotter trade that was processed
	db              dal.Database
	blotter         *blotter.TradeBlotter
	mdata           mdata.MarketDataManager
	rdata           rdata.ReferenceManager
	dividendsMgr    *dividends.DividendsManager
	mu              sync.Mutex
	logger          *logging.Logger
	fxCache         *cache.Cache
	fxCacheDuration time.Duration
}

func NewPortfolio(db dal.Database, mdata mdata.MarketDataManager, rdata rdata.ReferenceManager, dividendsSvc *dividends.DividendsManager) *Portfolio {
	var currentSeqNum int
	err := db.Get(string(types.HeadSequencePortfolioKey), currentSeqNum)
	if err != nil {
		currentSeqNum = -1
	}

	defaultFxExpiry := 1 * time.Hour

	return &Portfolio{
		positions:       make(map[string]map[string]*Position),
		currentSeqNum:   currentSeqNum,
		mdata:           mdata,
		rdata:           rdata,
		dividendsMgr:    dividendsSvc,
		db:              db,
		logger:          logging.GetLogger(),
		fxCache:         cache.New(defaultFxExpiry, 2*defaultFxExpiry),
		fxCacheDuration: defaultFxExpiry,
	}
}

func NewPortfolioWithConfigurableExpiry(db dal.Database, mdata mdata.MarketDataManager, rdata rdata.ReferenceManager, dividendsSvc *dividends.DividendsManager, fxDuration time.Duration) *Portfolio {
	var currentSeqNum int
	err := db.Get(string(types.HeadSequencePortfolioKey), currentSeqNum)
	if err != nil {
		currentSeqNum = -1
	}

	return &Portfolio{
		positions:       make(map[string]map[string]*Position),
		currentSeqNum:   currentSeqNum,
		mdata:           mdata,
		rdata:           rdata,
		dividendsMgr:    dividendsSvc,
		db:              db,
		logger:          logging.GetLogger(),
		fxCache:         cache.New(fxDuration, 2*fxDuration),
		fxCacheDuration: fxDuration,
	}
}

// Add a helper function getFXRate that uses go-cache.
// It returns the FX rate relative to SGD. If the position currency is SGD, it returns 1.
func (p *Portfolio) getFXRate(ccy string) (float64, error) {
	// If currency is SGD, FX rate is 1.
	if ccy == "SGD" || ccy == "" {
		return 1, nil
	}

	// Define the pair as X-SGD (e.g. USD-SGD)
	pair := ccy + "-" + "SGD"

	// Check the cache.
	if fx, found := p.fxCache.Get(ccy); found {
		if rate, ok := fx.(float64); ok {
			return rate, nil
		}
	}

	// Not in cache, so retrieve via mdata (assumed to use Yahoo Finance).
	asset, err := p.mdata.GetAssetPrice(pair)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve fx rate for pair %s: %w", pair, err)
	}

	// Store in cache.
	rate := 1 / asset.Price
	p.fxCache.Set(pair, rate, p.fxCacheDuration)
	return rate, nil
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

	p.logger.Infof("Loaded %d positions from database", len(positionKeys))

	return nil
}

// DeletePositions deletes all positions from the database.
func (p *Portfolio) DeletePositions() error {
	positionKeys, err := p.db.GetAllKeysWithPrefix(string(types.PositionKeyPrefix))
	if err != nil {
		return err
	}

	for _, key := range positionKeys {
		err := p.db.Delete(key)
		if err != nil {
			return err
		}
	}

	p.positions = make(map[string]map[string]*Position) // reset positions in memory
	p.currentSeqNum = -1                                // reset current sequence number

	p.logger.Infof("Deleted %d positions from database", len(positionKeys))
	return nil
}

// GetMdataManager returns the market data manager.
func (p *Portfolio) GetMdataManager() mdata.MarketDataManager {
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

// SubscribeToBlotter subscribes to the blotter service, adds a reference to the svc and listens for new trade events.
func (p *Portfolio) SubscribeToBlotter(blotterSvc *blotter.TradeBlotter) {
	p.blotter = blotterSvc

	// Check if the currentSeqNum is less than the current sequence number of the blotter, i
	// if it is, replay the trades from the blotter starting from the currentSeqNum
	blotterSeqNum := blotterSvc.GetCurrentSeqNum()
	if p.currentSeqNum < blotterSeqNum {
		blotterSvc.GetTradesBySeqNumRangeWithCallback(p.currentSeqNum+1, blotterSeqNum, func(trade blotter.Trade) { p.updatePosition(&trade) })
	}

	blotterSvc.Subscribe(blotter.NewTradeEvent, event.NewEventHandler(func(e event.Event) {
		trade := e.Data.(blotter.TradeEventPayload).Trade
		p.logger.Infof("Received 'NEW' trade event. tradeID: %s ticker: %s, tradeDate: %s", trade.TradeID, trade.Ticker, trade.TradeDate)
		p.updatePosition(&trade)
	}))

	blotterSvc.Subscribe(blotter.RemoveTradeEvent, event.NewEventHandler(func(e event.Event) {
		trade := e.Data.(blotter.TradeEventPayload).Trade
		p.logger.Infof("Received 'REMOVE' trade event. tradeID: %s ticker: %s, tradeDate: %s", trade.TradeID, trade.Ticker, trade.TradeDate)

		reverseTradeSide(&trade)

		p.updatePosition(&trade)
	}))

	blotterSvc.Subscribe(blotter.UpdateTradeEvent, event.NewEventHandler(func(e event.Event) {
		trade := e.Data.(blotter.TradeEventPayload).Trade
		originalTrade := e.Data.(blotter.TradeEventPayload).OriginalTrade
		p.logger.Infof("Received 'UPDATE' trade event. tradeID: %s ticker: %s, tradeDate: %s", trade.TradeID, trade.Ticker, trade.TradeDate)

		reverseTradeSide(&originalTrade)

		p.updatePosition(&originalTrade)
		p.updatePosition(&trade)
	}))

	p.logger.Info("Subscribed to blotter service")
}

func (p *Portfolio) AutoCloseTrades() ([]string, error) {
	// If blotter svc reference is nil, return an error since we need the blotter service to get the trades
	if p.blotter == nil {
		return []string{}, fmt.Errorf("blotter service reference is nil")
	}

	// Get all the trades from the blotter
	trades := p.blotter.GetTrades()

	var closedTrades []string
	for _, trade := range trades {
		if trade.Status == blotter.StatusOpen && trade.Side == blotter.TradeSideBuy {
			// currently only support auto closing trades for T-Bills
			if common.IsSgTBill(trade.Ticker) {
				// Check maturity date and compare it against today, if it is less than today, close the trade
				tickerRef, err := p.rdata.GetTicker(trade.Ticker)
				if err != nil {
					return closedTrades, fmt.Errorf("ticker reference not found for ticker %s", trade.Ticker)
				}

				if tickerRef.MaturityDate == "" {
					// needs to be updated by market data manager before attempting to close the trade
					continue
				}

				if !common.IsFutureDate(tickerRef.MaturityDate) {

					p.logger.Infof("Auto closing trade %s for ticker %s", trade.TradeID, trade.Ticker)

					origTrade := trade.Clone()

					// close the original trade
					trade.Status = blotter.StatusClosed
					err := p.blotter.UpdateTrade(trade)
					if err != nil {
						return closedTrades, err
					}

					// make a reversal trade
					trade.Side = blotter.TradeSideSell
					trade.OrigTradeID = trade.TradeID
					trade.TradeDate = tickerRef.MaturityDate
					trade.TradeID = common.GenerateTradeID()
					err = p.blotter.AddTrade(trade)
					if err != nil {
						// if the reversal trade fails, we should revert the amendments on the original trade
						p.blotter.UpdateTrade(origTrade)
						return closedTrades, err
					}

					closedTrades = append(closedTrades, trade.OrigTradeID)
				}
			}
		}
	}
	return closedTrades, nil
}

// reverseTradeSide reverses the trade side in preparation for a trade revert
// (there is an implicit assumption that all removal events have been validated before that the original trade exists, implemented in blotter svc)
func reverseTradeSide(trade *blotter.Trade) {
	if trade.Side == blotter.TradeSideSell {
		trade.Side = blotter.TradeSideBuy
	} else {
		trade.Side = blotter.TradeSideSell
	}
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
	positionToUpdate.TotalPaid = position.TotalPaid
	positionToUpdate.FxRate = position.FxRate

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

	totalPaid := position.TotalPaid + trade.Price*qty // qty is negative for sell trades
	position.TotalPaid = totalPaid
	position.Qty += qty
	position.FxRate, _ = p.getFXRate(position.Ccy)

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
	if positions == nil {
		positions = make([]*Position, 0)
	}
	return positions, err
}

func (p *Portfolio) enrichPositions(positions []*Position) error {
	var errs []error
	for _, position := range positions {
		// TODO: use goroutines to parallelize the enrichment process
		if err := p.enrichPosition(position); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("multiple errors: %v", errs)
	}
	return nil
}

// enrichPosition enriches the position with reference data and market data.
func (p *Portfolio) enrichPosition(position *Position) error {
	tickerRef, err := p.rdata.GetTicker(position.Ticker)
	if err != nil {
		return err
	}

	position.Ccy = tickerRef.Ccy
	position.AssetClass = tickerRef.AssetClass
	position.AssetSubClass = tickerRef.AssetSubClass
	position.FxRate, _ = p.getFXRate(position.Ccy)

	switch tickerRef.AssetClass {
	case rdata.AssetClassEquities, rdata.AssetClassBonds:
		// get dividends
		dividends, err := p.dividendsMgr.CalculateDividendsForSingleTicker(position.Ticker)
		if err != nil {
			// we don't exit here, some tickers might have changed their names over time
			p.logger.Warnf("Failed to get dividends for ticker %s: %v", position.Ticker, err)
		} else {
			position.Dividends = 0 // reset dividends
			for _, dividend := range dividends {
				position.Dividends += dividend.Amount
			}
		}

		if position.Qty == 0 {
			// when the position is closed, the PnL is the total paid + dividends, Mv should be 0
			position.PnL = (position.TotalPaid * -1) + position.Dividends
			position.Mv = 0
		} else {
			assetData, err := p.mdata.GetAssetPrice(position.Ticker)
			if err != nil {
				return err
			}

			position.Mv = position.Qty * assetData.Price
			position.PnL = (assetData.Price-position.AvgPx)*position.Qty + position.Dividends
			position.Px = assetData.Price
		}
	case "":
		// we allow this since we want somethimes want tests to skip position computation,
		// but leave a warning anyway, in case this happens in production
		p.logger.Warnf("Asset class not found for ticker %s", position.Ticker)
		return nil
	default:
		return fmt.Errorf("asset class %s not supported", tickerRef.AssetClass)
	}

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
