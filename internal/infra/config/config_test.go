package config

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"testing"
	"testing/fstest"

	oglconfig "github.com/piprim/mmw/platform/config"
	oglslog "github.com/piprim/mmw/platform/slog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDefaultTOML is a minimal default config used across Load tests.
// Database port is an integer to match oglconfig.Database.Port (Port int16).
const testDefaultTOML = `
[database]
user = "rcv"
host = "localhost"
port = 5432
name = "testdb"
`

func TestLoad_Success(t *testing.T) {
	origFS := getConfigFS
	defer func() { getConfigFS = origFS }()

	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": {Data: []byte(testDefaultTOML)},
			"configs/testing.toml": {Data: []byte(`
[database]
host = "test-host"
`)},
		}
	}

	t.Setenv("APP_ENV", "testing")
	t.Setenv("DB_PASSWORD", "secret123")

	ctx := context.Background()
	config, err := Load(ctx, "")

	require.NoError(t, err)
	require.NotNil(t, config)
	require.NotNil(t, config.Database)

	assert.Equal(t, oglconfig.EnvironmentTesting, config.Environment)
	assert.Equal(t, "rcv", config.Database.User)
	assert.Equal(t, "test-host", config.Database.Host)

	url := config.Database.URL()
	assert.Contains(t, url, "secret123")
	assert.Contains(t, url, "test-host")
}

func TestLoad_MissingPasswordEnv(t *testing.T) {
	origFS := getConfigFS
	defer func() { getConfigFS = origFS }()

	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": {Data: []byte(testDefaultTOML)},
		}
	}

	t.Setenv("APP_ENV", "development")
	t.Setenv("DB_PASSWORD", "")

	ctx := context.Background()
	config, err := Load(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "database password is empty")
}

func TestLoad_WithDefaultConfigOnly(t *testing.T) {
	origFS := getConfigFS
	defer func() { getConfigFS = origFS }()

	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": {Data: []byte(testDefaultTOML)},
		}
	}

	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_PASSWORD", "password")

	ctx := context.Background()
	config, err := Load(ctx, "")

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, oglconfig.EnvironmentProduction, config.Environment)
}

func TestConfig_GetAppEnv(t *testing.T) {
	config := &Config{
		Base: oglconfig.Base{Environment: oglconfig.EnvironmentStaging},
	}

	env := config.GetAppEnv()
	assert.NotNil(t, env)
	assert.Equal(t, "staging", env.String())
}

