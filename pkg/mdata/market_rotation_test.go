package mdata

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/require"
)

func TestJoinAndFilterSectorETFsExcludesUnsafeProducts(t *testing.T) {
	overviews := []types.ETFSectorOverview{
		sectorOverview("xlf", "XLF", "State Street Financial Select Sector SPDR ETF", "AMEX", "Financials", 100),
		sectorOverview("kbwb", "KBWB", "Invesco KBW Bank ETF", "NASDAQ", "Financials", 50),
		sectorOverview("tsll", "TSLL", "Direxion Daily TSLA Bull 2X ETF", "NASDAQ", "Consumer discretionary", 20),
		sectorOverview("foreign", "IITU", "iShares technology UCITS ETF", "LSE", "Information technology", 20),
		sectorOverview("theme", "AIQ", "Global X Artificial Intelligence ETF", "NASDAQ", "Theme", 20),
	}
	performances := make([]types.ETFFundFlowPerformance, 0, len(overviews))
	flows := make([]types.ETFFundFlows, 0, len(overviews))
	for _, row := range overviews {
		performances = append(performances, sectorPerformance(row.ID, row.Ticker, 1, 2, 4, 6, 8, 10))
		flows = append(flows, sectorFlows(row.ID, row.Ticker, 5, 8))
	}

	joined, excluded := joinAndFilterSectorETFs(overviews, performances, flows)
	require.Len(t, joined, 3)
	require.Equal(t, "broad", joined[0].class)
	require.Equal(t, "subsector", joined[1].class)
	require.Equal(t, "subsector", joined[2].class)
	require.Equal(t, "Information technology", joined[2].sector)
	require.ElementsMatch(t, []string{"Semiconductors", "Packaged Software"}, mappedIndustriesForETF(joined[2]))
	require.Equal(t, 1, excluded["leveraged_inverse_or_single_stock"])
	require.Equal(t, 1, excluded["foreign_listing"])
	require.Zero(t, excluded["thematic"])
}

func TestThemeETFsMapToEconomicSectorsWhenPossible(t *testing.T) {
	tests := []struct {
		ticker     string
		fund       string
		wantSector string
		wantHints  []string
	}{
		{"SMH", "VanEck Semiconductor ETF", "Information technology", []string{"Semiconductors"}},
		{"DRAM", "Roundhill Memory ETF", "Information technology", []string{"Semiconductors"}},
		{"AIQ", "Global X Artificial Intelligence & Technology ETF", "Information technology", []string{"Semiconductors", "Packaged Software"}},
		{"BAI", "iShares A.I. Innovation and Tech Active ETF", "Information technology", []string{"Semiconductors", "Packaged Software"}},
		{"ITA", "iShares U.S. Aerospace & Defense ETF", "Industrials", []string{"Aerospace & Defense"}},
		{"URA", "Global X Uranium ETF", "Materials", []string{"Precious Metals", "Other Metals/Minerals"}},
		{"ARKG", "ARK Genomic Revolution ETF", "Health care", []string{"Biotechnology", "Medical Specialties"}},
		{"MISC", "Miscellaneous Future Themes ETF", "Thematic", nil},
	}

	for _, test := range tests {
		t.Run(test.ticker, func(t *testing.T) {
			overview := sectorOverview(strings.ToLower(test.ticker), test.ticker, test.fund, "NASDAQ", "Theme", 100)
			require.Equal(t, test.wantSector, canonicalETFSector(overview))
			joined := joinedSectorETF{overview: overview}
			require.ElementsMatch(t, test.wantHints, mappedIndustriesForETF(joined))
		})
	}
}

func TestAggregateSectorFlowsSeparatesBroadAndSubsector(t *testing.T) {
	broad := joinedETFForTest("XLF", "broad", 100, 10, 15)
	subsector := joinedETFForTest("KBWB", "subsector", 50, -2, -1)

	rows := aggregateSectorFlows([]joinedSectorETF{broad, subsector})
	require.Len(t, rows, 1)
	row := rows[0]
	require.Equal(t, "Financials", row.Sector)
	require.InDelta(t, 8, row.OneMonthFlowUSD, 0.001)
	require.InDelta(t, 5.333, row.OneMonthFlowPercentAUM, 0.001)
	require.InDelta(t, 9.333, row.ThreeMonthFlowPercentAUM, 0.001)
	require.InDelta(t, 7.5, row.BroadFlowAccelerationPercentAUM, 0.001)
	require.InDelta(t, -5, row.SubsectorFlowAccelerationPercentAUM, 0.001)
	require.Equal(t, "broad_inflow_subsector_outflow", row.BroadSubsectorRelationship)
	require.Equal(t, "XLF", row.LargestContributors[0].Ticker)
	require.Equal(t, "KBWB", row.LargestDetractors[0].Ticker)
}

