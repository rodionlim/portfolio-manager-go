package optionpricer

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/types"
)

const (
	calendarDaysPerYear       = 365.0
	tradingDaysPerYear        = 252.0
	defaultVolatilityLookback = 60
	minReturnObservations     = 20
	blackScholesApproximation = "black_scholes"
	impliedVolMin             = 1e-6
	impliedVolInitialMax      = 5.0
	impliedVolAbsoluteTol     = 1e-8
	impliedVolMaxIterations   = 100
	impliedVolMaxExpansion    = 5
)

type PriceRequest struct {
	Ticker                 string   `json:"ticker"`
	OptionType             string   `json:"optionType"`
	Spot                   float64  `json:"spot"`
	Strike                 float64  `json:"strike"`
	Expiry                 string   `json:"expiry"`
	Rate                   *float64 `json:"rate,omitempty"`
	DividendYield          float64  `json:"dividendYield"`
	Premium                *float64 `json:"premium,omitempty"`
	Volatility             *float64 `json:"volatility,omitempty"`
	VolatilityLookbackDays int      `json:"volatilityLookbackDays,omitempty"`
}

type PriceResponse struct {
	Ticker             string   `json:"ticker"`
	Style              string   `json:"style"`
	PricingModel       string   `json:"pricingModel"`
	OptionType         string   `json:"optionType"`
	Spot               float64  `json:"spot"`
	Strike             float64  `json:"strike"`
	Expiry             string   `json:"expiry"`
	TimeToExpiryYears  float64  `json:"timeToExpiryYears"`
	Rate               float64  `json:"rate"`
	RateSource         string   `json:"rateSource,omitempty"`
	RateCurveDate      string   `json:"rateCurveDate,omitempty"`
	DividendYield      float64  `json:"dividendYield"`
	Premium            *float64 `json:"premium,omitempty"`
	Volatility         float64  `json:"volatility"`
	VolatilitySource   string   `json:"volatilitySource"`
	VolatilityLookback int      `json:"volatilityLookbackDays,omitempty"`
	NPV                float64  `json:"npv"`
	Delta              float64  `json:"delta"`
	Gamma              float64  `json:"gamma"`
	Theta              float64  `json:"theta"`
}

type Service struct {
	mdataSvc     mdata.MarketDataManager
	rateProvider rateProvider
	now          func() time.Time
}

func NewService(mdataSvc mdata.MarketDataManager) *Service {
	return &Service{
		mdataSvc:     mdataSvc,
		rateProvider: newFedTreasuryRateProvider(),
		now:          time.Now,
	}
}

func (s *Service) Price(req PriceRequest) (*PriceResponse, error) {
	ticker := strings.ToUpper(strings.TrimSpace(req.Ticker))
	if ticker == "" {
		return nil, fmt.Errorf("ticker is required")
	}

	optionType := strings.ToLower(strings.TrimSpace(req.OptionType))
	if optionType != "call" && optionType != "put" {
		return nil, fmt.Errorf("optionType must be either call or put")
	}

	if req.Spot <= 0 {
		return nil, fmt.Errorf("spot must be greater than 0")
	}

	if req.Strike <= 0 {
		return nil, fmt.Errorf("strike must be greater than 0")
	}

	if req.Premium != nil && *req.Premium <= 0 {
		return nil, fmt.Errorf("premium must be greater than 0")
	}

	if req.Rate != nil && *req.Rate < 0 {
		return nil, fmt.Errorf("rate must be greater than or equal to 0")
	}

	volatilityLookback, err := normalizeVolatilityLookbackDays(req.VolatilityLookbackDays)
	if err != nil {
		return nil, err
	}

	expiry, err := parseExpiry(req.Expiry)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	if !expiry.After(now) {
		return nil, fmt.Errorf("expiry must be in the future")
	}

	timeToExpiryYears := expiry.Sub(now).Hours() / 24 / calendarDaysPerYear
	if timeToExpiryYears <= 0 {
		return nil, fmt.Errorf("time to expiry must be greater than 0")
	}

	resolvedRiskFreeRate := resolvedRate{}
	if req.Rate != nil {
		resolvedRiskFreeRate = resolvedRate{
			Rate:   *req.Rate,
			Source: "input",
		}
	} else {
		resolvedRiskFreeRate = s.rateProvider.Resolve(timeToExpiryYears)
	}

	volatilitySource := "input"
	responseLookback := 0
	volatility := 0.0
	if req.Volatility != nil {
		volatility = *req.Volatility
	} else if req.Premium != nil {
		volatility, err = implyVolatility(
			optionType,
			req.Spot,
			req.Strike,
			timeToExpiryYears,
			resolvedRiskFreeRate.Rate,
			req.DividendYield,
			*req.Premium,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to imply volatility from premium: %w", err)
		}
		volatilitySource = "implied_from_premium"
	} else {
		volatility, err = s.estimateAnnualizedVolatility(ticker, now, volatilityLookback)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate volatility: %w", err)
		}
		volatilitySource = "estimated_historical"
		responseLookback = volatilityLookback
	}

	if volatility <= 0 {
		return nil, fmt.Errorf("volatility must be greater than 0")
	}

	npv, delta, gamma, theta, err := calculateBlackScholes(
		optionType,
		req.Spot,
		req.Strike,
		timeToExpiryYears,
		resolvedRiskFreeRate.Rate,
		req.DividendYield,
		volatility,
	)
	if err != nil {
		return nil, err
	}

	return &PriceResponse{
		Ticker:             ticker,
		Style:              "american",
		PricingModel:       blackScholesApproximation,
		OptionType:         optionType,
		Spot:               req.Spot,
		Strike:             req.Strike,
		Expiry:             expiry.Format("2006-01-02"),
		TimeToExpiryYears:  timeToExpiryYears,
		Rate:               resolvedRiskFreeRate.Rate,
		RateSource:         resolvedRiskFreeRate.Source,
		RateCurveDate:      resolvedRiskFreeRate.CurveDate,
		DividendYield:      req.DividendYield,
		Premium:            req.Premium,
		Volatility:         volatility,
		VolatilitySource:   volatilitySource,
		VolatilityLookback: responseLookback,
		NPV:                npv,
		Delta:              delta,
		Gamma:              gamma,
		Theta:              theta,
	}, nil
}

