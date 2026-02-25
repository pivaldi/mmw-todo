package config

import (
	"context"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_URL(t *testing.T) {
	tests := []struct {
		name     string
		database Database
		want     string
	}{
		{
			name: "complete database config with user and password",
			database: Database{
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     "5432",
				Name:     "testdb",
			},
			want: "postgres://testuser:testpass@localhost:5432/testdb",
		},
		{
			name: "database config with user but no password",
			database: Database{
				User: "testuser",
				Host: "localhost",
				Port: "5432",
				Name: "testdb",
			},
			want: "postgres://testuser@localhost:5432/testdb",
		},
		{
			name: "database config without user and password",
			database: Database{
				Host: "localhost",
				Port: "5432",
				Name: "testdb",
			},
			want: "postgres://localhost:5432/testdb",
		},
		{
			name: "database config with empty user",
			database: Database{
				User:     "",
				Password: "testpass",
				Host:     "db.example.com",
				Port:     "5432",
				Name:     "proddb",
			},
			want: "postgres://db.example.com:5432/proddb",
		},
		{
			name: "database config with custom host and port",
			database: Database{
				User:     "admin",
				Password: "secret",
				Host:     "192.168.1.100",
				Port:     "54321",
				Name:     "myapp",
			},
			want: "postgres://admin:secret@192.168.1.100:54321/myapp",
		},
		{
			name: "database config with special characters in password",
			database: Database{
				User:     "user",
				Password: "p@ss:word#123",
				Host:     "localhost",
				Port:     "5432",
				Name:     "db",
			},
			want: "postgres://user:p%40ss%3Aword%23123@localhost:5432/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.database.URL()
			assert.Equal(t, tt.want, got, "Database.URL() should return correct postgres URL")
		})
	}
}

func TestLoad_MissingEnvironmentVariables(t *testing.T) {
	ctx := context.Background()

	t.Run("empty environment variables map", func(t *testing.T) {
		envs := map[string]string{}
		_, err := Load(ctx, envs)
		assert.Error(t, err, "Load() should return error when required environment variables are missing")
	})

	t.Run("missing APP_EN", func(t *testing.T) {
		envs := map[string]string{
			"DB_PASSWORD": "test_password",
		}
		_, err := Load(ctx, envs)
		assert.Error(t, err, "Load() should return error when APP_EN is missing")
	})

	t.Run("missing DB_PASSWORD", func(t *testing.T) {
		envs := map[string]string{
			"APP_EN": "test",
		}
		_, err := Load(ctx, envs)
		assert.Error(t, err, "Load() should return error when DB_PASSWORD is missing")
	})
}

func TestLoad_Success(t *testing.T) {
	ctx := context.Background()

	envs := map[string]string{
		"DB_PASSWORD": "test_password",
		"APP_EN":      "default",
	}

	config, err := Load(ctx, envs)
	require.NoError(t, err, "Load() should not return error with valid environment variables")
	require.NotNil(t, config, "Load() should return non-nil config")
	require.NotNil(t, config.Database, "Config should have non-nil Database")

	// Verify database config was loaded from default config
	assert.Equal(t, "rcv", config.Database.User, "Database.User should match default config")
	assert.Equal(t, "localhost", config.Database.Host, "Database.Host should match default config")
	assert.Equal(t, "4332", config.Database.Port, "Database.Port should match default config")
	assert.Equal(t, "poc", config.Database.Name, "Database.Name should match default config")

	// Verify environment variable was processed
	assert.Equal(t, "test_password", config.Database.Password, "Database.Password should match environment variable")
	assert.Equal(t, "default", config.Environment, "Config.Environment should match environment variable")
}

func TestLoad_OptionalEnvironmentConfig(t *testing.T) {
	ctx := context.Background()

	// Environment-specific config is optional and should not cause an error
	envs := map[string]string{
		"DB_PASSWORD": "test_password",
		"APP_EN":      "nonexistent_env",
	}

	config, err := Load(ctx, envs)
	assert.NoError(t, err, "Load() should not error when environment-specific config doesn't exist")
	require.NotNil(t, config, "Load() should return non-nil config")
	require.NotNil(t, config.Database, "Config should have non-nil Database")

	// Verify default config was loaded
	assert.Equal(t, "rcv", config.Database.User, "Database.User should match default config when env-specific config doesn't exist")
}

