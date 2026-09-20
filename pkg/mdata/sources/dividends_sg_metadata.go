package sources

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"portfolio-manager/pkg/types"

	"github.com/PuerkitoBio/goquery"
)

var dividendsSgReference = regexp.MustCompile(`Reference:\s*([A-Z0-9]+)\b`)

// A date alone is not an event identity: distributions can contain several
// taxable, tax-exempt, or capital components, including equal-sized payments.
type dividendsSgComponent struct {
	exDate, payDate, currency, particulars string
	amount                                 float64
}

type dividendsSgCopies struct {
	references          map[string]struct{}
	unreferencedSources map[string]struct{}
	unreferencedCount   int
}

func parseDividendsSgMetadata(doc *goquery.Document, ticker string, withholdingTax float64) ([]types.DividendsMetadata, error) {
	components := make(map[dividendsSgComponent]*dividendsSgCopies)
	tables := doc.Find(".dividend-history-table")
	if tables.Length() == 0 {
		return nil, fmt.Errorf("dividend history tables not found for %s", ticker)
	}
	var parseErr error
	tables.Each(func(_ int, table *goquery.Selection) {
		if parseErr != nil {
			return
		}
		headers := []string{}
		table.Find("thead th").Each(func(_ int, cell *goquery.Selection) {
			headers = append(headers, strings.Join(strings.Fields(cell.Text()), " "))
		})
		isBond := strings.Join(headers, "|") == "Ex Date|Pay Date|Particulars / source"
		if !isBond && strings.Join(headers, "|") != "Amount|Ex Date|Pay Date|Particulars / source" {
			parseErr = fmt.Errorf("unrecognized dividend history columns for %s: %v", ticker, headers)
			return
		}
		table.Find("tbody > tr").Each(func(_ int, row *goquery.Selection) {
			cells := row.ChildrenFiltered("td")
			if cells.Length() != len(headers) {
				parseErr = fmt.Errorf("unexpected dividend row for %s: got %d columns", ticker, cells.Length())
				return
			}
			amountIdx, dateIdx := 0, 1
			if isBond {
				amountIdx, dateIdx = 2, 0
			}

			detail := cells.Last()
			action := detail.Find(`a[href*="/dividend/show"]`).First()
			particulars := action.Text()
			particulars = strings.Join(strings.Fields(particulars), " ")
			amountText := strings.TrimSpace(cells.Eq(amountIdx).Text())
			if isBond {
				amountText = particulars
			}
			currency := ""
			var amount float64
			var err error
			if strings.Contains(amountText, "%") {
				if !strings.HasPrefix(amountText, "Rate:") {
					return
				}
				rate := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(amountText, "Rate:")), "%"))
				amount, err = strconv.ParseFloat(rate, 64)
				// Some capital events are also labelled "Rate" (e.g. Astrea VI
				// at 100.5%). Do not interpret principal-sized rates as coupons.
				if amount >= 100 {
					return
				}
				// Preserve the existing semiannual bond coupon convention.
				amount /= 100 * 2
			} else {
				if amountText == "-" {
					amountText = "0"
				}
				for _, ccy := range []string{"SGD", "USD"} {
					if strings.HasPrefix(amountText, ccy) {
						currency = ccy
						amountText = strings.TrimSpace(strings.TrimPrefix(amountText, ccy))
						break
					}
				}
				amount, err = strconv.ParseFloat(amountText, 64)
			}
			if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
				parseErr = fmt.Errorf("invalid dividend amount for %s: %q", ticker, amountText)
				return
			}
			exDate := strings.TrimSpace(cells.Eq(dateIdx).Text())
			if _, err := time.Parse("2006-01-02", exDate); err != nil {
				parseErr = fmt.Errorf("invalid dividend ex-date for %s: %q", ticker, exDate)
				return
			}
			key := dividendsSgComponent{
				exDate: exDate, payDate: strings.TrimSpace(cells.Eq(dateIdx + 1).Text()),
				currency: currency, particulars: particulars, amount: amount,
			}
			copies := components[key]
			if copies == nil {
				copies = &dividendsSgCopies{references: make(map[string]struct{}), unreferencedSources: make(map[string]struct{})}
				components[key] = copies
			}
			if match := dividendsSgReference.FindStringSubmatch(detail.Text()); match != nil {
				copies.references[match[1]] = struct{}{}
			} else {
				// Identical copies of the same unreferenced source are one event, but
				// different source IDs (or unidentified rows) may be distinct events.
				source, _ := action.Attr("href")
				if source != "" {
					if _, seen := copies.unreferencedSources[source]; seen {
						return
					}
					copies.unreferencedSources[source] = struct{}{}
				}
				copies.unreferencedCount++
			}
		})
	})

	if parseErr != nil {
		return nil, parseErr
	}
	dividendMap := make(map[string]float64)
	for component, copies := range components {
		// The current page still exposes unreferenced rows alongside matching SGX references.
		// Match the two sets one-for-one only when every economic field agrees.
		// Keep distinct references, unmatched unreferenced rows, and separate components.
		count := max(len(copies.references), copies.unreferencedCount)
		dividendMap[component.exDate] += component.amount * float64(count)
	}
	var dividends []types.DividendsMetadata
	for date, amount := range dividendMap {
		dividends = append(dividends, types.DividendsMetadata{
			Ticker: ticker, ExDate: date, Amount: math.Round(amount*10000) / 10000,
			WithholdingTax: withholdingTax,
		})
	}
	sort.Slice(dividends, func(i, j int) bool { return dividends[i].ExDate < dividends[j].ExDate })
	return dividends, nil
}
