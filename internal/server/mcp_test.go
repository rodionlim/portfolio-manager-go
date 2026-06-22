package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMCPInstructionsDisambiguateStockIndustriesAndSectorETFs(t *testing.T) {
	require.Contains(t, mcpInstructions, "Stock industries")
	require.Contains(t, mcpInstructions, "sector ETF fund-flow tool")
	require.Contains(t, mcpInstructions, "ask whether they mean stock-market industries or sector ETFs")
	require.Contains(t, mcpInstructions, "1M, 3M, 1Y, 3Y, and YTD")
}