func TestLoad_WithNilEnvs(t *testing.T) {
	ctx := context.Background()

	// When envs is nil, Load should use actual OS environment variables
	// This test will fail if the required environment variables are not set
	// but that's expected behavior
	_, err := Load(ctx, nil)

	// We can't reliably test this without manipulating OS environment
	// Just verify the function handles nil envs without panicking
	if err != nil {
		// Expected to fail if OS env vars are not set
		t.Logf("Load() with nil envs failed as expected when env vars not set: %v", err)
	}
}

func TestEnvUnmarshal(t *testing.T) {
	ctx := context.Background()

	t.Run("with envs map", func(t *testing.T) {
		config := new(Config)
		envs := map[string]string{
			"DB_PASSWORD": "secret_pass",
			"APP_EN":      "production",
		}

		err := envUnmarshal(ctx, config, envs)
		require.NoError(t, err, "envUnmarshal() should not return error with valid envs map")

		if config.Database == nil {
			config.Database = &Database{}
		}

		assert.Equal(t, "production", config.Environment, "Environment should be set from envs map")
	})

	t.Run("with nil envs uses OS environment", func(t *testing.T) {
		config := new(Config)

		// This will likely fail since OS env vars aren't set, but shouldn't panic
		err := envUnmarshal(ctx, config, nil)
		if err != nil {
			t.Logf("envUnmarshal() with nil envs failed as expected: %v", err)
		}
	})
}

func TestLoad_FullURL(t *testing.T) {
	ctx := context.Background()

	envs := map[string]string{
		"DB_PASSWORD": "mypassword123",
		"APP_EN":      "default",
	}

	config, err := Load(ctx, envs)
	require.NoError(t, err, "Load() should not return error")

	// Test that we can generate a full database URL
	dbURL := config.Database.URL()
	expected := "postgres://rcv:mypassword123@localhost:4332/poc"

	assert.Equal(t, expected, dbURL, "Database.URL() should return complete postgres connection URL")
}

func TestLoad_WithRealOSEnvironment(t *testing.T) {
	ctx := context.Background()

	// Save original environment variables
	origDBPassword := os.Getenv("DB_PASSWORD")
	origAppEnv := os.Getenv("APP_EN")
	defer func() {
		if origDBPassword != "" {
			os.Setenv("DB_PASSWORD", origDBPassword)
		} else {
			os.Unsetenv("DB_PASSWORD")
		}
		if origAppEnv != "" {
			os.Setenv("APP_EN", origAppEnv)
		} else {
			os.Unsetenv("APP_EN")
		}
	}()

	// Set environment variables using os.Setenv
	err := os.Setenv("DB_PASSWORD", "os_test_password")
	require.NoError(t, err, "os.Setenv should not fail")
	err = os.Setenv("APP_EN", "default")
	require.NoError(t, err, "os.Setenv should not fail")

	// Load config with nil envs to use OS environment
	config, err := Load(ctx, nil)
	require.NoError(t, err, "Load() should not return error with OS environment variables set")
	require.NotNil(t, config, "Load() should return non-nil config")
	require.NotNil(t, config.Database, "Config should have non-nil Database")

	// Verify OS environment variables were used
	assert.Equal(t, "os_test_password", config.Database.Password, "Database.Password should match OS environment variable")
	assert.Equal(t, "default", config.Environment, "Config.Environment should match OS environment variable")

	// Verify default config was still loaded
	assert.Equal(t, "rcv", config.Database.User, "Database.User should match default config")
	assert.Equal(t, "localhost", config.Database.Host, "Database.Host should match default config")
	assert.Equal(t, "4332", config.Database.Port, "Database.Port should match default config")
	assert.Equal(t, "poc", config.Database.Name, "Database.Name should match default config")
}

