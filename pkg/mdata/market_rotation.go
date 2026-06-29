package mdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"portfolio-manager/pkg/types"
)

const (
	defaultMaxStockCandidates = 5
	maxRotationHistory        = 20
	maxIndustryDrilldowns     = 3
	marketRotationSourceURL   = "https://www.tradingview.com/markets/etfs/funds-sector-etfs/"
)

var broadSectorETFs = map[string]string{
	"XLC": "Communication services", "VOX": "Communication services",
	"XLY": "Consumer discretionary", "VCR": "Consumer discretionary",
	"XLP": "Consumer staples", "VDC": "Consumer staples",
	"XLE": "Energy", "VDE": "Energy",
	"XLF": "Financials", "VFH": "Financials",
	"XLV": "Health care", "VHT": "Health care",
	"XLI": "Industrials", "VIS": "Industrials",
	"XLB": "Materials", "VAW": "Materials",
	"XLRE": "Real estate", "VNQ": "Real estate",
	"XLK": "Information technology", "VGT": "Information technology",
	"XLU": "Utilities", "VPU": "Utilities",
}

// Subsector hints are deliberately explicit. Unmapped subsector ETFs remain visible in the
// subsector lane but do not trigger an unrelated stock-industry drill-down.
var subsectorIndustryHints = map[string][]string{
	"KBWB": {"Major Banks", "Regional Banks"},
	"KRE":  {"Regional Banks"},
	"KBE":  {"Major Banks", "Regional Banks", "Savings Banks"},
	"XBI":  {"Biotechnology"},
	"IBB":  {"Biotechnology"},
	"DRAM": {"Semiconductors"},
	"SMH":  {"Semiconductors"},
	"SOXX": {"Semiconductors"},
	"IGV":  {"Packaged Software"},
	"ITA":  {"Aerospace & Defense"},
	"PPA":  {"Aerospace & Defense"},
	"XAR":  {"Aerospace & Defense"},
	"GDX":  {"Precious Metals"},
	"GDXJ": {"Precious Metals"},
	"COPX": {"Other Metals/Minerals"},
	"XME":  {"Steel", "Aluminum", "Other Metals/Minerals"},
}

type thematicETFMapping struct {
	sector     string
	industries []string
	keywords   []string
	tickers    []string
}

var thematicETFMappings = []thematicETFMapping{
	{
		sector:     "Information technology",
		industries: []string{"Semiconductors"},
		keywords:   []string{"semiconductor", "semiconductors", "memory", "chip ", "chips"},
		tickers:    []string{"SMH", "SOXX", "SOXQ"},
	},
	{
		sector:     "Information technology",
		industries: []string{"Semiconductors", "Packaged Software"},
		keywords:   []string{"artificial intelligence", " a.i. ", " ai ", "robotics", "automation", "quantum"},
		tickers:    []string{"AIQ", "BAI", "BOTZ", "ROBO", "ARKQ", "IRBO", "CHAT"},
	},
	{
		sector:     "Information technology",
		industries: []string{"Packaged Software"},
		keywords:   []string{"software", "cloud", "cybersecurity", "cyber security"},
		tickers:    []string{"IGV", "CIBR", "HACK", "SKYY", "WCLD"},
	},
	{
		sector:     "Industrials",
		industries: []string{"Aerospace & Defense"},
		keywords:   []string{"aerospace", "defense", "defence", "space", "drone", "drones"},
		tickers:    []string{"ITA", "PPA", "XAR", "UFO"},
	},
	{
		sector:     "Materials",
		industries: []string{"Precious Metals", "Other Metals/Minerals"},
		keywords:   []string{"gold", "silver", "copper", "lithium", "rare earth", "uranium", "metals", "mining"},
		tickers:    []string{"GDX", "GDXJ", "COPX", "XME", "LIT", "URA", "URNM", "REMX"},
	},
	{
		sector:     "Energy",
		industries: []string{"Oil & Gas Production", "Oilfield Services/Equipment"},
		keywords:   []string{"energy", "oil", "gas", "solar", "wind", "hydrogen", "clean power", "renewable"},
		tickers:    []string{"TAN", "ICLN", "QCLN", "FAN", "PBW"},
	},
	{
		sector:     "Health care",
		industries: []string{"Biotechnology", "Medical Specialties"},
		keywords:   []string{"biotech", "biotechnology", "genomics", "genomic", "medical", "health care", "healthcare"},
		tickers:    []string{"XBI", "IBB", "ARKG"},
	},
	{
		sector:     "Financials",
		industries: []string{"Investment Banks/Brokers", "Finance/Rental/Leasing"},
		keywords:   []string{"fintech", "blockchain", "crypto", "bitcoin"},
		tickers:    []string{"BLOK", "BITQ", "FINX"},
	},
	{
		sector:     "Communication services",
		industries: []string{"Internet Software/Services", "Movies/Entertainment"},
		keywords:   []string{"internet", "social media", "metaverse", "gaming", "video game", "streaming", "media"},
		tickers:    []string{"SOCL", "HERO", "ESPO", "META"},
	},
	{
		sector:     "Consumer discretionary",
		industries: []string{"Internet Retail", "Motor Vehicles", "Specialty Stores"},
		keywords:   []string{"ecommerce", "e-commerce", "retail", "consumer discretionary", "electric vehicle", "electric vehicles"},
		tickers:    []string{"IBUY", "ONLN", "DRIV", "IDRV"},
	},
}

type joinedSectorETF struct {
	overview    types.ETFSectorOverview
	performance types.ETFFundFlowPerformance
	flows       types.ETFFundFlows
	class       string
	sector      string
	flow1Pct    float64
	flow3Pct    float64
	flowPrior2  float64
	flowAccel   float64
	perf        types.RotationPerformance
}

type sectorAccumulator struct {
	sector     string
	etfs       []joinedSectorETF
	broad      []joinedSectorETF
	subsector  []joinedSectorETF
	aum        float64
	flow1      float64
	flow3      float64
	positive   int
	broadFlow1 float64
	broadFlow3 float64
	broadAUM   float64
	subFlow1   float64
	subFlow3   float64
	subAUM     float64
}

type marketRotationSnapshot struct {
	Fingerprint string                    `json:"fingerprint"`
	Brief       types.MarketRotationBrief `json:"brief"`
}

type industryAnalysis struct {
	signal     types.IndustryRotationSignal
	confidence float64
}