func TestAggregateSectorFlowsSortsByOneAndThreeMonthFlowPercentAUM(t *testing.T) {
	shortBurst := joinedETFForTest("XLE", "broad", 100, 12, 12)
	shortBurst.sector = "Energy"
	sustained := joinedETFForTest("XLK", "broad", 100, 8, 30)
	sustained.sector = "Information technology"

	rows := aggregateSectorFlows([]joinedSectorETF{shortBurst, sustained})
	require.Len(t, rows, 2)
	require.Equal(t, "Information technology", rows[0].Sector)
	require.Equal(t, "Energy", rows[1].Sector)
}

func TestIndustryCrosswalkCoversMixedTradingViewSectors(t *testing.T) {
	tests := []struct {
		stockSector string
		industry    string
		want        string
		mapped      bool
	}{
		{"Finance", "Major Banks", "Financials", true},
		{"Finance", "Real Estate Investment Trusts", "Real estate", true},
		{"Electronic technology", "Semiconductors", "Information technology", true},
		{"Electronic technology", "Aerospace & Defense", "Industrials", true},
		{"Industrial services", "Oilfield Services/Equipment", "Energy", true},
		{"Distribution services", "Medical Distributors", "Health care", true},
		{"Consumer non-durables", "Apparel/Footwear", "Consumer discretionary", true},
		{"Technology services", "Internet Software/Services", "Communication services", true},
		{"Miscellaneous", "Miscellaneous", "", false},
	}
	for _, test := range tests {
		t.Run(test.stockSector+"/"+test.industry, func(t *testing.T) {
			got, ok := mapIndustryToETFSector(test.stockSector, test.industry)
			require.Equal(t, test.mapped, ok)
			require.Equal(t, test.want, got)
		})
	}
}

func TestSetupAndOneDaySpikeClassification(t *testing.T) {
	recovery := types.RotationPerformance{OneDay: 1, OneWeek: 2, OneMonth: 4, SixMonths: -1, OneYear: 5}
	require.Equal(t, "recovery", setupType(recovery, 0.2, 0.9, 0.3))
	ignition := types.RotationPerformance{OneDay: 1, OneWeek: 3, OneMonth: 5, SixMonths: 2, OneYear: 4}
	require.Equal(t, "ignition", setupType(ignition, 0.5, 0.8, 0.5))
	require.True(t, oneDaySpike(types.RotationPerformance{OneDay: 4, OneWeek: 5}, 0.95))
	require.False(t, oneDaySpike(types.RotationPerformance{OneDay: 1, OneWeek: 5}, 0.95))
}

func TestHistoryTransitionsAndLostConfirmations(t *testing.T) {
	previous := &marketRotationSnapshot{Brief: types.MarketRotationBrief{
		SectorFundFlows:      []types.SectorFundFlowSummary{{Sector: "Technology", FlowAccelerationPercentAUM: -0.5}},
		BroadSectorRotations: []types.SectorRotationSignal{{Sector: "Technology", Score: 80}},
		StockCandidates:      []types.StockRotationCandidate{{Ticker: "OLD", Score: 75}},
	}}
	rows := []types.SectorFundFlowSummary{{Sector: "Technology", FlowAccelerationPercentAUM: 0.4}}
	applySectorFlowTransitions(rows, previous)
	require.Equal(t, "newly_accelerating", rows[0].HistoryTransition)

	lost := detectLostConfirmations(previous, &types.MarketRotationBrief{})
	require.Len(t, lost, 2)
	require.Equal(t, "broad_sector", lost[0].Kind)
	require.Equal(t, "stock", lost[1].Kind)
}

func TestScreenDailyMarketRotationPersistenceAndStaleDetection(t *testing.T) {
	screener := rotationTestScreener{
		overviews:    []types.ETFSectorOverview{sectorOverview("xlf", "XLF", "State Street Financial Select Sector SPDR ETF", "AMEX", "Financials", 100)},
		performances: []types.ETFFundFlowPerformance{sectorPerformance("xlf", "XLF", 1, 2, 4, 8, 10, 12)},
		flows:        []types.ETFFundFlows{sectorFlows("xlf", "XLF", 10, 12)},
	}
	db := newMemoryRotationDB()
	now := time.Date(2026, 6, 24, 10, 30, 0, 0, time.UTC)

	first, err := screenDailyMarketRotation(&screener, db, MarketRotationOptions{PersistHistory: true, MaxStockCandidates: 5}, now)
	require.NoError(t, err)
	require.True(t, first.DataQuality.Persisted)
	require.False(t, first.DataQuality.Stale)
	require.Len(t, db.values, 1)

	second, err := screenDailyMarketRotation(&screener, db, MarketRotationOptions{PersistHistory: true, MaxStockCandidates: 5}, now.Add(time.Hour))
	require.NoError(t, err)
	require.True(t, second.DataQuality.Stale)
	require.False(t, second.DataQuality.Persisted)
	require.Len(t, db.values, 1)

	dryRun, err := screenDailyMarketRotation(&screener, nil, MarketRotationOptions{PersistHistory: false, MaxStockCandidates: 5}, now)
	require.NoError(t, err)
	require.False(t, dryRun.DataQuality.Persisted)
}

