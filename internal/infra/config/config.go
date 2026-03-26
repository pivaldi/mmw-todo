// services/todo/config/config.go
package config

import (
	"context"
	"embed"
	"io/fs"

	oglconfig "github.com/ovya/ogl/config"
	oglpfconfig "github.com/ovya/ogl/platform/config"
	oglslog "github.com/ovya/ogl/slog"
	"github.com/rotisserie/eris"
)

//go:embed configs/*.toml
var embeddedFS embed.FS

// getConfigFS returns the filesystem to use for reading config files.
// This is a variable so it can be mocked in tests.
var getConfigFS = func() fs.FS {
	return embeddedFS
}

type Config struct {
	oglpfconfig.Base
	Database *oglpfconfig.Database `mapstructure:"database"`
	// Environment oglconfig.Environment `env:"APP_ENV, required" mapstructure:"environment"`
	Server     *oglpfconfig.Server `mapstructure:"server"`
	LogLevel   oglslog.LogLevel    `mapstructure:"log-level"`
	AuthServer *oglpfconfig.Server `mapstructure:"auth-server"`
}

var conf *Config

// Load reads the TOML files and automatically overrides them with Env Vars
// Load loads the configurations from embedded files:
// - configs/default.toml
// - configs/<APP_ENV>.toml if exist
// If envs is not nil, automatically overrides them with Env Vars.
func Load(ctx context.Context, envPrefix string) (*Config, error) {
	if conf != nil {
		return conf, nil
	}

	conf := new(Config)
	configFS := getConfigFS()

	err := oglconfig.NewContext(ctx, configFS, envPrefix).Fill(conf)
	if err != nil {
		return nil, eris.Wrap(err, "error filling config")
	}

	if conf.Database.Password == "" {
		return nil, eris.New("database password is empty")
	}

	return conf, nil
}
