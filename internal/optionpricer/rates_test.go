package optionpricer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFedTreasuryRateProviderInterpolatesLatestCurve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`"Series Description","Market yield on U.S. Treasury securities at 1-month  constant maturity, quoted on investment basis","Market yield on U.S. Treasury securities at 3-month  constant maturity, quoted on investment basis","Market yield on U.S. Treasury securities at 6-month  constant maturity, quoted on investment basis","Market yield on U.S. Treasury securities at 1-year  constant maturity, quoted on investment basis","Market yield on U.S. Treasury securities at 2-year  constant maturity, quoted on investment basis","Market yield on U.S. Treasury securities at 3-year  constant maturity, quoted on investment basis","Market yield on U.S. Treasury securities at 5-year  constant maturity, quoted on investment basis"
"Unit:","Percent:_Per_Year","Percent:_Per_Year","Percent:_Per_Year","Percent:_Per_Year","Percent:_Per_Year","Percent:_Per_Year","Percent:_Per_Year"
"Multiplier:","1","1","1","1","1","1","1"
"Currency:","NA","NA","NA","NA","NA","NA","NA"
"Unique Identifier:","H15/H15/RIFLGFCM01_N.B","H15/H15/RIFLGFCM03_N.B","H15/H15/RIFLGFCM06_N.B","H15/H15/RIFLGFCY01_N.B","H15/H15/RIFLGFCY02_N.B","H15/H15/RIFLGFCY03_N.B","H15/H15/RIFLGFCY05_N.B"
"Time Period","RIFLGFCM01_N.B","RIFLGFCM03_N.B","RIFLGFCM06_N.B","RIFLGFCY01_N.B","RIFLGFCY02_N.B","RIFLGFCY03_N.B","RIFLGFCY05_N.B"
2026-04-03,3.71,3.71,3.73,3.72,3.84,3.88,3.99
2026-04-06,3.72,3.72,3.74,3.72,3.84,3.88,3.98
`))
		require.NoError(t, err)
	}))
	defer server.Close()

	provider := &fedTreasuryRateProvider{
		client: server.Client(),
		url:    server.URL,
		now: func() time.Time {
			return time.Date(2026, time.April, 7, 18, 0, 0, 0, time.UTC)
		},
	}

	result := provider.Resolve(4.0)

	assert.Equal(t, "fed_h15_treasury_constant_maturity", result.Source)
	assert.Equal(t, "2026-04-06", result.CurveDate)
	assert.InDelta(t, 0.0393, result.Rate, 1e-9)
}

func TestFedTreasuryRateProviderFallsBackOnFetchFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := &fedTreasuryRateProvider{
		client: server.Client(),
		url:    server.URL,
		now:    time.Now,
	}

	result := provider.Resolve(1.5)

	assert.Equal(t, "fallback_default", result.Source)
	assert.Equal(t, "", result.CurveDate)
	assert.Equal(t, defaultRiskFreeRate, result.Rate)
}
