package sources

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

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
	references    map[string]struct{}
	legacySources map[string]struct{}
	legacyCount   int
}

func parseDividendsSgMetadata(doc *goquery.Document, ticker string, withholdingTax float64) []types.DividendsMetadata {
	components := make(map[dividendsSgComponent]*dividendsSgCopies)
	doc.Find("table.table-bordered").Each(func(_ int, table *goquery.Selection) {
		isBond := table.Find("tr").First().Find("th").Length() == 4
		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() < 4 {
				return
			}
			// Equity rows omit the three year-summary cells after the first row.
			amountIdx, dateIdx := 3, 4
			if cells.Length() == 4 {
				amountIdx, dateIdx = 0, 1
			}
			if isBond {
				amountIdx, dateIdx = 3, 1
			}
			detail := cells.Last()
			action := detail.Find(`a[href*="/dividend/show"]`).First()
			particulars := action.Text()
			if action.Length() == 0 {
				// Legacy pages have plain particulars. Exclude source annotations.
				clean := detail.Clone()
				clean.Find(`small, a[href*="links.sgx.com"]`).Remove()
				particulars = clean.Text()
			}
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
				return
			}
			exDate := strings.TrimSpace(cells.Eq(dateIdx).Text())
			if exDate == "" || exDate == "-" {
				return
			}
			key := dividendsSgComponent{
				exDate: exDate, payDate: strings.TrimSpace(cells.Eq(dateIdx + 1).Text()),
				currency: currency, particulars: particulars, amount: amount,
			}
			copies := components[key]
			if copies == nil {
				copies = &dividendsSgCopies{references: make(map[string]struct{}), legacySources: make(map[string]struct{})}
				components[key] = copies
			}
			if match := dividendsSgReference.FindStringSubmatch(detail.Text()); match != nil {
				copies.references[match[1]] = struct{}{}
			} else {
				// Identical copies of the same legacy source are one event, but
				// different source IDs (or unidentified rows) may be distinct events.
				source, _ := action.Attr("href")
				if source != "" {
					if _, seen := copies.legacySources[source]; seen {
						return
					}
					copies.legacySources[source] = struct{}{}
				}
				copies.legacyCount++
			}
		})
	})

	dividendMap := make(map[string]float64)
	for component, copies := range components {
		// Upstream exposes old rows alongside refreshed rows with SGX references.
		// Match the two sets one-for-one only when every economic field agrees.
		// Keep distinct references, unmatched legacy rows, and separate components.
		count := max(len(copies.references), copies.legacyCount)
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
	return dividends
}