// ScreenDailyMarketRotation builds the deterministic payload consumed by OpenClaw.
func (m *Manager) ScreenDailyMarketRotation(options MarketRotationOptions) (*types.MarketRotationBrief, error) {
	if options.MaxStockCandidates <= 0 || options.MaxStockCandidates > defaultMaxStockCandidates {
		options.MaxStockCandidates = defaultMaxStockCandidates
	}
	return screenDailyMarketRotation(m, m.db, options, time.Now())
}

func screenDailyMarketRotation(screener MarketDataScreener, db databaseStore, options MarketRotationOptions, now time.Time) (*types.MarketRotationBrief, error) {
	overview, err := screener.FetchETFSectorOverview()
	if err != nil {
		return nil, fmt.Errorf("fetch sector ETF overview: %w", err)
	}
	performance, err := screener.FetchETFSectorPerformance()
	if err != nil {
		return nil, fmt.Errorf("fetch sector ETF performance: %w", err)
	}
	flows, err := screener.FetchETFSectorFundFlows()
	if err != nil {
		return nil, fmt.Errorf("fetch sector ETF fund flows: %w", err)
	}
	industryOverview, err := screener.FetchUSAIndustryOverview()
	if err != nil {
		return nil, fmt.Errorf("fetch industry overview: %w", err)
	}
	industryPerformance, err := screener.FetchUSAIndustryPerformance()
	if err != nil {
		return nil, fmt.Errorf("fetch industry performance: %w", err)
	}

	joined, exclusions := joinAndFilterSectorETFs(overview, performance, flows)
	sectorFlows := aggregateSectorFlows(joined)
	previous, historyErr := loadLatestRotationSnapshot(db)
	if historyErr != nil && options.PersistHistory {
		return nil, fmt.Errorf("load market rotation history: %w", historyErr)
	}
	applySectorFlowTransitions(sectorFlows, previous)
	broadSignals := buildBroadSectorSignals(sectorFlows, previous)
	subsectorSignals := buildSubsectorSignals(joined, previous)
	industries, unmapped := buildIndustrySignals(industryOverview, industryPerformance, sectorFlows, broadSignals, subsectorSignals, previous)

	rejected := make(map[string]int)
	stocks, err := buildStockCandidates(screener, industries, previous, options.MaxStockCandidates, rejected)
	if err != nil {
		return nil, err
	}

	newYork, tzErr := time.LoadLocation("America/New_York")
	if tzErr != nil {
		newYork = time.FixedZone("ET", -5*60*60)
	}
	brief := &types.MarketRotationBrief{
		AsOf:                     now.UTC().Format(time.RFC3339),
		USSessionDate:            now.In(newYork).Format("2006-01-02"),
		MethodologyVersion:       types.MarketRotationMethodologyVersion,
		SectorFundFlows:          sectorFlows,
		BroadSectorRotations:     broadSignals,
		SubsectorRotations:       subsectorSignals,
		PerformanceLedIndustries: topIndustrySignals(industries, 10),
		StockCandidates:          stocks,
		RejectedStockCounts:      rejected,
		DataQuality: types.MarketRotationDataQuality{
			Source:                "TradingView sector ETFs and USA stock industries",
			SourceURL:             marketRotationSourceURL,
			InputETFCount:         len(overview),
			EligibleETFCount:      len(joined),
			ExcludedETFCount:      len(overview) - len(joined),
			ExclusionsByReason:    exclusions,
			UnmappedIndustryCount: unmapped,
		},
		BriefInstructions: []string{
			"Use the precomputed values and rankings exactly; do not recalculate or reorder them.",
			"Start with the 1M/3M sector fund-flow landscape, then performance alignment and divergences.",
			"Cover broad-sector rotations, subsector rotations, performance-led opportunities, and zero to five stock candidates.",
			"Describe ETF product flows, never direct stock-level institutional flows.",
			"End with data-quality, exclusion, and invalidation notes.",
		},
	}
	brief.LostConfirmations = detectLostConfirmations(previous, brief)

	fingerprint, err := rotationFingerprint(brief)
	if err != nil {
		return nil, err
	}
	if previous != nil && previous.Fingerprint == fingerprint {
		brief.DataQuality.Stale = true
		brief.DataQuality.Warnings = append(brief.DataQuality.Warnings, "The deterministic input fingerprint is unchanged from the latest stored session.")
	}
	if options.PersistHistory {
		if db == nil {
			return nil, errors.New("market rotation history persistence requires a database")
		}
		if !brief.DataQuality.Stale {
			if err := persistRotationSnapshot(db, *brief, fingerprint); err != nil {
				return nil, err
			}
			brief.DataQuality.Persisted = true
		}
	}
	return brief, nil
}

type databaseStore interface {
	Get(key string, value interface{}) error
	Put(key string, value interface{}) error
	Delete(key string) error
	GetAllKeysWithPrefix(prefix string) ([]string, error)
}

func joinAndFilterSectorETFs(overviews []types.ETFSectorOverview, performances []types.ETFFundFlowPerformance, flows []types.ETFFundFlows) ([]joinedSectorETF, map[string]int) {
	perfByID := make(map[string]types.ETFFundFlowPerformance, len(performances))
	flowByID := make(map[string]types.ETFFundFlows, len(flows))
	for _, row := range performances {
		perfByID[row.ID] = row
	}
	for _, row := range flows {
		flowByID[row.ID] = row
	}
	exclusions := make(map[string]int)
	result := make([]joinedSectorETF, 0, len(overviews))
	for _, row := range overviews {
		perf, perfOK := perfByID[row.ID]
		flow, flowOK := flowByID[row.ID]
		if !perfOK || !flowOK {
			exclusions["missing_joined_tab"]++
			continue
		}
		reason := exclusionReason(row, flow)
		if reason != "" {
			exclusions[reason]++
			continue
		}
		if row.AssetsUnderManagement == nil || *row.AssetsUnderManagement <= 0 || flow.OneMonth == nil || flow.ThreeMonths == nil || !completePerformance(perf) {
			exclusions["missing_required_value"]++
			continue
		}
		class := "subsector"
		sector := canonicalETFSector(row)
		if broadSector, ok := broadSectorETFs[strings.ToUpper(row.Ticker)]; ok {
			class = "broad"
			sector = broadSector
		}
		if sector == "" {
			exclusions["unsupported_focus"]++
			continue
		}
		aum := *row.AssetsUnderManagement
		flow1 := *flow.OneMonth
		flow3 := *flow.ThreeMonths
		result = append(result, joinedSectorETF{
			overview: row, performance: perf, flows: flow, class: class, sector: sector,
			flow1Pct:   100 * flow1 / aum,
			flow3Pct:   100 * flow3 / aum,
			flowPrior2: 100 * ((flow3 - flow1) / 2) / aum,
			flowAccel:  100 * (flow1 - ((flow3 - flow1) / 2)) / aum,
			perf:       rotationPerformance(perf.Change, perf.OneWeek, perf.OneMonth, perf.ThreeMonths, perf.SixMonths, perf.OneYear),
		})
	}
	return result, exclusions
}

