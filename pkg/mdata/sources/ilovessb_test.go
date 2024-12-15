//go:build integration

package sources_test

import (
	"portfolio-manager/pkg/mdata/sources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestILoveSsb_GetDividendsMetadata_Integration(t *testing.T) {
	src := sources.NewILoveSsb(nil)

	coupons, err := src.GetDividendsMetadata("SBMAR24")
	require.NoError(t, err)
	require.NotEmpty(t, coupons)
	assert.Equal(t, 20, len(coupons))
}
