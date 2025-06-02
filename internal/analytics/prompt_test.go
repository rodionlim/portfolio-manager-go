package analytics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInstitutionalFocusedPrompt(t *testing.T) {
	// Sample AI response using the new holistic prompt format
	sampleResponse := `
1. CONCISE SUMMARY:
Institutional flows show +S$40.6m weekly net buying, reversing from -S$121.3m last week, but YTD institutional flows reveal persistent DBS accumulation (+S$1.2B YTD vs +S$224.4m weekly) indicating sustained confidence. SIA shows weekly momentum (+S$50.2m) aligning with positive YTD trends (+S$380m), while retail investors contrarian selling (-S$32.5m weekly) creates opportunity. Based on flow consistency analysis and retail divergence, recommend DBS (D05) for sustained institutional backing and CapitaLand Ascendas REIT (A17U) showing consistent institutional accumulation despite retail indifference.

2. KEY INSIGHTS:
- DBS demonstrates institutional flow consistency: weekly (+S$224.4m) aligns with strong YTD accumulation (+S$1.2B)
- SIA weekly flows (+S$50.2m) match positive YTD trajectory, while retail contrarian selling creates entry opportunity
- Singtel shows institutional YTD selling (-S$890m) accelerating weekly (-S$62.5m), indicating structural concerns
- CapitaLand Ascendas REIT attracts steady institutional flows (+S$6.3m weekly, +S$156m YTD) with minimal retail attention
- Average daily institutional flows 3x higher than retail, reinforcing smart money dominance in current cycle
`

	summary, insights := parseAnalysisText(sampleResponse)

	// Test that summary includes holistic analysis
	assert.NotEmpty(t, summary, "Summary should not be empty")
	assert.Contains(t, strings.ToLower(summary), "ytd", "Summary should mention YTD trends")
	assert.Contains(t, strings.ToLower(summary), "weekly", "Summary should mention weekly patterns")
	assert.Contains(t, strings.ToLower(summary), "retail", "Summary should consider retail flows")
	assert.Contains(t, strings.ToLower(summary), "consistency", "Summary should focus on flow consistency")
	assert.Contains(t, strings.ToLower(summary), "recommend", "Summary should include recommendations")

	// Test that insights cover multiple dimensions
	assert.Greater(t, len(insights), 3, "Should have multiple insights")

	// Test that insights mention holistic factors
	holisticFactors := []string{"ytd", "weekly", "consistency", "retail", "average", "trajectory"}
	factorMentions := 0
	for _, insight := range insights {
		lowerInsight := strings.ToLower(insight)
		for _, factor := range holisticFactors {
			if strings.Contains(lowerInsight, factor) {
				factorMentions++
				break // Count each insight only once
			}
		}
	}
	assert.Greater(t, factorMentions, 2, "Insights should mention multiple holistic factors")

	// Print the results for manual inspection
	t.Logf("HOLISTIC ANALYSIS SUMMARY:\n%s\n", summary)
	t.Logf("HOLISTIC INSIGHTS (%d items):", len(insights))
	for i, insight := range insights {
		t.Logf("%d. %s", i+1, insight)
	}

	// Verify conciseness but comprehensive coverage
	summaryWords := len(strings.Fields(summary))
	assert.Less(t, summaryWords, 200, "Summary should be concise (less than 200 words)")
	assert.Greater(t, summaryWords, 50, "Summary should be comprehensive (more than 50 words)")
}