func parseExpiry(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("expiry is required")
	}

	expiry, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("expiry must be in YYYY-MM-DD format")
	}

	return time.Date(expiry.Year(), expiry.Month(), expiry.Day(), 23, 59, 59, 0, time.UTC), nil
}

func normalizeVolatilityLookbackDays(lookbackDays int) (int, error) {
	switch lookbackDays {
	case 0:
		return defaultVolatilityLookback, nil
	case 30, 60, 180, 360:
		return lookbackDays, nil
	default:
		return 0, fmt.Errorf("volatilityLookbackDays must be one of 30, 60, 180, or 360")
	}
}

func (s *Service) estimateAnnualizedVolatility(ticker string, now time.Time, lookbackDays int) (float64, error) {
	if s.mdataSvc == nil {
		return 0, fmt.Errorf("market data service is not configured")
	}

	from := now.AddDate(0, 0, -lookbackDays).Unix()
	to := now.Unix()
	history, _, err := s.mdataSvc.GetHistoricalData(ticker, from, to)
	if err != nil {
		return 0, err
	}

	if len(history) < minReturnObservations+1 {
		return 0, fmt.Errorf("not enough historical data points to estimate volatility")
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp < history[j].Timestamp
	})

	prices := make([]float64, 0, len(history))
	for _, point := range history {
		price := pickPrice(point)
		if price > 0 {
			prices = append(prices, price)
		}
	}

	returns := make([]float64, 0, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		if prices[i-1] <= 0 || prices[i] <= 0 {
			continue
		}
		returns = append(returns, math.Log(prices[i]/prices[i-1]))
	}

	if len(returns) < minReturnObservations {
		return 0, fmt.Errorf("not enough valid return observations to estimate volatility")
	}

	stdDev := sampleStdDev(returns)
	if stdDev <= 0 {
		return 0, fmt.Errorf("estimated volatility is not positive")
	}

	return stdDev * math.Sqrt(tradingDaysPerYear), nil
}

