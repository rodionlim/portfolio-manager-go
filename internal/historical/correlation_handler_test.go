package historical

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }

func TestBuildFlags_NormalizesIntervalFrequencyForRolling(t *testing.T) {
	opts := CorrelationOptions{
		DateMethod:        strPtr("rolling"),
		IntervalFrequency: strPtr("D"),
	}

	flags := buildFlags(opts)
	assert.Contains(t, flags, "--interval-frequency")

	// Find the value after --interval-frequency
	for i := 0; i < len(flags)-1; i++ {
		if flags[i] == "--interval-frequency" {
			assert.Equal(t, "1D", flags[i+1])
			return
		}
	}

	t.Fatalf("--interval-frequency flag not found in %v", flags)
}

func TestBuildFlags_DoesNotNormalizeIntervalFrequencyWhenNotRolling(t *testing.T) {
	opts := CorrelationOptions{
		DateMethod:        strPtr("in_sample"),
		IntervalFrequency: strPtr("D"),
	}

	flags := buildFlags(opts)
	for i := 0; i < len(flags)-1; i++ {
		if flags[i] == "--interval-frequency" {
			assert.Equal(t, "D", flags[i+1])
			return
		}
	}

	t.Fatalf("--interval-frequency flag not found in %v", flags)
}

func TestBuildFlags_PreservesNumericIntervalFrequencyForRolling(t *testing.T) {
	opts := CorrelationOptions{
		DateMethod:        strPtr("rolling"),
		IntervalFrequency: strPtr("7D"),
	}

	flags := buildFlags(opts)
	for i := 0; i < len(flags)-1; i++ {
		if flags[i] == "--interval-frequency" {
			assert.Equal(t, "7D", flags[i+1])
			return
		}
	}

	t.Fatalf("--interval-frequency flag not found in %v", flags)
}

func TestBuildFlags_StripsSignAndNormalizesForRolling(t *testing.T) {
	opts := CorrelationOptions{
		DateMethod:        strPtr("rolling"),
		IntervalFrequency: strPtr("-D"),
	}

	flags := buildFlags(opts)
	for i := 0; i < len(flags)-1; i++ {
		if flags[i] == "--interval-frequency" {
			assert.Equal(t, "1D", flags[i+1])
			return
		}
	}

	t.Fatalf("--interval-frequency flag not found in %v", flags)
}

func TestResolveCorrelationTickers_UsesRequestTickersWhenProvided(t *testing.T) {
	configs := []AssetConfig{{Ticker: "AAA", Enabled: true}, {Ticker: "BBB", Enabled: true}}
	out := resolveCorrelationTickers([]string{" C09.SI ", "D05.SI", "D05.SI", ""}, configs)
	assert.Equal(t, []string{"C09.SI", "D05.SI"}, out)
}

func TestResolveCorrelationTickers_UsesEnabledConfigsWhenEmpty(t *testing.T) {
	configs := []AssetConfig{
		{Ticker: "D05.SI", Enabled: true},
		{Ticker: " C09.SI ", Enabled: true},
		{Ticker: "ZZZ", Enabled: false},
		{Ticker: "C09.SI", Enabled: true},
	}
	out := resolveCorrelationTickers(nil, configs)
	assert.Equal(t, []string{"C09.SI", "D05.SI"}, out)
}
