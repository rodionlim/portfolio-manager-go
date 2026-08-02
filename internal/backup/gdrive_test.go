package backup

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestIsInvalidGrant(t *testing.T) {
	assert.True(t, isInvalidGrant(&oauth2.RetrieveError{ErrorCode: "invalid_grant"}))
	assert.False(t, isInvalidGrant(&oauth2.RetrieveError{ErrorCode: "temporarily_unavailable"}))
	assert.False(t, isInvalidGrant(errors.New("network unavailable")))
}

func TestArchiveExpiredToken(t *testing.T) {
	tempDir := t.TempDir()
	tokenFile := filepath.Join(tempDir, "token.json")
	require.NoError(t, os.WriteFile(tokenFile, []byte("expired-token"), 0600))

	expiredFile, err := archiveExpiredToken(tokenFile)
	require.NoError(t, err)
	assert.Equal(t, tokenFile+".expired", expiredFile)
	assert.NoFileExists(t, tokenFile)
	assert.FileExists(t, expiredFile)

	contents, err := os.ReadFile(expiredFile)
	require.NoError(t, err)
	assert.Equal(t, "expired-token", string(contents))
}

func TestArchiveExpiredTokenPreservesPreviousArchive(t *testing.T) {
	tempDir := t.TempDir()
	tokenFile := filepath.Join(tempDir, "token.json")
	expiredFile := tokenFile + ".expired"
	require.NoError(t, os.WriteFile(tokenFile, []byte("new-expired-token"), 0600))
	require.NoError(t, os.WriteFile(expiredFile, []byte("previous-expired-token"), 0600))

	_, err := archiveExpiredToken(tokenFile)
	require.NoError(t, err)

	contents, err := os.ReadFile(expiredFile)
	require.NoError(t, err)
	assert.Equal(t, "new-expired-token", string(contents))

	previousArchives, err := filepath.Glob(expiredFile + ".*")
	require.NoError(t, err)
	require.Len(t, previousArchives, 1)
	previousContents, err := os.ReadFile(previousArchives[0])
	require.NoError(t, err)
	assert.Equal(t, "previous-expired-token", string(previousContents))
}
