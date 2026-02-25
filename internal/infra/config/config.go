package config

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/rotisserie/eris"
	"github.com/sethvargo/go-envconfig"
	"github.com/spf13/viper"
)

//go:embed configs/*.toml
var embeddedFS embed.FS

// getConfigFS returns the filesystem to use for reading config files.
// This is a variable so it can be mocked in tests.
var getConfigFS = func() fs.FS {
	return embeddedFS
}

type Database struct {
	User     string `toml:"user"`
	Password string `env:"DB_PASSWORD, required"`
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	Name     string `toml:"name"`
}

func (d Database) URL() string {
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", d.Host, d.Port),
		Path:   d.Name,
	}

	if d.User != "" {
		if d.Password != "" {
			u.User = url.UserPassword(d.User, d.Password)
		} else {
			u.User = url.User(d.User)
		}
	}

	return u.String()
}

type Config struct {
	Database    *Database `toml:"database"`
	Port        string    `toml:"port"`
	Environment string    `env:"APP_EN, required"`
}

// Load loads the configurations from enbended files:
// - configs/default.toml
// - configs/<APP_ENV>.toml if exist
// If envs is not nil, use as environnement variable (eg. unit-tests)
func Load(ctx context.Context, envs map[string]string) (*Config, error) {
	config := new(Config)

	if err := envUnmarshal(ctx, config, envs); err != nil {
		return nil, err
	}

	viper.SetConfigType("toml")
	configFS := getConfigFS()
	defaultConfig, err := fs.ReadFile(configFS, "configs/default.toml")
	if err != nil {
		return nil, eris.Wrap(err, "failed to read the default configuration")
	}
	if err := viper.ReadConfig(bytes.NewBuffer(defaultConfig)); err != nil {
		return nil, eris.Wrap(err, "viper failed to read the default configuration")
	}

	file := config.Environment + ".toml"
	envConfig, err := fs.ReadFile(configFS, "configs/"+file)
	if err == nil { // Env config may not exist.
		// Merge environment-specific config
		if err := viper.MergeConfig(bytes.NewBuffer(envConfig)); err != nil {
			return nil, eris.Wrapf(err, "viper failed merging the configuration %s", file)
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, eris.Wrap(err, "failed to unmarshal config")
	}

	return config, nil
}

func envUnmarshal(ctx context.Context, config *Config, envs map[string]string) error {
	if envs == nil {
		if err := envconfig.Process(ctx, config); err != nil {
			return eris.Wrap(err, "envconfig process error")
		}

		return nil
	}

	if err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target:   config,
		Lookuper: envconfig.MapLookuper(envs),
	}); err != nil {
		return eris.Wrapf(err, "envconfig process error with given envs: %v", envs)
	}

	return nil
}
