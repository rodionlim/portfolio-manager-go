//go:build integration

package sources_test

import (
	"portfolio-manager/pkg/mdata/sources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMas_GetDividendsMetadata_Integration(t *testing.T) {
	src := sources.NewMas(nil)

	coupons, err := src.GetDividendsMetadata("BS24124Z", 0.0)
	require.NoError(t, err)
	require.NotEmpty(t, coupons)
	assert.Equal(t, 1, len(coupons))
}