func TestLoad_ConfigMergingWithEnvironmentFile(t *testing.T) {
	ctx := context.Background()

	// Mock the filesystem with default and testing configs
	origFS := getConfigFS
	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": &fstest.MapFile{
				Data: []byte(`[database]
user = "rcv"
host = "localhost"
port = "4332"
name = "poc"
`),
			},
			"configs/testing.toml": &fstest.MapFile{
				Data: []byte(`[database]
user = "test_user"
host = "test.example.com"
port = "5433"
name = "testdb"
`),
			},
		}
	}
	defer func() { getConfigFS = origFS }()

	// Set environment variables to use the testing environment
	envs := map[string]string{
		"DB_PASSWORD": "merged_password",
		"APP_EN":      "testing",
	}

	config, err := Load(ctx, envs)
	require.NoError(t, err, "Load() should not return error with environment-specific config file")
	require.NotNil(t, config, "Load() should return non-nil config")
	require.NotNil(t, config.Database, "Config should have non-nil Database")

	// Verify that environment-specific config overrides default config
	assert.Equal(t, "test_user", config.Database.User, "Database.User should be from testing config, not default")
	assert.Equal(t, "test.example.com", config.Database.Host, "Database.Host should be from testing config, not default")
	assert.Equal(t, "5433", config.Database.Port, "Database.Port should be from testing config, not default")
	assert.Equal(t, "testdb", config.Database.Name, "Database.Name should be from testing config, not default")

	// Verify environment variable was still processed
	assert.Equal(t, "merged_password", config.Database.Password, "Database.Password should match environment variable")
	assert.Equal(t, "testing", config.Environment, "Config.Environment should match environment variable")

	// Verify the merged URL
	dbURL := config.Database.URL()
	expected := "postgres://test_user:merged_password@test.example.com:5433/testdb"
	assert.Equal(t, expected, dbURL, "Database.URL() should reflect merged configuration")
}

func TestLoad_ConfigMergingPartialOverride(t *testing.T) {
	ctx := context.Background()

	// Mock the filesystem with default and partial configs
	origFS := getConfigFS
	getConfigFS = func() fs.FS {
		return fstest.MapFS{
			"configs/default.toml": &fstest.MapFile{
				Data: []byte(`[database]
user = "rcv"
host = "localhost"
port = "4332"
name = "poc"
`),
			},
			"configs/partial.toml": &fstest.MapFile{
				Data: []byte(`[database]
host = "partial.example.com"
port = "8888"
`),
			},
		}
	}
	defer func() { getConfigFS = origFS }()

	envs := map[string]string{
		"DB_PASSWORD": "partial_password",
		"APP_EN":      "partial",
	}

	config, err := Load(ctx, envs)
	require.NoError(t, err, "Load() should not return error with partial environment-specific config")
	require.NotNil(t, config, "Load() should return non-nil config")
	require.NotNil(t, config.Database, "Config should have non-nil Database")

	// Verify that only specified fields are overridden
	assert.Equal(t, "partial.example.com", config.Database.Host, "Database.Host should be overridden from partial config")

	// Verify that non-specified fields still come from default config
	assert.Equal(t, "rcv", config.Database.User, "Database.User should still be from default config")
	assert.Equal(t, "8888", config.Database.Port, "Database.Port should still be from default config")
	assert.Equal(t, "poc", config.Database.Name, "Database.Name should still be from default config")

	// Verify environment variable was processed
	assert.Equal(t, "partial_password", config.Database.Password, "Database.Password should match environment variable")
	assert.Equal(t, "partial", config.Environment, "Config.Environment should match environment variable")
}

func TestLoad_WithRealOSEnvironment_MissingRequired(t *testing.T) {
	ctx := context.Background()

	// Save original environment variables
	origDBPassword := os.Getenv("DB_PASSWORD")
	origAppEnv := os.Getenv("APP_EN")
	defer func() {
		if origDBPassword != "" {
			os.Setenv("DB_PASSWORD", origDBPassword)
		} else {
			os.Unsetenv("DB_PASSWORD")
		}
		if origAppEnv != "" {
			os.Setenv("APP_EN", origAppEnv)
		} else {
			os.Unsetenv("APP_EN")
		}
	}()

	// Unset required environment variables
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("APP_EN")

	// Load config with nil envs should fail
	_, err := Load(ctx, nil)
	assert.Error(t, err, "Load() should return error when required OS environment variables are missing")
}
