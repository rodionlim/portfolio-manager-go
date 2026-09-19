package sources

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func TestParseDividendsSgPrice(t *testing.T) {
	for _, tt := range []struct {
		name string
		html string
		want float64
	}{
		{
			name: "separate price and currency layout",
			html: `<div class="dividend-company-quote"><strong class="dividend-company-price">1.001</strong><span class="dividend-company-currency">SGD</span><span class="dividend-company-change is-neutral">+0.00% +0.00</span></div>`,
			want: 1.001,
		},
		{
			name: "unavailable separate price does not use change",
			html: `<div class="dividend-company-quote"><strong class="dividend-company-price">N/A</strong><span class="dividend-company-currency">SGD</span><span class="dividend-company-change is-neutral">+0.00% +0.00</span></div>`,
		},
		{
			name: "current TEMB quote layout",
			html: `<header class="dividend-company-header"><h1>TEMASEK S$500M 1.8% B 261124 <span>(TEMB)</span></h1><div class="dividend-company-quote"><span class="badge badge-secondary">SGD 1.001</span><a class="badge badge-success">+0.00% +0.00</a><small>Price updated (UTC): 2026-09-17 10:55:04</small></div></header>`,
			want: 1.001,
		},
		{
			name: "legacy currency badge",
			html: `<h4>TEMASEK (TEMB) <span class="badge">SGD 1.013</span> +0.79% +0.01</h4>`,
			want: 1.013,
		},
		{
			name: "legacy numeric badge",
			html: `<h4>TEMASEK (TEMB) SGD <span class="badge">1.012</span></h4>`,
			want: 1.012,
		},
		{
			name: "missing quote does not use unrelated badge",
			html: `<span class="badge">123</span><div class="dividend-company-quote"><span class="badge">SGD N/A</span><a>+0.00% +0.00</a></div>`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)
			price, err := parseDividendsSgPrice(doc, "TEMB")
			if tt.want == 0 {
				require.ErrorContains(t, err, "could not find price for TEMB")
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, price)
		})
	}
}
