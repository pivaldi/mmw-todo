package config

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"strconv"

	oglconfig "github.com/ovya/ogl/config"
	"github.com/rotisserie/eris"
)

//go:embed configs/*.toml
var embeddedFS embed.FS

// getConfigFS returns the filesystem to use for reading config files.
// This is a variable so it can be mocked in tests.
var getConfigFS = func() fs.FS {
	return embeddedFS
}

type Database struct {
	User     string `mapstructure:"user"`
	password string
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Name     string `mapstructure:"name"`
}

func (d *Database) URL() string {
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", d.Host, d.Port),
		Path:   d.Name,
	}

	if d.User != "" {
		if d.password != "" {
			u.User = url.UserPassword(d.User, d.password)
		} else {
			u.User = url.User(d.User)
		}
	}

	return u.String()
}

type Port int16

func (p Port) String() string {
	return ":" + strconv.Itoa(int(p))
}

type LogLevel string

// SlogLevel returns the slog.SlogLevel value corresponding to the string level
func (l LogLevel) SlogLevel() slog.Level {
	switch string(l) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // default to info
	}
}

// String implements the Stringer interface
func (l LogLevel) String() string {
	return string(l)
}

// IsValid checks if the LogLevel value is valid
func (l LogLevel) IsValid() bool {
	switch string(l) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

type Server struct {
	Port   Port   `mapstructure:"port"`
	Scheme string `mapstructure:"scheme"`
	Host   string `mapstructure:"host"`
}

func (s Server) URL(path string, queries map[string]string) string {
	port := ""
	if (s.Scheme != "http" || s.Port != 80) && (s.Scheme != "https" || s.Port != 443) {
		port = s.Port.String()
	}

	u := &url.URL{
		Scheme: s.Scheme,
		Host:   s.Host + port,
		Path:   path,
	}

	q := u.Query()
	for key, value := range queries {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

type Config struct {
	Database    *Database   `mapstructure:"database"`
	Port        string      `mapstructure:"port"`
	Environment Environment `env:"APP_ENV, required" mapstructure:"environment"`
	AppName     string      `env:"APP_NAME"`
	Server      *Server     `mapstructure:"server"`
	LogLevel    LogLevel    `mapstructure:"log-level"`
}

// GetAppEnv returns the App environnement variable (prod, testing, etc)
func (c *Config) GetAppEnv() fmt.Stringer {
	return c.Environment
}

// Load loads the configurations from embedded files:
// - configs/default.toml
// - configs/<APP_ENV>.toml if exist
// If envs is not nil, use as environment variable (eg. unit-tests)
func Load(ctx context.Context, envs map[string]string) (*Config, error) {
	config := new(Config)

	// Get password from environment
	var password string
	if envs != nil {
		password = envs["DB_PASSWORD"]
	} else {
		password = os.Getenv("DB_PASSWORD")
	}

	if password == "" {
		return nil, eris.New("env var DB_PASSWORD not set")
	}

	configFS := getConfigFS()
	err := oglconfig.NewContext(ctx, configFS, envs).Fill(config)
	if err != nil {
		return nil, eris.Wrap(err, "error filling config")
	}

	if config.Database != nil {
		config.Database.password = password
	}

	return config, nil
}
