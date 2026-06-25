package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"portfolio-manager/pkg/rdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	cli := NewCLI()

	t.Run("Version command", func(t *testing.T) {
		// This test simply checks that handleVersion doesn't panic and works correctly
		// when the VERSION file exists. Since getVersion now handles multiple paths,
		// it should find the VERSION file in the repository root.
		err := cli.handleVersion([]string{})
		assert.NoError(t, err)
	})

	t.Run("ParseAndExecute with version", func(t *testing.T) {
		err := cli.ParseAndExecute([]string{"program", "-v"})
		assert.NoError(t, err)

		err = cli.ParseAndExecute([]string{"program", "--version"})
		assert.NoError(t, err)
	})

	t.Run("ParseAndExecute with no command", func(t *testing.T) {
		err := cli.ParseAndExecute([]string{"program"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no command specified")
	})

	t.Run("ParseAndExecute with unknown command", func(t *testing.T) {
		err := cli.ParseAndExecute([]string{"program", "unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := formatFileSize(test.bytes)
		assert.Equal(t, test.expected, result)
	}
}

func TestGetVersion(t *testing.T) {
	// This test reads the actual VERSION file
	version, err := getVersion()
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
}

func TestReferenceDataCLICommands(t *testing.T) {
	t.Run("list reference data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/api/v1/refdata", r.URL.Path)
			json.NewEncoder(w).Encode(map[string]rdata.TickerReferenceWithSGXMapped{
				"AAPL": {
					TickerReference: rdata.TickerReference{
						ID:            "AAPL",
						Name:          "Apple",
						AssetClass:    rdata.AssetClassEquities,
						AssetSubClass: rdata.AssetSubClassStock,
						Category:      rdata.CategoryTechnology,
						Ccy:           "USD",
						Domicile:      "US",
					},
				},
			})
		}))
		defer server.Close()

		cli := NewCLI()
		cli.SetBaseURL(server.URL)

		require.NoError(t, cli.handleReferenceDataList([]string{"--asset-class", rdata.AssetClassEquities}))
	})

	t.Run("get reference data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(map[string]rdata.TickerReferenceWithSGXMapped{
				"AAPL": {
					TickerReference: rdata.TickerReference{
						ID:   "AAPL",
						Name: "Apple",
					},
				},
			})
		}))
		defer server.Close()

		cli := NewCLI()
		cli.SetBaseURL(server.URL)

		require.NoError(t, cli.handleReferenceDataGet([]string{"AAPL"}))
	})

	t.Run("add reference data", func(t *testing.T) {
		var payload rdata.TickerReference
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			json.NewEncoder(w).Encode(payload.ID)
		}))
		defer server.Close()

		cli := NewCLI()
		cli.SetBaseURL(server.URL)

		require.NoError(t, cli.handleReferenceDataAdd([]string{
			"--id", "aapl",
			"--name", "Apple",
			"--underlying-ticker", "aapl",
			"--asset-class", rdata.AssetClassEquities,
			"--asset-sub-class", rdata.AssetSubClassStock,
			"--category", rdata.CategoryTechnology,
			"--ccy", "usd",
			"--domicile", "us",
		}))
		require.Equal(t, "AAPL", payload.ID)
		require.Equal(t, "AAPL", payload.UnderlyingTicker)
		require.Equal(t, "USD", payload.Ccy)
		require.Equal(t, "US", payload.Domicile)
	})

	t.Run("update reference data from file", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "refdata.yaml")
		require.NoError(t, os.WriteFile(filePath, []byte(`
id: msft
name: Microsoft
underlying_ticker: msft
asset_class: eq
asset_sub_class: stock
category: technology
ccy: usd
domicile: us
`), 0o600))

		var payload rdata.TickerReference
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			json.NewEncoder(w).Encode(payload)
		}))
		defer server.Close()

		cli := NewCLI()
		cli.SetBaseURL(server.URL)

		require.NoError(t, cli.handleReferenceDataUpdate([]string{"--file", filePath}))
		require.Equal(t, "MSFT", payload.ID)
		require.Equal(t, "MSFT", payload.UnderlyingTicker)
		require.Equal(t, "USD", payload.Ccy)
	})

	t.Run("delete reference data", func(t *testing.T) {
		var payload []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			json.NewEncoder(w).Encode(map[string]string{"message": "deleted"})
		}))
		defer server.Close()

		cli := NewCLI()
		cli.SetBaseURL(server.URL)

		require.NoError(t, cli.handleReferenceDataDelete([]string{"--id", "aapl,msft", "--yes"}))
		require.Equal(t, []string{"AAPL", "MSFT"}, payload)
	})
}