func exclusionReason(row types.ETFSectorOverview, flow types.ETFFundFlows) string {
	if !isUSExchange(row.Exchange) {
		return "foreign_listing"
	}
	if row.AUMCurrency != "USD" || row.Currency != "USD" || flow.Currency != "USD" {
		return "non_usd"
	}
	if !strings.EqualFold(row.AssetClass, "Equity") {
		return "non_equity"
	}
	name := strings.ToLower(row.Fund + " " + row.Ticker)
	for _, token := range []string{" leveraged", " inverse", " ultra", "ultrapro", " bull ", " bear ", "daily ", " daily", " 2x", " 3x", "-2x", "-3x", "single-stock", "single stock"} {
		if strings.Contains(name, token) {
			return "leveraged_inverse_or_single_stock"
		}
	}
	return ""
}

func isUSExchange(exchange string) bool {
	switch strings.ToUpper(strings.TrimSpace(exchange)) {
	case "AMEX", "NASDAQ", "NYSE", "NYSEARCA", "ARCA", "CBOE", "BATS":
		return true
	default:
		return false
	}
}

func canonicalETFSector(row types.ETFSectorOverview) string {
	if strings.EqualFold(row.Focus, "Theme") {
		if mapping, ok := thematicMappingForETF(row.Ticker, row.Fund); ok {
			return mapping.sector
		}
		return "Thematic"
	}
	for _, sector := range []string{"Communication services", "Consumer discretionary", "Consumer staples", "Energy", "Financials", "Health care", "Industrials", "Information technology", "Materials", "Real estate", "Utilities"} {
		if strings.EqualFold(row.Focus, sector) {
			return sector
		}
	}
	return ""
}

func thematicMappingForETF(ticker, fund string) (thematicETFMapping, bool) {
	upperTicker := strings.ToUpper(strings.TrimSpace(ticker))
	searchText := " " + strings.ToLower(fund+" "+ticker) + " "
	for _, mapping := range thematicETFMappings {
		for _, candidate := range mapping.tickers {
			if upperTicker == candidate {
				return mapping, true
			}
		}
		for _, keyword := range mapping.keywords {
			if strings.Contains(searchText, keyword) {
				return mapping, true
			}
		}
	}
	return thematicETFMapping{}, false
}

func mappedIndustriesForETF(row joinedSectorETF) []string {
	if industries := subsectorIndustryHints[strings.ToUpper(row.overview.Ticker)]; len(industries) > 0 {
		return industries
	}
	if strings.EqualFold(row.overview.Focus, "Theme") {
		if mapping, ok := thematicMappingForETF(row.overview.Ticker, row.overview.Fund); ok {
			return mapping.industries
		}
	}
	return nil
}

