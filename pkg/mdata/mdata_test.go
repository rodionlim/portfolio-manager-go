package mdata

import (
	"testing"

	"portfolio-manager/internal/config"
)

func TestMapDomicileToWitholdingTax(t *testing.T) {
	// Mock configuration
	cfg := &config.Config{
		DivWitholdingTaxSG: 0.1,
		DivWitholdingTaxUS: 0.2,
		DivWitholdingTaxHK: 0.15,
		DivWitholdingTaxIE: 0.05,
	}
	config.SetConfig(cfg)

	manager := &Manager{}

	tests := []struct {
		domicile string
		expected float64
	}{
		{"SG", 0.1},
		{"US", 0.2},
		{"HK", 0.15},
		{"IE", 0.05},
		{"UNKNOWN", 0.0},
	}

	for _, test := range tests {
		t.Run(test.domicile, func(t *testing.T) {
			result := manager.MapDomicileToWitholdingTax(test.domicile)
			if result != test.expected {
				t.Errorf("expected %f, got %f", test.expected, result)
			}
		})
	}
}
