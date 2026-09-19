package sources

import (
	"fmt"
	"strings"
	"testing"

	testifydb "portfolio-manager/internal/mocks/testify/database"
	"portfolio-manager/pkg/types"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const equityDividendHeader = `<table class="table table-bordered"><tr><th>Year</th><th>Yield</th><th>Annual DPS</th><th>Amount</th><th>Ex Date</th><th>Pay Date</th><th>Particulars / source</th></tr>`

func equityDividendRow(amount, exDate, payDate, particulars, source, reference string) string {
	detail := fmt.Sprintf(`<div><a href="/dividend/show?key=%s">%s</a></div>`, source, particulars)
	if source == "" {
		detail = particulars
	}
	if reference != "" {
		detail += `<small>Reference: ` + reference + `</small>`
	}
	return fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, amount, exDate, payDate, detail)
}

func parseMetadataHTML(t *testing.T, html string) []types.DividendsMetadata {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	return parseDividendsSgMetadata(doc, "CLR", 0.15)
}

// The July 2026 CLR distribution appeared as three legacy components plus
// three matching components carrying new SGX references on September 19.
func clrDuplicateRows() []string {
	return []string{
		equityDividendRow("SGD 0.0067", "2026-07-30", "2026-08-28", "Rate: SGD 0.0067 Per Security", "2229140", "SG260722CAPD8955"),
		equityDividendRow("SGD 0.0038", "2026-07-30", "2026-08-28", "Rate: SGD 0.0038  Per Security", "2229507", "SG260722DVCA5VT0"),
		equityDividendRow("SGD 0.0125", "2026-07-30", "2026-08-28", "Rate: SGD 0.0125  Per Security", "2229508", "SG260722DVCAPBW3"),
		equityDividendRow("SGD 0.0038", "2026-07-30", "2026-08-28", "Rate: SGD 0.0038 Per Security", "2228278", ""),
		equityDividendRow("SGD 0.0125", "2026-07-30", "2026-08-28", "Rate: SGD 0.0125 Per Security", "2228279", ""),
		equityDividendRow("SGD 0.0067", "2026-07-30", "2026-08-28", "Rate: SGD 0.0067 Per Security", "2228265", ""),
	}
}

func TestParseDividendsSgMetadata_CLRDuplicateComponents(t *testing.T) {
	rows := clrDuplicateRows()
	for _, reverse := range []bool{false, true} {
		if reverse {
			for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
		// Exercise the year-summary cells present only in the first equity row.
		first := strings.Replace(rows[0], "<tr>", "<tr><td>2026</td><td>6%</td><td>SGD 0.045</td>", 1)
		result := parseMetadataHTML(t, equityDividendHeader+first+strings.Join(rows[1:], "")+"</table>")
		require.Equal(t, []types.DividendsMetadata{{Ticker: "CLR", ExDate: "2026-07-30", Amount: 0.023, WithholdingTax: 0.15}}, result)
	}
}

func TestParseDividendsSgMetadata_PreservesDistinctEvents(t *testing.T) {
	row := func(source, reference, payDate, particulars string) string {
		return equityDividendRow("SGD 0.01", "2026-07-30", payDate, particulars, source, reference)
	}
	for _, tt := range []struct {
		name string
		rows string
		want float64
	}{
		{"distinct SGX references with equal amounts", row("1", "SG1", "2026-08-28", "Rate") + row("2", "SG2", "2026-08-28", "Rate") + row("3", "", "2026-08-28", "Rate"), 0.02},
		{"different payment dates", row("1", "SG1", "2026-08-28", "Rate") + row("2", "", "2026-09-28", "Rate"), 0.02},
		{"different component descriptions", row("1", "SG1", "2026-08-28", "Ordinary") + row("2", "", "2026-08-28", "Special"), 0.02},
		{"unmatched legacy events", row("1", "SG1", "2026-08-28", "Rate") + row("2", "", "2026-08-28", "Rate") + row("3", "", "2026-08-28", "Rate"), 0.02},
		{"distinct legacy sources", row("1", "", "2026-08-28", "Rate") + row("2", "", "2026-08-28", "Rate"), 0.02},
		{"unidentified legacy events", row("", "", "2026-08-28", "Rate") + row("", "", "2026-08-28", "Rate"), 0.02},
		{"repeated SGX reference", row("1", "SG1", "2026-08-28", "Rate") + row("2", "SG1", "2026-08-28", "Rate"), 0.01},
		{"repeated legacy source", row("1", "", "2026-08-28", "Rate") + row("1", "", "2026-08-28", "Rate"), 0.01},
		{"different currencies", row("1", "SG1", "2026-08-28", "Rate") + equityDividendRow("USD 0.01", "2026-07-30", "2026-08-28", "Rate", "2", ""), 0.02},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMetadataHTML(t, equityDividendHeader+tt.rows+"</table>")
			require.Len(t, result, 1)
			require.Equal(t, tt.want, result[0].Amount)
		})
	}
}

