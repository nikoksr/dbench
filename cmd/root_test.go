package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/mocks"
)

// Test for determineDefaultDataPath
func TestDetermineDefaultDataPath(t *testing.T) {
	t.Parallel()

	appName := "testapp"

	// Test cases
	tests := []struct {
		name            string
		envDataDir      string
		homeDir         string
		homeDirErr      error
		expectedDataDir string
		expectErr       bool
	}{
		{
			name:            "env set",
			envDataDir:      "/custom/dir",
			expectedDataDir: "/custom/dir",
			expectErr:       false,
		},
		{
			name:            "standard path",
			homeDir:         "/home/user",
			expectedDataDir: "/home/user/.local/share/testapp",
			expectErr:       false,
		},
		{
			name:       "home dir error",
			homeDirErr: fmt.Errorf("error"),
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockFS := new(mocks.MockFileSystem)
			mockEnv := new(mocks.MockEnvironment)

			mockEnv.On("Getenv", envDataDir).Return(tc.envDataDir)
			if tc.envDataDir == "" {
				mockFS.On("UserHomeDir").Return(tc.homeDir, tc.homeDirErr).Once()
			}

			dataDir, err := determineDefaultDataPath(appName, mockEnv, mockFS)

			if (err != nil) != tc.expectErr {
				t.Errorf("determineDefaultDataPath() error = %v, expectErr %v", err, tc.expectErr)
				return
			}
			if err == nil && dataDir != tc.expectedDataDir {
				t.Errorf("determineDefaultDataPath() = %v, want %v", dataDir, tc.expectedDataDir)
			}

			mockEnv.AssertExpectations(t)
			mockFS.AssertExpectations(t)
		})
	}
}

func TestBuildDSN(t *testing.T) {
	t.Parallel()

	// Test cases
	tests := []struct {
		name     string
		dataDir  string
		expected string
	}{
		{
			name:     "valid path",
			dataDir:  "/path/to/data",
			expected: "file:/path/to/data/" + build.AppName + ".db?cache=shared&_fk=1",
		},
		{
			name:     "empty path",
			dataDir:  "",
			expected: "file:" + build.AppName + ".db?cache=shared&_fk=1",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dsn := buildDSN(tc.dataDir)
			assert.Equal(t, tc.expected, dsn)
		})
	}
}
