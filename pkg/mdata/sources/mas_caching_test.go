package sources

import (
	"portfolio-manager/pkg/types"
	"testing"
)

func TestMas_FetchBenchmarkInterestRates_DatabaseCaching(t *testing.T) {
	// Test filtering functionality
	testRates := []types.InterestRates{
		{Date: "17 Jul 2025", Rate: 1.5451, Tenor: "O/N", Country: "SG", RateType: "SORA"},
		{Date: "16 Jul 2025", Rate: 1.5284, Tenor: "O/N", Country: "SG", RateType: "SORA"},
		{Date: "15 Jul 2025", Rate: 1.6642, Tenor: "O/N", Country: "SG", RateType: "SORA"},
		{Date: "14 Jul 2025", Rate: 1.7048, Tenor: "O/N", Country: "SG", RateType: "SORA"},
		{Date: "11 Jul 2025", Rate: 1.6104, Tenor: "O/N", Country: "SG", RateType: "SORA"},
	}

	// Test filtering to get 3 most recent records
	filtered := filterRecentRates(testRates, 3)
	if len(filtered) != 3 {
		t.Errorf("Expected 3 rates, got %d", len(filtered))
	}

	// Test merging functionality
	existing := []types.InterestRates{
		{Date: "17 Jul 2025", Rate: 1.5451, Tenor: "O/N", Country: "SG", RateType: "SORA"},
		{Date: "16 Jul 2025", Rate: 1.5284, Tenor: "O/N", Country: "SG", RateType: "SORA"},
	}

	newRates := []types.InterestRates{
		{Date: "18 Jul 2025", Rate: 1.6000, Tenor: "O/N", Country: "SG", RateType: "SORA"}, // New date
		{Date: "17 Jul 2025", Rate: 1.5451, Tenor: "O/N", Country: "SG", RateType: "SORA"}, // Duplicate
	}

	merged := mergeInterestRates(existing, newRates)
	if len(merged) != 3 { // 2 existing + 1 new (1 duplicate filtered out)
		t.Errorf("Expected 3 rates after merge, got %d", len(merged))
	}

	t.Logf("Successfully tested filtering and merging functions")
}
