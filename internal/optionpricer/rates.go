package optionpricer

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"portfolio-manager/pkg/common"
)

const (
	defaultRiskFreeRate            = 0.0375
	fedTreasuryCurveCacheTTL       = 12 * time.Hour
	fedTreasuryConstantMaturityURL = "https://www.federalreserve.gov/datadownload/Output.aspx?rel=H15&series=bf17364827e38702b42a58cf8eaa3f78&lastobs=&from=&to=&filetype=csv&label=include&layout=seriescolumn&type=package"
)

type resolvedRate struct {
	Rate      float64
	Source    string
	CurveDate string
}

type rateProvider interface {
	Resolve(timeToExpiryYears float64) resolvedRate
}

type treasuryCurvePoint struct {
	MaturityYears float64
	Rate          float64
}

type treasuryCurveSnapshot struct {
	CurveDate string
	FetchedAt time.Time
	Points    []treasuryCurvePoint
}

type fedTreasuryRateProvider struct {
	client *http.Client
	url    string
	now    func() time.Time

	mutex sync.RWMutex
	cache *treasuryCurveSnapshot
}

func newFedTreasuryRateProvider() *fedTreasuryRateProvider {
	return &fedTreasuryRateProvider{
		client: &http.Client{Timeout: 15 * time.Second},
		url:    fedTreasuryConstantMaturityURL,
		now:    time.Now,
	}
}

func (p *fedTreasuryRateProvider) Resolve(timeToExpiryYears float64) resolvedRate {
	curve, err := p.loadCurve()
	if err != nil {
		return resolvedRate{
			Rate:   defaultRiskFreeRate,
			Source: "fallback_default",
		}
	}

	rate, err := interpolateTreasuryCurve(curve.Points, timeToExpiryYears)
	if err != nil {
		return resolvedRate{
			Rate:   defaultRiskFreeRate,
			Source: "fallback_default",
		}
	}

	return resolvedRate{
		Rate:      rate,
		Source:    "fed_h15_treasury_constant_maturity",
		CurveDate: curve.CurveDate,
	}
}

func (p *fedTreasuryRateProvider) loadCurve() (*treasuryCurveSnapshot, error) {
	p.mutex.RLock()
	if p.cache != nil && p.now().Sub(p.cache.FetchedAt) < fedTreasuryCurveCacheTTL {
		cached := p.cache
		p.mutex.RUnlock()
		return cached, nil
	}
	p.mutex.RUnlock()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cache != nil && p.now().Sub(p.cache.FetchedAt) < fedTreasuryCurveCacheTTL {
		return p.cache, nil
	}

	req, err := common.NewHttpRequestWithUserAgent(http.MethodGet, p.url)
	if err != nil {
		return nil, err
	}

	resp, err := common.DoWithRetry(p.client, req, 2, 1500*time.Millisecond, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	curveDate, points, err := parseTreasuryCurveCSV(resp.Body)
	if err != nil {
		return nil, err
	}

	p.cache = &treasuryCurveSnapshot{
		CurveDate: curveDate,
		FetchedAt: p.now(),
		Points:    points,
	}

	return p.cache, nil
}

func parseTreasuryCurveCSV(reader io.Reader) (string, []treasuryCurvePoint, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1

	records, err := csvReader.ReadAll()
	if err != nil {
		return "", nil, fmt.Errorf("failed to read Treasury curve CSV: %w", err)
	}

	if len(records) < 7 {
		return "", nil, fmt.Errorf("Treasury curve CSV did not contain enough rows")
	}

	maturities, err := parseTreasuryCurveMaturities(records[0])
	if err != nil {
		return "", nil, err
	}

	latestDate := ""
	latestPoints := []treasuryCurvePoint{}
	for _, record := range records[6:] {
		if len(record) == 0 {
			continue
		}

		if _, err := time.Parse("2006-01-02", strings.TrimSpace(record[0])); err != nil {
			continue
		}

		points := make([]treasuryCurvePoint, 0, len(maturities))
		for idx, maturity := range maturities {
			column := idx + 1
			if column >= len(record) {
				continue
			}

			rate, ok := parseTreasuryRateValue(record[column])
			if !ok {
				continue
			}

			points = append(points, treasuryCurvePoint{
				MaturityYears: maturity,
				Rate:          rate / 100.0,
			})
		}

		if len(points) >= 2 {
			latestDate = strings.TrimSpace(record[0])
			latestPoints = points
		}
	}

	if latestDate == "" || len(latestPoints) < 2 {
		return "", nil, fmt.Errorf("Treasury curve CSV did not contain a usable curve")
	}

	sort.Slice(latestPoints, func(i, j int) bool {
		return latestPoints[i].MaturityYears < latestPoints[j].MaturityYears
	})

	return latestDate, latestPoints, nil
}

func parseTreasuryCurveMaturities(header []string) ([]float64, error) {
	if len(header) < 2 {
		return nil, fmt.Errorf("Treasury curve CSV header is incomplete")
	}

	maturities := make([]float64, 0, len(header)-1)
	for _, column := range header[1:] {
		maturity, err := maturityYearsFromDescription(column)
		if err != nil {
			return nil, err
		}
		maturities = append(maturities, maturity)
	}

	return maturities, nil
}

func maturityYearsFromDescription(description string) (float64, error) {
	trimmed := strings.TrimSpace(strings.Trim(description, `"`))
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return 0, fmt.Errorf("unexpected Treasury curve description: %s", description)
	}

	for _, part := range parts {
		lower := strings.ToLower(strings.Trim(part, `",.`))
		segments := strings.SplitN(lower, "-", 2)
		if len(segments) != 2 {
			continue
		}

		value, err := strconv.ParseFloat(segments[0], 64)
		if err != nil {
			continue
		}

		if strings.HasPrefix(segments[1], "month") {
			return value / 12.0, nil
		}
		if strings.HasPrefix(segments[1], "year") {
			return value, nil
		}
	}

	return 0, fmt.Errorf("unexpected Treasury maturity unit in description: %s", description)
}

func parseTreasuryRateValue(value string) (float64, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, false
	}

	lower := strings.ToLower(trimmed)
	if lower == "nd" || lower == "n.d." || lower == "n.a." || lower == "na" {
		return 0, false
	}

	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func interpolateTreasuryCurve(points []treasuryCurvePoint, maturityYears float64) (float64, error) {
	if len(points) == 0 {
		return 0, fmt.Errorf("Treasury curve is empty")
	}

	if maturityYears <= points[0].MaturityYears {
		return points[0].Rate, nil
	}

	last := points[len(points)-1]
	if maturityYears >= last.MaturityYears {
		return last.Rate, nil
	}

	for idx := 1; idx < len(points); idx++ {
		left := points[idx-1]
		right := points[idx]
		if maturityYears > right.MaturityYears {
			continue
		}

		span := right.MaturityYears - left.MaturityYears
		if span <= 0 {
			return left.Rate, nil
		}

		weight := (maturityYears - left.MaturityYears) / span
		return left.Rate + weight*(right.Rate-left.Rate), nil
	}

	return last.Rate, nil
}