func implyVolatility(optionType string, spot, strike, timeToExpiryYears, rate, dividendYield, premium float64) (float64, error) {
	if premium <= 0 {
		return 0, fmt.Errorf("premium must be greater than 0")
	}

	lowerBound, upperBound := optionPriceBounds(optionType, spot, strike, timeToExpiryYears, rate, dividendYield)
	if premium < lowerBound-impliedVolAbsoluteTol {
		return 0, fmt.Errorf("premium is below theoretical lower bound %.8f", lowerBound)
	}
	if premium > upperBound+impliedVolAbsoluteTol {
		return 0, fmt.Errorf("premium is above theoretical upper bound %.8f", upperBound)
	}

	low := impliedVolMin
	high := impliedVolInitialMax
	lowPrice, _, _, _, err := calculateBlackScholes(optionType, spot, strike, timeToExpiryYears, rate, dividendYield, low)
	if err != nil {
		return 0, err
	}
	if math.Abs(lowPrice-premium) <= impliedVolAbsoluteTol {
		return low, nil
	}

	highPrice, _, _, _, err := calculateBlackScholes(optionType, spot, strike, timeToExpiryYears, rate, dividendYield, high)
	if err != nil {
		return 0, err
	}
	for expansion := 0; expansion < impliedVolMaxExpansion && highPrice < premium; expansion++ {
		high *= 2
		highPrice, _, _, _, err = calculateBlackScholes(optionType, spot, strike, timeToExpiryYears, rate, dividendYield, high)
		if err != nil {
			return 0, err
		}
	}
	if highPrice < premium {
		return 0, fmt.Errorf("premium is too large to imply a volatility in the solver range")
	}

	for i := 0; i < impliedVolMaxIterations; i++ {
		mid := 0.5 * (low + high)
		price, _, _, _, err := calculateBlackScholes(optionType, spot, strike, timeToExpiryYears, rate, dividendYield, mid)
		if err != nil {
			return 0, err
		}

		diff := price - premium
		if math.Abs(diff) <= impliedVolAbsoluteTol {
			return mid, nil
		}

		if diff > 0 {
			high = mid
		} else {
			low = mid
		}
	}

	return 0.5 * (low + high), nil
}

func optionPriceBounds(optionType string, spot, strike, timeToExpiryYears, rate, dividendYield float64) (float64, float64) {
	discountedSpot := spot * math.Exp(-dividendYield*timeToExpiryYears)
	discountedStrike := strike * math.Exp(-rate*timeToExpiryYears)

	if optionType == "call" {
		return math.Max(discountedSpot-discountedStrike, 0), discountedSpot
	}

	return math.Max(discountedStrike-discountedSpot, 0), discountedStrike
}

func pickPrice(point *types.AssetData) float64 {
	if point == nil {
		return 0
	}

	if point.AdjClose > 0 {
		return point.AdjClose
	}

	return point.Price
}

func sampleStdDev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}

	variance /= float64(len(values) - 1)
	if variance <= 0 {
		return 0
	}

	return math.Sqrt(variance)
}

func calculateBlackScholes(optionType string, spot, strike, timeToExpiryYears, rate, dividendYield, volatility float64) (float64, float64, float64, float64, error) {
	if timeToExpiryYears <= 0 {
		return 0, 0, 0, 0, fmt.Errorf("time to expiry must be greater than 0")
	}

	if volatility <= 0 {
		return 0, 0, 0, 0, fmt.Errorf("volatility must be greater than 0")
	}

	sqrtT := math.Sqrt(timeToExpiryYears)
	d1 := (math.Log(spot/strike) + (rate-dividendYield+0.5*volatility*volatility)*timeToExpiryYears) / (volatility * sqrtT)
	d2 := d1 - volatility*sqrtT

	discountedSpot := spot * math.Exp(-dividendYield*timeToExpiryYears)
	discountedStrike := strike * math.Exp(-rate*timeToExpiryYears)
	normD1 := normCDF(d1)
	normD2 := normCDF(d2)
	pdfD1 := normPDF(d1)

	gamma := math.Exp(-dividendYield*timeToExpiryYears) * pdfD1 / (spot * volatility * sqrtT)

	if optionType == "call" {
		npv := discountedSpot*normD1 - discountedStrike*normD2
		delta := math.Exp(-dividendYield*timeToExpiryYears) * normD1
		thetaAnnual := -(discountedSpot*pdfD1*volatility)/(2*sqrtT) - rate*discountedStrike*normD2 + dividendYield*discountedSpot*normD1
		return npv, delta, gamma, thetaAnnual / calendarDaysPerYear, nil
	}

	npv := discountedStrike*normCDF(-d2) - discountedSpot*normCDF(-d1)
	delta := math.Exp(-dividendYield*timeToExpiryYears) * (normD1 - 1)
	thetaAnnual := -(discountedSpot*pdfD1*volatility)/(2*sqrtT) + rate*discountedStrike*normCDF(-d2) - dividendYield*discountedSpot*normCDF(-d1)
	return npv, delta, gamma, thetaAnnual / calendarDaysPerYear, nil
}

func normCDF(value float64) float64 {
	return 0.5 * (1 + math.Erf(value/math.Sqrt2))
}

func normPDF(value float64) float64 {
	return math.Exp(-0.5*value*value) / math.Sqrt(2*math.Pi)
}