func TestScreenDailyMarketRotationReturnsOnlyQualifiedStocks(t *testing.T) {
	screener := rotationTestScreener{
		overviews:    []types.ETFSectorOverview{sectorOverview("xlf", "XLF", "State Street Financial Select Sector SPDR ETF", "AMEX", "Financials", 100)},
		performances: []types.ETFFundFlowPerformance{sectorPerformance("xlf", "XLF", 0.2, 2, 4, 8, 10, 12)},
		flows:        []types.ETFFundFlows{sectorFlows("xlf", "XLF", 10, 12)},
		industries: []types.USAIndustryOverview{{
			ID: "banks", Industry: "Major Banks", Sector: "Finance",
		}},
		industryPerf:     []types.USAIndustryPerformance{industryPerformance("banks", "Major Banks", 0.2, 2, 5, 4, -3, -2)},
		stockOverviews:   map[string][]types.USAIndustryStockOverview{"Major Banks": makeTestStockOverviews()},
		stockPerformance: map[string][]types.USAIndustryStockPerformance{"Major Banks": makeTestStockPerformance()},
	}

	brief, err := screenDailyMarketRotation(&screener, nil, MarketRotationOptions{PersistHistory: false, MaxStockCandidates: 5}, time.Date(2026, 6, 24, 10, 30, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotEmpty(t, brief.StockCandidates)
	require.LessOrEqual(t, len(brief.StockCandidates), 5)
	for _, stock := range brief.StockCandidates {
		require.GreaterOrEqual(t, stock.MarketCapUSD, 2_000_000_000.0)
		require.GreaterOrEqual(t, stock.DailyTurnoverUSD, 20_000_000.0)
		require.Greater(t, stock.Performance.ThreeMonths, 0.0)
		require.NotEqual(t, "LOWCAP", stock.Ticker)
		require.NotEqual(t, "NEG3M", stock.Ticker)
	}
	require.Greater(t, brief.RejectedStockCounts["market_cap_below_2b"], 0)
	require.Greater(t, brief.RejectedStockCounts["three_month_performance_non_positive"], 0)
}

func TestSelectIndustryDrilldownsUsesTopThreeSortedIndustries(t *testing.T) {
	industries := []industryAnalysis{
		{signal: types.IndustryRotationSignal{Industry: "Semiconductors", Lane: "subsector_rotation", Score: 95}},
		{signal: types.IndustryRotationSignal{Industry: "Packaged Software", Lane: "performance_led", Score: 90}},
		{signal: types.IndustryRotationSignal{Industry: "Aerospace & Defense", Lane: "broad_sector_confirmed", Score: 85}},
		{signal: types.IndustryRotationSignal{Industry: "Major Banks", Lane: "broad_sector_confirmed", Score: 80}},
	}

	selected := selectIndustryDrilldowns(industries)
	require.Len(t, selected, 3)
	require.Equal(t, "Semiconductors", selected[0].signal.Industry)
	require.Equal(t, "Packaged Software", selected[1].signal.Industry)
	require.Equal(t, "Aerospace & Defense", selected[2].signal.Industry)
}

type rotationTestScreener struct {
	MarketDataScreener
	overviews        []types.ETFSectorOverview
	performances     []types.ETFFundFlowPerformance
	flows            []types.ETFFundFlows
	industries       []types.USAIndustryOverview
	industryPerf     []types.USAIndustryPerformance
	stockOverviews   map[string][]types.USAIndustryStockOverview
	stockPerformance map[string][]types.USAIndustryStockPerformance
}

func (s *rotationTestScreener) FetchETFSectorOverview() ([]types.ETFSectorOverview, error) {
	return s.overviews, nil
}

func (s *rotationTestScreener) FetchETFSectorPerformance() ([]types.ETFFundFlowPerformance, error) {
	return s.performances, nil
}

func (s *rotationTestScreener) FetchETFSectorFundFlows() ([]types.ETFFundFlows, error) {
	return s.flows, nil
}

func (s *rotationTestScreener) FetchUSAIndustryOverview() ([]types.USAIndustryOverview, error) {
	return s.industries, nil
}

func (s *rotationTestScreener) FetchUSAIndustryPerformance() ([]types.USAIndustryPerformance, error) {
	return s.industryPerf, nil
}

func (s *rotationTestScreener) FetchUSAIndustryStocksOverview(industry string) ([]types.USAIndustryStockOverview, error) {
	return s.stockOverviews[industry], nil
}

func (s *rotationTestScreener) FetchUSAIndustryStocksPerformance(industry string) ([]types.USAIndustryStockPerformance, error) {
	return s.stockPerformance[industry], nil
}

func (s *rotationTestScreener) FetchUSAStockUnusualVolumeOverview() ([]types.USAStockUnusualVolumeOverview, error) {
	return nil, nil
}

func (s *rotationTestScreener) FetchUSAStockPreMarketMostActiveOverview() ([]types.USAStockPreMarketMostActiveOverview, error) {
	return nil, nil
}

type memoryRotationDB struct {
	values map[string][]byte
}

func newMemoryRotationDB() *memoryRotationDB {
	return &memoryRotationDB{values: make(map[string][]byte)}
}

func (db *memoryRotationDB) Get(key string, value interface{}) error {
	data, ok := db.values[key]
	if !ok {
		return errors.New("not found")
	}
	return json.Unmarshal(data, value)
}

func (db *memoryRotationDB) Put(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	db.values[key] = data
	return nil
}

func (db *memoryRotationDB) Delete(key string) error {
	delete(db.values, key)
	return nil
}

func (db *memoryRotationDB) GetAllKeysWithPrefix(prefix string) ([]string, error) {
	keys := make([]string, 0)
	for key := range db.values {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys, nil
}

func sectorOverview(id, ticker, fund, exchange, focus string, aum float64) types.ETFSectorOverview {
	price := 10.0
	return types.ETFSectorOverview{ID: id, Ticker: ticker, Fund: fund, Exchange: exchange, AssetsUnderManagement: &aum, AUMCurrency: "USD", Price: &price, Currency: "USD", AssetClass: "Equity", Focus: focus}
}

func sectorPerformance(id, ticker string, day, week, month, three, six, year float64) types.ETFFundFlowPerformance {
	return types.ETFFundFlowPerformance{ID: id, Ticker: ticker, Change: &day, OneWeek: &week, OneMonth: &month, ThreeMonths: &three, SixMonths: &six, OneYear: &year}
}

func sectorFlows(id, ticker string, month, three float64) types.ETFFundFlows {
	return types.ETFFundFlows{ID: id, Ticker: ticker, Currency: "USD", OneMonth: &month, ThreeMonths: &three}
}

func joinedETFForTest(ticker, class string, aum, month, three float64) joinedSectorETF {
	overview := sectorOverview(strings.ToLower(ticker), ticker, ticker+" fund", "AMEX", "Financials", aum)
	flows := sectorFlows(strings.ToLower(ticker), ticker, month, three)
	return joinedSectorETF{overview: overview, flows: flows, class: class, sector: "Financials", perf: types.RotationPerformance{OneWeek: 2, OneMonth: 4, ThreeMonths: 6}}
}

func industryPerformance(id, industry string, day, week, month, three, six, year float64) types.USAIndustryPerformance {
	return types.USAIndustryPerformance{ID: id, Industry: industry, Change: &day, OneWeek: &week, OneMonth: &month, ThreeMonths: &three, SixMonths: &six, OneYear: &year}
}

func makeTestStockOverviews() []types.USAIndustryStockOverview {
	result := make([]types.USAIndustryStockOverview, 0, 7)
	for i, ticker := range []string{"BANKA", "NEG3M", "BANKB", "BANKC", "BANKD", "BANKE", "LOWCAP"} {
		marketCap := 10_000_000_000.0 + float64(i)*1_000_000_000
		if ticker == "LOWCAP" {
			marketCap = 500_000_000
		}
		price, volume := 50.0, 1_000_000.0
		result = append(result, types.USAIndustryStockOverview{ID: strings.ToLower(ticker), Ticker: ticker, Company: ticker + " Inc", Currency: "USD", MarketCap: &marketCap, Price: &price, Volume: &volume})
	}
	return result
}

func makeTestStockPerformance() []types.USAIndustryStockPerformance {
	result := make([]types.USAIndustryStockPerformance, 0, 7)
	for i, ticker := range []string{"BANKA", "NEG3M", "BANKB", "BANKC", "BANKD", "BANKE", "LOWCAP"} {
		day := 0.1
		week := 2.0 + float64(i)/10
		month := 5.0 + float64(i)
		three, six, year := 4.0, -3.0, -2.0
		if ticker == "NEG3M" {
			three = -1.0
		}
		volatility := 1.0 + float64(i)/10
		result = append(result, types.USAIndustryStockPerformance{ID: strings.ToLower(ticker), Ticker: ticker, Change: &day, OneWeek: &week, OneMonth: &month, ThreeMonths: &three, SixMonths: &six, OneYear: &year, VolatilityOneMonth: &volatility})
	}
	return result
}
