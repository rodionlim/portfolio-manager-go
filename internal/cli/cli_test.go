package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
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