func aggregateSectorFlows(etfs []joinedSectorETF) []types.SectorFundFlowSummary {
	groups := make(map[string]*sectorAccumulator)
	for _, etf := range etfs {
		group := groups[etf.sector]
		if group == nil {
			group = &sectorAccumulator{sector: etf.sector}
			groups[etf.sector] = group
		}
		group.etfs = append(group.etfs, etf)
		group.aum += *etf.overview.AssetsUnderManagement
		group.flow1 += *etf.flows.OneMonth
		group.flow3 += *etf.flows.ThreeMonths
		if *etf.flows.OneMonth > 0 {
			group.positive++
		}
		if etf.class == "broad" {
			group.broad = append(group.broad, etf)
			group.broadFlow1 += *etf.flows.OneMonth
			group.broadFlow3 += *etf.flows.ThreeMonths
			group.broadAUM += *etf.overview.AssetsUnderManagement
		} else {
			group.subsector = append(group.subsector, etf)
			group.subFlow1 += *etf.flows.OneMonth
			group.subFlow3 += *etf.flows.ThreeMonths
			group.subAUM += *etf.overview.AssetsUnderManagement
		}
	}
	result := make([]types.SectorFundFlowSummary, 0, len(groups))
	for _, group := range groups {
		priorFlow := (group.flow3 - group.flow1) / 2
		broadPriorFlow := (group.broadFlow3 - group.broadFlow1) / 2
		subPriorFlow := (group.subFlow3 - group.subFlow1) / 2
		flowPct := safePercent(group.flow1, group.aum)
		threeMonthFlowPct := safePercent(group.flow3, group.aum)
		priorPct := safePercent(priorFlow, group.aum)
		broadPerf := weightedPerformance(group.broad)
		result = append(result, types.SectorFundFlowSummary{
			Sector: group.sector, ETFCount: len(group.etfs), BroadETFCount: len(group.broad), SubsectorETFCount: len(group.subsector),
			PositiveFlowETFCount: group.positive, PositiveFlowBreadthPercent: round(safePercent(float64(group.positive), float64(len(group.etfs))), 1),
			CombinedAUMUSD: round(group.aum, 0), OneMonthFlowUSD: round(group.flow1, 0), ThreeMonthFlowUSD: round(group.flow3, 0),
			OneMonthFlowPercentAUM: round(flowPct, 3), ThreeMonthFlowPercentAUM: round(threeMonthFlowPct, 3), PriorTwoMonthMonthlyFlowUSD: round(priorFlow, 0),
			PriorTwoMonthMonthlyPercentAUM: round(priorPct, 3), FlowAccelerationPercentAUM: round(flowPct-priorPct, 3),
			BroadOneMonthFlowUSD: round(group.broadFlow1, 0), SubsectorOneMonthFlowUSD: round(group.subFlow1, 0),
			BroadFlowAccelerationPercentAUM:     round(safePercent(group.broadFlow1-broadPriorFlow, group.broadAUM), 3),
			SubsectorFlowAccelerationPercentAUM: round(safePercent(group.subFlow1-subPriorFlow, group.subAUM), 3),
			BroadSubsectorRelationship:          broadSubsectorRelationship(group.broad, group.subsector, group.broadFlow1, group.subFlow1),
			FlowTrend:                           flowTrend(group.flow1, priorFlow), BroadPerformance: broadPerf,
			PerformanceFlowAlignment: performanceFlowAlignment(group.flow1, broadPerf),
			LargestContributors:      contributions(group.etfs, true), LargestDetractors: contributions(group.etfs, false),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return sectorFlowSortScore(result[i]) > sectorFlowSortScore(result[j])
	})
	return result
}

func sectorFlowSortScore(row types.SectorFundFlowSummary) float64 {
	return 0.5*row.OneMonthFlowPercentAUM + 0.5*row.ThreeMonthFlowPercentAUM
}

func applySectorFlowTransitions(rows []types.SectorFundFlowSummary, previous *marketRotationSnapshot) {
	if previous == nil {
		for i := range rows {
			rows[i].HistoryTransition = "baseline"
		}
		return
	}
	prior := make(map[string]types.SectorFundFlowSummary, len(previous.Brief.SectorFundFlows))
	for _, row := range previous.Brief.SectorFundFlows {
		prior[row.Sector] = row
	}
	for i := range rows {
		old, ok := prior[rows[i].Sector]
		if !ok {
			rows[i].HistoryTransition = "new"
			continue
		}
		current := rows[i].FlowAccelerationPercentAUM
		before := old.FlowAccelerationPercentAUM
		switch {
		case before <= 0 && current > 0:
			rows[i].HistoryTransition = "newly_accelerating"
		case current >= before+0.25:
			rows[i].HistoryTransition = "strengthening"
		case current <= before-0.25:
			rows[i].HistoryTransition = "weakening"
		default:
			rows[i].HistoryTransition = "unchanged"
		}
	}
}

func contributions(etfs []joinedSectorETF, descending bool) []types.ETFContribution {
	rows := make([]joinedSectorETF, 0, len(etfs))
	for _, row := range etfs {
		if descending && *row.flows.OneMonth > 0 {
			rows = append(rows, row)
		}
		if !descending && *row.flows.OneMonth < 0 {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if descending {
			return *rows[i].flows.OneMonth > *rows[j].flows.OneMonth
		}
		return *rows[i].flows.OneMonth < *rows[j].flows.OneMonth
	})
	if len(rows) > 3 {
		rows = rows[:3]
	}
	result := make([]types.ETFContribution, 0, len(rows))
	for _, row := range rows {
		result = append(result, types.ETFContribution{Ticker: row.overview.Ticker, Fund: row.overview.Fund, Classification: row.class, OneMonthFlow: round(*row.flows.OneMonth, 0)})
	}
	return result
}

func broadSubsectorRelationship(broad, subsector []joinedSectorETF, broadFlow, subsectorFlow float64) string {
	if len(broad) == 0 || len(subsector) == 0 {
		return "insufficient_comparison"
	}
	if broadFlow >= 0 && subsectorFlow >= 0 {
		return "aligned_inflow"
	}
	if broadFlow < 0 && subsectorFlow < 0 {
		return "aligned_outflow"
	}
	if broadFlow >= 0 {
		return "broad_inflow_subsector_outflow"
	}
	return "broad_outflow_subsector_inflow"
}

func flowTrend(current, prior float64) string {
	switch {
	case current >= 0 && current >= prior:
		return "accelerating_inflow"
	case current >= 0:
		return "decelerating_inflow"
	case current > prior:
		return "improving_outflow"
	default:
		return "worsening_outflow"
	}
}

func performanceFlowAlignment(flow float64, perf types.RotationPerformance) string {
	performancePositive := perf.OneMonth > 0 && perf.Acceleration > 0
	if flow > 0 && performancePositive {
		return "aligned_positive"
	}
	if flow < 0 && !performancePositive {
		return "aligned_negative"
	}
	if performancePositive {
		return "performance_leads_flow"
	}
	return "flow_leads_performance"
}

func weightedPerformance(etfs []joinedSectorETF) types.RotationPerformance {
	var result types.RotationPerformance
	var total float64
	for _, etf := range etfs {
		weight := *etf.overview.AssetsUnderManagement
		total += weight
		result.OneDay += etf.perf.OneDay * weight
		result.OneWeek += etf.perf.OneWeek * weight
		result.OneMonth += etf.perf.OneMonth * weight
		result.ThreeMonths += etf.perf.ThreeMonths * weight
		result.SixMonths += etf.perf.SixMonths * weight
		result.OneYear += etf.perf.OneYear * weight
	}
	if total == 0 {
		return result
	}
	result.OneDay /= total
	result.OneWeek /= total
	result.OneMonth /= total
	result.ThreeMonths /= total
	result.SixMonths /= total
	result.OneYear /= total
	result.PriorTwoMonthPace = priorTwoMonthReturnPace(result.OneMonth, result.ThreeMonths)
	result.Acceleration = result.OneMonth - result.PriorTwoMonthPace
	return roundedPerformance(result)
}

func buildBroadSectorSignals(flows []types.SectorFundFlowSummary, previous *marketRotationSnapshot) []types.SectorRotationSignal {
	accels, sixMonths, threeMonths := make([]float64, 0, len(flows)), make([]float64, 0, len(flows)), make([]float64, 0, len(flows))
	for _, row := range flows {
		accels = append(accels, row.BroadPerformance.Acceleration)
		sixMonths = append(sixMonths, row.BroadPerformance.SixMonths)
		threeMonths = append(threeMonths, row.BroadPerformance.ThreeMonths)
	}
	result := make([]types.SectorRotationSignal, 0)
	for _, row := range flows {
		perf := row.BroadPerformance
		setup := setupType(perf, percentileRank(sixMonths, perf.SixMonths), percentileRank(accels, perf.Acceleration), percentileRank(threeMonths, perf.ThreeMonths))
		aligned := row.BroadOneMonthFlowUSD > 0 && row.BroadFlowAccelerationPercentAUM > 0 && perf.OneWeek > 0 && perf.OneMonth > 0 && perf.Acceleration > 0
		if !aligned || setup == "unconfirmed" {
			continue
		}
		score := 100 * (0.20*percentileRank(flowMetric(flows, func(x types.SectorFundFlowSummary) float64 { return x.OneMonthFlowPercentAUM }), row.OneMonthFlowPercentAUM) +
			0.20*percentileRank(flowMetric(flows, func(x types.SectorFundFlowSummary) float64 { return x.ThreeMonthFlowPercentAUM }), row.ThreeMonthFlowPercentAUM) +
			0.25*percentileRank(flowMetric(flows, func(x types.SectorFundFlowSummary) float64 { return x.FlowAccelerationPercentAUM }), row.FlowAccelerationPercentAUM) +
			0.35*percentileRank(accels, perf.Acceleration))
		result = append(result, types.SectorRotationSignal{Sector: row.Sector, Setup: setup, Score: round(score, 1), Transition: sectorTransition(previous, row.Sector, score), FlowTrend: row.FlowTrend, Performance: perf})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })
	return result
}