func TestLoad_WithDebugLevel(t *testing.T) {
	origFS := getConfigFS
	defer func() { getConfigFS = origFS }()

	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": {Data: []byte(`
log-level = "debug"

[database]
user = "rcv"
host = "localhost"
port = 5432
name = "testdb"
`)},
		}
	}

	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_PASSWORD", "secret123")

	ctx := context.Background()
	config, err := Load(ctx, "")

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, oglslog.LogLevel("debug"), config.LogLevel)
	assert.Equal(t, slog.LevelDebug, config.LogLevel.SlogLevel())
}

func TestLoad_WithDifferentDebugLevels(t *testing.T) {
	tests := []struct {
		name          string
		levelString   string
		expectedLevel slog.Level
	}{
		{
			name:          "debug level",
			levelString:   "debug",
			expectedLevel: slog.LevelDebug,
		},
		{
			name:          "info level",
			levelString:   "info",
			expectedLevel: slog.LevelInfo,
		},
		{
			name:          "warn level",
			levelString:   "warn",
			expectedLevel: slog.LevelWarn,
		},
		{
			name:          "error level",
			levelString:   "error",
			expectedLevel: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origFS := getConfigFS
			defer func() { getConfigFS = origFS }()

			getConfigFS = func() fs.FS {
				return fstest.MapFS{
					"configs/default.toml": {Data: fmt.Appendf(nil, `
log-level = "%s"

[database]
user = "rcv"
host = "localhost"
port = 5432
name = "testdb"
`, tt.levelString)},
				}
			}

			t.Setenv("APP_ENV", "development")
			t.Setenv("DB_PASSWORD", "secret123")

			ctx := context.Background()
			config, err := Load(ctx, "")

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.Equal(t, oglslog.LogLevel(tt.levelString), config.LogLevel)
			assert.Equal(t, tt.expectedLevel, config.LogLevel.SlogLevel())
		})
	}
}

func TestLoad_AllEnvironments(t *testing.T) {
	tests := []struct {
		name        string
		appEnv      string
		expectedEnv oglconfig.Environment
	}{
		{
			name:        "development environment",
			appEnv:      "development",
			expectedEnv: oglconfig.EnvironmentDevelopment,
		},
		{
			name:        "staging environment",
			appEnv:      "staging",
			expectedEnv: oglconfig.EnvironmentStaging,
		},
		{
			name:        "production environment",
			appEnv:      "production",
			expectedEnv: oglconfig.EnvironmentProduction,
		},
		{
			name:        "testing environment",
			appEnv:      "testing",
			expectedEnv: oglconfig.EnvironmentTesting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origFS := getConfigFS
			defer func() { getConfigFS = origFS }()

			getConfigFS = func() fs.FS {
				return fstest.MapFS{
					"configs/default.toml": {Data: []byte(testDefaultTOML)},
				}
			}

			t.Setenv("APP_ENV", tt.appEnv)
			t.Setenv("DB_PASSWORD", "secret123")

			ctx := context.Background()
			config, err := Load(ctx, "")

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.Equal(t, tt.expectedEnv, config.Environment)
			assert.Equal(t, tt.appEnv, config.Environment.String())
		})
	}
}

func TestEnvironment_IsDev(t *testing.T) {
	tests := []struct {
		name     string
		env      oglconfig.Environment
		expected bool
	}{
		{
			name:     "development is dev",
			env:      oglconfig.EnvironmentDevelopment,
			expected: true,
		},
		{
			name:     "staging is not dev",
			env:      oglconfig.EnvironmentStaging,
			expected: false,
		},
		{
			name:     "production is not dev",
			env:      oglconfig.EnvironmentProduction,
			expected: false,
		},
		{
			name:     "testing is not dev",
			env:      oglconfig.EnvironmentTesting,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.env.IsDev()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEnvironment_String(t *testing.T) {
	tests := []struct {
		name     string
		env      oglconfig.Environment
		expected string
	}{
		{
			name:     "development",
			env:      oglconfig.EnvironmentDevelopment,
			expected: "development",
		},
		{
			name:     "staging",
			env:      oglconfig.EnvironmentStaging,
			expected: "staging",
		},
		{
			name:     "production",
			env:      oglconfig.EnvironmentProduction,
			expected: "production",
		},
		{
			name:     "testing",
			env:      oglconfig.EnvironmentTesting,
			expected: "testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.env.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEnvironment_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		env      oglconfig.Environment
		expected bool
	}{
		{
			name:     "valid development",
			env:      oglconfig.EnvironmentDevelopment,
			expected: true,
		},
		{
			name:     "valid staging",
			env:      oglconfig.EnvironmentStaging,
			expected: true,
		},
		{
			name:     "valid production",
			env:      oglconfig.EnvironmentProduction,
			expected: true,
		},
		{
			name:     "valid testing",
			env:      oglconfig.EnvironmentTesting,
			expected: true,
		},
		{
			name:     "invalid environment",
			env:      oglconfig.Environment("invalid"),
			expected: false,
		},
		{
			name:     "empty environment",
			env:      oglconfig.Environment(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.env.IsValid()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  oglconfig.Environment
		shouldErr bool
	}{
		{
			name:      "parse development",
			input:     "development",
			expected:  oglconfig.EnvironmentDevelopment,
			shouldErr: false,
		},
		{
			name:      "parse staging",
			input:     "staging",
			expected:  oglconfig.EnvironmentStaging,
			shouldErr: false,
		},
		{
			name:      "parse production",
			input:     "production",
			expected:  oglconfig.EnvironmentProduction,
			shouldErr: false,
		},
		{
			name:      "parse testing",
			input:     "testing",
			expected:  oglconfig.EnvironmentTesting,
			shouldErr: false,
		},
		{
			name:      "parse invalid environment",
			input:     "invalid",
			expected:  oglconfig.Environment(""),
			shouldErr: true,
		},
		{
			name:      "parse empty string",
			input:     "",
			expected:  oglconfig.Environment(""),
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := oglconfig.ParseEnvironment(tt.input)
			if tt.shouldErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, oglconfig.ErrInvalidEnvironment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestEnvironmentValues(t *testing.T) {
	values := oglconfig.EnvironmentValues()

	assert.Len(t, values, 4)
	assert.Contains(t, values, oglconfig.EnvironmentDevelopment)
	assert.Contains(t, values, oglconfig.EnvironmentStaging)
	assert.Contains(t, values, oglconfig.EnvironmentProduction)
	assert.Contains(t, values, oglconfig.EnvironmentTesting)
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    oglslog.LogLevel
		expected string
	}{
		{
			name:     "debug level",
			level:    oglslog.LogLevel("debug"),
			expected: "debug",
		},
		{
			name:     "info level",
			level:    oglslog.LogLevel("info"),
			expected: "info",
		},
		{
			name:     "warn level",
			level:    oglslog.LogLevel("warn"),
			expected: "warn",
		},
		{
			name:     "error level",
			level:    oglslog.LogLevel("error"),
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestLogLevel_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		level    oglslog.LogLevel
		expected bool
	}{
		{
			name:     "valid debug",
			level:    oglslog.LogLevel("debug"),
			expected: true,
		},
		{
			name:     "valid info",
			level:    oglslog.LogLevel("info"),
			expected: true,
		},
		{
			name:     "valid warn",
			level:    oglslog.LogLevel("warn"),
			expected: true,
		},
		{
			name:     "valid error",
			level:    oglslog.LogLevel("error"),
			expected: true,
		},
		{
			name:     "invalid level",
			level:    oglslog.LogLevel("invalid"),
			expected: false,
		},
		{
			name:     "empty string",
			level:    oglslog.LogLevel(""),
			expected: false,
		},
		{
			name:     "uppercase",
			level:    oglslog.LogLevel("DEBUG"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.IsValid()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestLogLevel_Level(t *testing.T) {
	tests := []struct {
		name     string
		logLevel oglslog.LogLevel
		expected slog.Level
	}{
		{
			name:     "debug level",
			logLevel: oglslog.LogLevel("debug"),
			expected: slog.LevelDebug,
		},
		{
			name:     "info level",
			logLevel: oglslog.LogLevel("info"),
			expected: slog.LevelInfo,
		},
		{
			name:     "warn level",
			logLevel: oglslog.LogLevel("warn"),
			expected: slog.LevelWarn,
		},
		{
			name:     "error level",
			logLevel: oglslog.LogLevel("error"),
			expected: slog.LevelError,
		},
		{
			name:     "invalid defaults to info",
			logLevel: oglslog.LogLevel("invalid"),
			expected: slog.LevelInfo,
		},
		{
			name:     "empty defaults to info",
			logLevel: oglslog.LogLevel(""),
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.logLevel.SlogLevel()
			assert.Equal(t, tt.expected, got)
		})
	}
}
