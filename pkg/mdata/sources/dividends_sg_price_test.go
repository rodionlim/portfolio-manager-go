package sources

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParseDividendsSgPrice(t *testing.T) {
	for _, tt := range []struct {
		name, html string
		want       float64
	}{
		{"current TEMB quote", `<div class="company-quote-line"><span>SGD <strong>1.001</strong></span><span class="small">+0.00%</span><small>Updated 18 Sep 2026</small></div>`, 1.001},
		{"current equity quote", `<div class="company-quote-line"><span>SGD <strong>0.772</strong></span><span>-0.26%</span></div>`, 0.772},
		{"unavailable price", `<div class="company-quote-line"><span>SGD <strong>N/A</strong></span><span>+1.2%</span></div>`, 0},
		{"no quote", `<strong>100</strong>`, 0},
		{"nonfinite price", `<div class="company-quote-line"><span>SGD <strong>NaN</strong></span></div>`, 0},
		{"zero price", `<div class="company-quote-line"><span>SGD <strong>0</strong></span></div>`, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)
			price, err := parseDividendsSgPrice(doc, "TEMB")
			if tt.want == 0 {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, price)
		})
	}
}