func buildSubsectorSignals(etfs []joinedSectorETF, previous *marketRotationSnapshot) []types.SubsectorRotationSignal {
	accels, sixMonths, threeMonths := make([]float64, 0), make([]float64, 0), make([]float64, 0)
	for _, row := range etfs {
		if row.class == "subsector" {
			accels = append(accels, row.perf.Acceleration)
			sixMonths = append(sixMonths, row.perf.SixMonths)
			threeMonths = append(threeMonths, row.perf.ThreeMonths)
		}
	}
	result := make([]types.SubsectorRotationSignal, 0)
	for _, row := range etfs {
		if row.class != "subsector" || *row.flows.OneMonth <= 0 || row.flowAccel <= 0 || row.perf.OneWeek <= 0 || row.perf.OneMonth <= 0 || row.perf.Acceleration <= 0 {
			continue
		}
		setup := setupType(row.perf, percentileRank(sixMonths, row.perf.SixMonths), percentileRank(accels, row.perf.Acceleration), percentileRank(threeMonths, row.perf.ThreeMonths))
		if setup == "unconfirmed" {
			continue
		}
		score := 100 * (0.20*percentileRank(etfMetric(etfs, func(x joinedSectorETF) float64 { return x.flow1Pct }), row.flow1Pct) +
			0.20*percentileRank(etfMetric(etfs, func(x joinedSectorETF) float64 { return x.flow3Pct }), row.flow3Pct) +
			0.25*percentileRank(etfMetric(etfs, func(x joinedSectorETF) float64 { return x.flowAccel }), row.flowAccel) +
			0.35*percentileRank(accels, row.perf.Acceleration))
		result = append(result, types.SubsectorRotationSignal{
			Ticker: row.overview.Ticker, Fund: row.overview.Fund, Sector: row.sector, MappedIndustries: mappedIndustriesForETF(row),
			Setup: setup, Score: round(score, 1), Transition: subsectorTransition(previous, row.overview.Ticker, score), OneMonthFlowUSD: round(*row.flows.OneMonth, 0),
			FlowPercentAUM: round(row.flow1Pct, 3), ThreeMonthFlowPercentAUM: round(row.flow3Pct, 3), FlowAcceleration: round(row.flowAccel, 3), Performance: row.perf,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

func buildIndustrySignals(overviews []types.USAIndustryOverview, performances []types.USAIndustryPerformance, sectorFlows []types.SectorFundFlowSummary, broad []types.SectorRotationSignal, subsectors []types.SubsectorRotationSignal, previous *marketRotationSnapshot) ([]industryAnalysis, int) {
	overviewByID := make(map[string]types.USAIndustryOverview, len(overviews))
	for _, row := range overviews {
		overviewByID[row.ID] = row
	}
	flowBySector := make(map[string]types.SectorFundFlowSummary, len(sectorFlows))
	for _, row := range sectorFlows {
		flowBySector[row.Sector] = row
	}
	broadBySector := make(map[string]types.SectorRotationSignal, len(broad))
	for _, row := range broad {
		broadBySector[row.Sector] = row
	}
	hinted := make(map[string]types.SubsectorRotationSignal)
	for _, row := range subsectors {
		for _, industry := range row.MappedIndustries {
			if existing, ok := hinted[industry]; !ok || row.Score > existing.Score {
				hinted[industry] = row
			}
		}
	}

	accels, sixMonths, threeMonths, days, weeks, months := make([]float64, 0), make([]float64, 0), make([]float64, 0), make([]float64, 0), make([]float64, 0), make([]float64, 0)
	joined := make([]struct {
		overview types.USAIndustryOverview
		perf     types.USAIndustryPerformance
		shape    types.RotationPerformance
	}, 0, len(performances))
	unmapped := 0
	for _, perf := range performances {
		overview, ok := overviewByID[perf.ID]
		if !ok || !completeIndustryPerformance(perf) {
			continue
		}
		if _, ok := mapIndustryToETFSector(overview.Sector, overview.Industry); !ok {
			unmapped++
			continue
		}
		shape := rotationPerformance(perf.Change, perf.OneWeek, perf.OneMonth, perf.ThreeMonths, perf.SixMonths, perf.OneYear)
		joined = append(joined, struct {
			overview types.USAIndustryOverview
			perf     types.USAIndustryPerformance
			shape    types.RotationPerformance
		}{overview, perf, shape})
		accels, sixMonths, threeMonths, days = append(accels, shape.Acceleration), append(sixMonths, shape.SixMonths), append(threeMonths, shape.ThreeMonths), append(days, shape.OneDay)
		weeks, months = append(weeks, shape.OneWeek), append(months, shape.OneMonth)
	}

	result := make([]industryAnalysis, 0)
	for _, row := range joined {
		sector, _ := mapIndustryToETFSector(row.overview.Sector, row.overview.Industry)
		setup := setupType(row.shape, percentileRank(sixMonths, row.shape.SixMonths), percentileRank(accels, row.shape.Acceleration), percentileRank(threeMonths, row.shape.ThreeMonths))
		spike := oneDaySpike(row.shape, percentileRank(days, row.shape.OneDay))
		if row.shape.OneWeek <= 0 || row.shape.OneMonth <= 0 || row.shape.Acceleration <= 0 || spike || setup == "unconfirmed" {
			continue
		}
		lane := "performance_led"
		confidence := 0.60
		alignmentScore := 0.60
		if broadSignal, ok := broadBySector[sector]; ok {
			lane, confidence, alignmentScore = "broad_sector_confirmed", 1.0, broadSignal.Score/100
		}
		if hint, ok := hinted[row.overview.Industry]; ok && confidence < 0.85 {
			lane, confidence, alignmentScore = "subsector_rotation", 0.85, hint.Score/100
		}
		if flow, ok := flowBySector[sector]; ok && flow.PerformanceFlowAlignment == "performance_leads_flow" && lane == "performance_led" {
			confidence = 0.55
		}
		score := 100 * (0.60*(0.50*percentileRank(accels, row.shape.Acceleration)+0.25*percentileRank(weeks, row.shape.OneWeek)+0.25*percentileRank(months, row.shape.OneMonth)) + 0.40*alignmentScore)
		result = append(result, industryAnalysis{signal: types.IndustryRotationSignal{
			ID: row.overview.ID, Industry: row.overview.Industry, StockSector: row.overview.Sector, ETFSector: sector,
			Lane: lane, Setup: setup, Score: round(score, 1), Transition: industryTransition(previous, row.overview.Industry, score), Performance: row.shape,
		}, confidence: confidence})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].signal.Score > result[j].signal.Score })
	return result, unmapped
}

func buildStockCandidates(screener MarketDataScreener, industries []industryAnalysis, previous *marketRotationSnapshot, maxCandidates int, rejected map[string]int) ([]types.StockRotationCandidate, error) {
	selected := selectIndustryDrilldowns(industries)
	all := make([]types.StockRotationCandidate, 0)
	for _, industry := range selected {
		overviews, err := screener.FetchUSAIndustryStocksOverview(industry.signal.Industry)
		if err != nil {
			return nil, fmt.Errorf("fetch %s stock overview: %w", industry.signal.Industry, err)
		}
		performances, err := screener.FetchUSAIndustryStocksPerformance(industry.signal.Industry)
		if err != nil {
			return nil, fmt.Errorf("fetch %s stock performance: %w", industry.signal.Industry, err)
		}
		all = append(all, scoreIndustryStocks(industry, overviews, performances, previous, rejected)...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Score > all[j].Score })
	if len(all) > maxCandidates {
		all = all[:maxCandidates]
	}
	return all, nil
}

func scoreIndustryStocks(industry industryAnalysis, overviews []types.USAIndustryStockOverview, performances []types.USAIndustryStockPerformance, previous *marketRotationSnapshot, rejected map[string]int) []types.StockRotationCandidate {
	overviewByID := make(map[string]types.USAIndustryStockOverview, len(overviews))
	for _, row := range overviews {
		overviewByID[row.ID] = row
	}
	type joinedStock struct {
		overview types.USAIndustryStockOverview
		shape    types.RotationPerformance
		vol      float64
		turnover float64
	}
	base := make([]joinedStock, 0)
	for _, perf := range performances {
		overview, ok := overviewByID[perf.ID]
		if !ok || !completeStockPerformance(perf) || overview.MarketCap == nil || overview.Price == nil || overview.Volume == nil {
			rejected["missing_required_value"]++
			continue
		}
		turnover := *overview.Price * *overview.Volume
		switch {
		case overview.Currency != "USD":
			rejected["non_usd_stock"]++
		case *overview.MarketCap < 2_000_000_000:
			rejected["market_cap_below_2b"]++
		case *overview.Price < 5:
			rejected["price_below_5"]++
		case turnover < 20_000_000:
			rejected["turnover_below_20m"]++
		default:
			base = append(base, joinedStock{overview: overview, shape: rotationPerformance(perf.Change, perf.OneWeek, perf.OneMonth, perf.ThreeMonths, perf.SixMonths, perf.OneYear), vol: *perf.VolatilityOneMonth, turnover: turnover})
		}
	}
	if len(base) == 0 {
		return nil
	}
	vols, accels, sixMonths, threeMonths, days, weeks, months := make([]float64, 0), make([]float64, 0), make([]float64, 0), make([]float64, 0), make([]float64, 0), make([]float64, 0), make([]float64, 0)
	for _, row := range base {
		vols, accels, sixMonths = append(vols, row.vol), append(accels, row.shape.Acceleration), append(sixMonths, row.shape.SixMonths)
		threeMonths, days, weeks, months = append(threeMonths, row.shape.ThreeMonths), append(days, row.shape.OneDay), append(weeks, row.shape.OneWeek), append(months, row.shape.OneMonth)
	}
	result := make([]types.StockRotationCandidate, 0)
	for _, row := range base {
		if percentileRank(vols, row.vol) > 0.80 {
			rejected["highest_volatility_quintile"]++
			continue
		}
		setup := setupType(row.shape, percentileRank(sixMonths, row.shape.SixMonths), percentileRank(accels, row.shape.Acceleration), percentileRank(threeMonths, row.shape.ThreeMonths))
		if row.shape.ThreeMonths <= 0 {
			rejected["three_month_performance_non_positive"]++
			continue
		}
		if row.shape.OneWeek <= 0 || row.shape.OneMonth <= 0 || row.shape.Acceleration <= 0 || setup == "unconfirmed" {
			rejected["unconfirmed_return_shape"]++
			continue
		}
		if oneDaySpike(row.shape, percentileRank(days, row.shape.OneDay)) {
			rejected["one_day_spike"]++
			continue
		}
		performanceScore := 0.25*percentileRank(weeks, row.shape.OneWeek) + 0.25*percentileRank(months, row.shape.OneMonth) + 0.50*percentileRank(accels, row.shape.Acceleration)
		score := 100 * (0.60*performanceScore + 0.30*industry.confidence + 0.10*historyScore(previous, row.overview.Ticker))
		result = append(result, types.StockRotationCandidate{
			Ticker: row.overview.Ticker, Company: row.overview.Company, Industry: industry.signal.Industry, ETFSector: industry.signal.ETFSector,
			Lane: industry.signal.Lane, Setup: setup, Score: round(score, 1), Transition: stockTransition(previous, row.overview.Ticker, score),
			MarketCapUSD: round(*row.overview.MarketCap, 0), DailyTurnoverUSD: round(row.turnover, 0), MonthlyVolatility: round(row.vol, 3),
			EPSGrowthYoY: row.overview.EPSDilutedGrowthYoYTTM, AnalystRating: row.overview.AnalystRating, Performance: row.shape,
			Invalidation: "The setup loses confirmation if 1W, 1M, or 3M performance turns non-positive, return acceleration turns negative, or its sector/industry lane loses confirmation.",
		})
	}
	return result
}

func selectIndustryDrilldowns(industries []industryAnalysis) []industryAnalysis {
	if len(industries) <= maxIndustryDrilldowns {
		return append([]industryAnalysis(nil), industries...)
	}
	return append([]industryAnalysis(nil), industries[:maxIndustryDrilldowns]...)
}

func topIndustrySignals(industries []industryAnalysis, limit int) []types.IndustryRotationSignal {
	if len(industries) < limit {
		limit = len(industries)
	}
	result := make([]types.IndustryRotationSignal, 0, limit)
	for _, row := range industries[:limit] {
		result = append(result, row.signal)
	}
	return result
}

func mapIndustryToETFSector(stockSector, industry string) (string, bool) {
	stockSector = strings.ToLower(strings.TrimSpace(stockSector))
	industryLower := strings.ToLower(strings.TrimSpace(industry))
	containsAny := func(values ...string) bool {
		for _, value := range values {
			if strings.Contains(industryLower, value) {
				return true
			}
		}
		return false
	}
	switch stockSector {
	case "communications":
		return "Communication services", true
	case "energy minerals":
		return "Energy", true
	case "health services", "health technology":
		return "Health care", true
	case "non-energy minerals":
		return "Materials", true
	case "transportation":
		return "Industrials", true
	case "utilities":
		return "Utilities", true
	case "finance":
		if containsAny("real estate investment trust", "real estate development") {
			return "Real estate", true
		}
		return "Financials", true
	case "electronic technology":
		if containsAny("aerospace & defense") {
			return "Industrials", true
		}
		return "Information technology", true
	case "technology services":
		if containsAny("internet software/services") {
			return "Communication services", true
		}
		return "Information technology", true
	case "industrial services":
		if containsAny("oil & gas pipeline", "contract drilling", "oilfield services") {
			return "Energy", true
		}
		return "Industrials", true
	case "process industries":
		if containsAny("agricultural commodities") {
			return "Consumer staples", true
		}
		return "Materials", true
	case "consumer non-durables":
		if containsAny("apparel/footwear") {
			return "Consumer discretionary", true
		}
		return "Consumer staples", true
	case "consumer durables":
		return "Consumer discretionary", true
	case "consumer services":
		if containsAny("movies/entertainment", "media conglomerates", "cable/satellite", "broadcasting", "publishing") {
			return "Communication services", true
		}
		return "Consumer discretionary", true
	case "retail trade":
		if containsAny("food retail", "drugstore", "discount stores") {
			return "Consumer staples", true
		}
		return "Consumer discretionary", true
	case "distribution services":
		if containsAny("medical distributors") {
			return "Health care", true
		}
		if containsAny("food distributors") {
			return "Consumer staples", true
		}
		return "Industrials", true
	case "commercial services":
		if containsAny("advertising/marketing") {
			return "Communication services", true
		}
		return "Industrials", true
	case "producer manufacturing":
		if containsAny("auto parts") {
			return "Consumer discretionary", true
		}
		return "Industrials", true
	default:
		return "", false
	}
}

func setupType(perf types.RotationPerformance, sixMonthPercentile, accelerationPercentile, threeMonthPercentile float64) string {
	if perf.SixMonths <= 0 || perf.OneYear <= 0 {
		return "recovery"
	}
	if sixMonthPercentile <= 0.60 && accelerationPercentile >= 0.75 {
		return "ignition"
	}
	if threeMonthPercentile >= 0.75 && sixMonthPercentile >= 0.75 {
		return "continuation"
	}
	return "unconfirmed"
}

func oneDaySpike(perf types.RotationPerformance, dayPercentile float64) bool {
	return perf.OneDay > 0 && perf.OneWeek > 0 && perf.OneDay/perf.OneWeek > 0.70 && dayPercentile >= 0.90
}

func rotationPerformance(day, week, month, three, six, year *float64) types.RotationPerformance {
	result := types.RotationPerformance{OneDay: value(day), OneWeek: value(week), OneMonth: value(month), ThreeMonths: value(three), SixMonths: value(six), OneYear: value(year)}
	result.PriorTwoMonthPace = priorTwoMonthReturnPace(result.OneMonth, result.ThreeMonths)
	result.Acceleration = result.OneMonth - result.PriorTwoMonthPace
	return roundedPerformance(result)
}

func roundedPerformance(result types.RotationPerformance) types.RotationPerformance {
	result.OneDay = round(result.OneDay, 3)
	result.OneWeek = round(result.OneWeek, 3)
	result.OneMonth = round(result.OneMonth, 3)
	result.ThreeMonths = round(result.ThreeMonths, 3)
	result.SixMonths = round(result.SixMonths, 3)
	result.OneYear = round(result.OneYear, 3)
	result.PriorTwoMonthPace = round(result.PriorTwoMonthPace, 3)
	result.Acceleration = round(result.Acceleration, 3)
	return result
}

func priorTwoMonthReturnPace(oneMonth, threeMonth float64) float64 {
	denominator := 1 + oneMonth/100
	if denominator <= 0 || 1+threeMonth/100 <= 0 {
		return 0
	}
	return 100 * (math.Sqrt((1+threeMonth/100)/denominator) - 1)
}

func completePerformance(row types.ETFFundFlowPerformance) bool {
	return row.Change != nil && row.OneWeek != nil && row.OneMonth != nil && row.ThreeMonths != nil && row.SixMonths != nil && row.OneYear != nil
}

func completeIndustryPerformance(row types.USAIndustryPerformance) bool {
	return row.Change != nil && row.OneWeek != nil && row.OneMonth != nil && row.ThreeMonths != nil && row.SixMonths != nil && row.OneYear != nil
}

func completeStockPerformance(row types.USAIndustryStockPerformance) bool {
	return row.Change != nil && row.OneWeek != nil && row.OneMonth != nil && row.ThreeMonths != nil && row.SixMonths != nil && row.OneYear != nil && row.VolatilityOneMonth != nil
}

func percentileRank(values []float64, value float64) float64 {
	if len(values) == 0 {
		return 0
	}
	count := 0
	for _, candidate := range values {
		if candidate <= value {
			count++
		}
	}
	return float64(count) / float64(len(values))
}

func flowMetric(rows []types.SectorFundFlowSummary, metric func(types.SectorFundFlowSummary) float64) []float64 {
	result := make([]float64, 0, len(rows))
	for _, row := range rows {
		result = append(result, metric(row))
	}
	return result
}

func etfMetric(rows []joinedSectorETF, metric func(joinedSectorETF) float64) []float64 {
	result := make([]float64, 0, len(rows))
	for _, row := range rows {
		if row.class == "subsector" {
			result = append(result, metric(row))
		}
	}
	return result
}

func safePercent(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return 100 * numerator / denominator
}

func value(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func round(value float64, places int) float64 {
	factor := math.Pow10(places)
	return math.Round(value*factor) / factor
}

func rotationFingerprint(brief *types.MarketRotationBrief) (string, error) {
	type rankedShape struct {
		Key         string                    `json:"key"`
		Setup       string                    `json:"setup"`
		Score       float64                   `json:"score"`
		Performance types.RotationPerformance `json:"performance"`
	}
	normalizedFlows := append([]types.SectorFundFlowSummary(nil), brief.SectorFundFlows...)
	for i := range normalizedFlows {
		normalizedFlows[i].HistoryTransition = ""
	}
	payload := struct {
		Version    string                        `json:"version"`
		Flows      []types.SectorFundFlowSummary `json:"flows"`
		Broad      []rankedShape                 `json:"broad"`
		Subsectors []rankedShape                 `json:"subsectors"`
		Industries []rankedShape                 `json:"industries"`
		Stocks     []rankedShape                 `json:"stocks"`
	}{Version: brief.MethodologyVersion, Flows: normalizedFlows}
	for _, row := range brief.BroadSectorRotations {
		payload.Broad = append(payload.Broad, rankedShape{row.Sector, row.Setup, row.Score, row.Performance})
	}
	for _, row := range brief.SubsectorRotations {
		payload.Subsectors = append(payload.Subsectors, rankedShape{row.Ticker, row.Setup, row.Score, row.Performance})
	}
	for _, row := range brief.PerformanceLedIndustries {
		payload.Industries = append(payload.Industries, rankedShape{row.ID, row.Setup, row.Score, row.Performance})
	}
	for _, row := range brief.StockCandidates {
		payload.Stocks = append(payload.Stocks, rankedShape{row.Ticker, row.Setup, row.Score, row.Performance})
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal rotation fingerprint: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func detectLostConfirmations(previous *marketRotationSnapshot, current *types.MarketRotationBrief) []types.RotationLostConfirmation {
	if previous == nil {
		return nil
	}
	currentKeys := make(map[string]bool)
	for _, row := range current.BroadSectorRotations {
		currentKeys["broad:"+row.Sector] = true
	}
	for _, row := range current.SubsectorRotations {
		currentKeys["subsector:"+row.Ticker] = true
	}
	for _, row := range current.StockCandidates {
		currentKeys["stock:"+row.Ticker] = true
	}
	result := make([]types.RotationLostConfirmation, 0)
	for _, row := range previous.Brief.BroadSectorRotations {
		if !currentKeys["broad:"+row.Sector] {
			result = append(result, types.RotationLostConfirmation{Kind: "broad_sector", Key: row.Sector, PreviousScore: row.Score})
		}
	}
	for _, row := range previous.Brief.SubsectorRotations {
		if !currentKeys["subsector:"+row.Ticker] {
			result = append(result, types.RotationLostConfirmation{Kind: "subsector", Key: row.Ticker, PreviousScore: row.Score})
		}
	}
	for _, row := range previous.Brief.StockCandidates {
		if !currentKeys["stock:"+row.Ticker] {
			result = append(result, types.RotationLostConfirmation{Kind: "stock", Key: row.Ticker, PreviousScore: row.Score})
		}
	}
	return result
}

func loadLatestRotationSnapshot(db databaseStore) (*marketRotationSnapshot, error) {
	if db == nil {
		return nil, nil
	}
	keys, err := db.GetAllKeysWithPrefix(string(types.MarketRotationKeyPrefix))
	if err != nil || len(keys) == 0 {
		return nil, err
	}
	sort.Strings(keys)
	var snapshot marketRotationSnapshot
	if err := db.Get(keys[len(keys)-1], &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func persistRotationSnapshot(db databaseStore, brief types.MarketRotationBrief, fingerprint string) error {
	key := fmt.Sprintf("%s:%s", types.MarketRotationKeyPrefix, brief.USSessionDate)
	if err := db.Put(key, marketRotationSnapshot{Fingerprint: fingerprint, Brief: brief}); err != nil {
		return fmt.Errorf("persist market rotation snapshot: %w", err)
	}
	keys, err := db.GetAllKeysWithPrefix(string(types.MarketRotationKeyPrefix))
	if err != nil {
		return fmt.Errorf("list market rotation history: %w", err)
	}
	sort.Strings(keys)
	for len(keys) > maxRotationHistory {
		if err := db.Delete(keys[0]); err != nil {
			return fmt.Errorf("prune market rotation history: %w", err)
		}
		keys = keys[1:]
	}
	return nil
}

func sectorTransition(previous *marketRotationSnapshot, sector string, score float64) string {
	if previous == nil {
		return "baseline"
	}
	for _, row := range previous.Brief.BroadSectorRotations {
		if row.Sector == sector {
			return scoreTransition(row.Score, score)
		}
	}
	return "new"
}

func subsectorTransition(previous *marketRotationSnapshot, ticker string, score float64) string {
	if previous == nil {
		return "baseline"
	}
	for _, row := range previous.Brief.SubsectorRotations {
		if row.Ticker == ticker {
			return scoreTransition(row.Score, score)
		}
	}
	return "new"
}

func industryTransition(previous *marketRotationSnapshot, industry string, score float64) string {
	if previous == nil {
		return "baseline"
	}
	for _, row := range previous.Brief.PerformanceLedIndustries {
		if row.Industry == industry {
			return scoreTransition(row.Score, score)
		}
	}
	return "new"
}

func stockTransition(previous *marketRotationSnapshot, ticker string, score float64) string {
	if previous == nil {
		return "baseline"
	}
	for _, row := range previous.Brief.StockCandidates {
		if row.Ticker == ticker {
			return scoreTransition(row.Score, score)
		}
	}
	return "new"
}

func historyScore(previous *marketRotationSnapshot, ticker string) float64 {
	if previous == nil {
		return 0.5
	}
	for _, row := range previous.Brief.StockCandidates {
		if row.Ticker == ticker {
			return 0.7
		}
	}
	return 1.0
}

func scoreTransition(previous, current float64) string {
	switch {
	case current >= previous+5:
		return "strengthening"
	case current <= previous-5:
		return "weakening"
	default:
		return "unchanged"
	}
}