func TestParseDividendsSgMetadata_BondSourceAnnotations(t *testing.T) {
	html := `<table class="table-bordered"><tr><th>Year</th><th>Ex Date</th><th>Pay Date</th><th>Particulars</th></tr>
	<tr><td>2025</td><td>2025-05-16</td><td>2025-05-26</td><td><div><a href="/dividend/show?key=1">Rate: 1.8%</a></div><a href="https://links.sgx.com/1">Original source</a><small>Reference: SG250429INTR3MT9</small></td></tr>
	<tr><td>2025</td><td>2025-05-16</td><td>2025-05-26</td><td><a href="/dividend/show?key=2">Rate: 1.8%</a></td></tr>
	<tr><td>2024</td><td>2024-05-15</td><td>2024-05-24</td><td>Rate: 1.8%</td></tr></table>`
	result := parseMetadataHTML(t, html)
	require.Len(t, result, 2)
	require.Equal(t, "2024-05-15", result[0].ExDate)
	require.Equal(t, 0.009, result[0].Amount)
	require.Equal(t, 0.009, result[1].Amount)
}

func TestParseDividendsSgMetadata_DoesNotCountBondPrincipalAsCoupon(t *testing.T) {
	html := `<table class="table-bordered"><tr><th>Year</th><th>Ex Date</th><th>Pay Date</th><th>Particulars</th></tr>
	<tr><td>2026</td><td>2026-03-10</td><td>2026-03-18</td><td><a href="/dividend/show?key=2214040">Rate: 3%</a><a href="https://links.sgx.com/1">Original source</a></td></tr>
	<tr><td>2026</td><td>2026-03-10</td><td>2026-03-18</td><td><a href="/dividend/show?key=2214881">Rate: 100.5%</a><a href="https://links.sgx.com/2">Original source</a></td></tr></table>`
	result := parseMetadataHTML(t, html)
	require.Len(t, result, 1)
	require.Equal(t, 0.015, result[0].Amount)
}

func TestDividendsSgRefresh_RepairsInflatedStoredTotal(t *testing.T) {
	db := new(testifydb.MockDatabase)
	src := NewDividendsSg(db)
	key := fmt.Sprintf("%s:CLR", types.DividendsKeyPrefix)
	old := types.DividendsMetadata{Ticker: "CLR", ExDate: "2019-01-30", Amount: 0.0161, Source: types.DividendSourceOfficial}
	wrong := types.DividendsMetadata{Ticker: "CLR", ExDate: "2026-07-30", Amount: 0.046, WithholdingTax: 0.15, Source: types.DividendSourceOfficial}
	correct := wrong
	correct.Amount = 0.023
	custom := types.DividendsMetadata{Ticker: "CLR", ExDate: "2026-01-28", Amount: 0, Source: types.DividendSourceCustom}
	stored := []types.DividendsMetadata{old, wrong}
	db.On("Get", key, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(1).(*[]types.DividendsMetadata) = stored
	}).Return(nil).Twice()
	db.On("Get", fmt.Sprintf("%s:CLR", types.DividendsCustomKeyPrefix), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(1).(*[]types.DividendsMetadata) = []types.DividendsMetadata{custom}
	}).Return(nil).Twice()
	db.On("Put", key, []types.DividendsMetadata{old, correct}).Run(func(args mock.Arguments) {
		stored = args.Get(1).([]types.DividendsMetadata)
	}).Return(nil).Once()
	fetched := parseMetadataHTML(t, equityDividendHeader+strings.Join(clrDuplicateRows(), "")+"</table>")
	for range 2 {
		result, err := src.upsertOfficialDividendsMetadata("CLR", fetched)
		require.NoError(t, err)
		require.Equal(t, []types.DividendsMetadata{old, custom, correct}, result)
		cached, found := src.cache.Get(key)
		require.True(t, found)
		require.Equal(t, result, cached)
	}
	db.AssertExpectations(t)
}
