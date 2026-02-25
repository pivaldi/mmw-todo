package config

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_URL(t *testing.T) {
	tests := []struct {
		name     string
		db       *Database
		expected string
	}{
		{
			name: "with user and password",
			db: &Database{
				User:     "testuser",
				password: "testpass",
				Host:     "localhost",
				Port:     "5432",
				Name:     "testdb",
			},
			expected: "postgres://testuser:testpass@localhost:5432/testdb",
		},
		{
			name: "with user only",
			db: &Database{
				User:     "testuser",
				password: "",
				Host:     "localhost",
				Port:     "5432",
				Name:     "testdb",
			},
			expected: "postgres://testuser@localhost:5432/testdb",
		},
		{
			name: "without credentials",
			db: &Database{
				User:     "",
				password: "",
				Host:     "localhost",
				Port:     "5432",
				Name:     "testdb",
			},
			expected: "postgres://localhost:5432/testdb",
		},
		{
			name: "with special characters in password",
			db: &Database{
				User:     "user",
				password: "p@ss:word",
				Host:     "db.example.com",
				Port:     "5432",
				Name:     "mydb",
			},
			expected: "postgres://user:p%40ss%3Aword@db.example.com:5432/mydb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.URL()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestLoad_Success(t *testing.T) {
	// Save and restore original getConfigFS
	origFS := getConfigFS
	defer func() { getConfigFS = origFS }()

	// Mock filesystem with config files
	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": &fstest.MapFile{
				Data: []byte(`app-name = "TestApp"
port = "8080"

[database]
user = "rcv"
host = "localhost"
port = "5432"
name = "testdb"
`),
			},
			"configs/testing.toml": &fstest.MapFile{
				Data: []byte(`port = "9090"

[database]
host = "test-host"
`),
			},
		}
	}

	ctx := context.Background()
	envs := map[string]string{
		"DB_PASSWORD": "secret123",
		"APP_ENV":     "testing",
	}

	config, err := Load(ctx, envs)

	require.NoError(t, err)
	require.NotNil(t, config)
	require.NotNil(t, config.Database)

	// Verify config was loaded and merged
	assert.Equal(t, "TestApp", config.AppName)
	assert.Equal(t, "9090", config.Port) // From testing.toml
	assert.Equal(t, "testing", config.Environment)

	// Verify database config
	assert.Equal(t, "rcv", config.Database.User)
	assert.Equal(t, "test-host", config.Database.Host) // From testing.toml
	assert.Equal(t, "5432", config.Database.Port)
	assert.Equal(t, "testdb", config.Database.Name)

	// Verify URL generation includes password
	url := config.Database.URL()
	assert.Contains(t, url, "secret123")
	assert.Contains(t, url, "test-host")
}

func TestLoad_MissingPasswordEnv(t *testing.T) {
	// Save and restore original getConfigFS
	origFS := getConfigFS
	defer func() { getConfigFS = origFS }()

	// Mock filesystem
	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": &fstest.MapFile{
				Data: []byte(`[database]
user = "rcv"
host = "localhost"
port = "5432"
name = "testdb"
`),
			},
		}
	}

	ctx := context.Background()
	envs := map[string]string{
		"APP_ENV": "development",
		// DB_PASSWORD is missing
	}

	config, err := Load(ctx, envs)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "DB_PASSWORD")
}

func TestLoad_WithDefaultConfigOnly(t *testing.T) {
	// Save and restore original getConfigFS
	origFS := getConfigFS
	defer func() { getConfigFS = origFS }()

	// Mock filesystem with only default config
	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": &fstest.MapFile{
				Data: []byte(`app-name = "DefaultApp"
port = "8080"

[database]
user = "admin"
host = "localhost"
port = "5432"
name = "defaultdb"
`),
			},
		}
	}

	ctx := context.Background()
	envs := map[string]string{
		"DB_PASSWORD": "password",
		"APP_ENV":     "production", // No production.toml exists
	}

	config, err := Load(ctx, envs)

	require.NoError(t, err)
	require.NotNil(t, config)

	// Should use default config values
	assert.Equal(t, "DefaultApp", config.AppName)
	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "production", config.Environment)
}

func TestConfig_GetAppEnv(t *testing.T) {
	config := &Config{
		Environment: "staging",
	}

	assert.Equal(t, "staging", config.GetAppEnv())
}
