package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	configDirName  = ".config"
	appDirName     = "dbc"
	configFileName = "config.toml"
)

var (
	ErrMissingDatabase     = errors.New("database config missing")
	ErrMissingDatabaseName = errors.New("database name is required")
	ErrMissingDatabasePath = errors.New("database path is required")
)

type Config struct {
	Database *DatabaseConfig `toml:"database"`
}

type DatabaseConfig struct {
	Name string `toml:"name"`
	Path string `toml:"db_path"`
}

func Decode(r io.Reader) (Config, error) {
	var cfg Config
	decoder := toml.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.Database == nil {
		return ErrMissingDatabase
	}
	if strings.TrimSpace(c.Database.Name) == "" {
		return ErrMissingDatabaseName
	}
	if strings.TrimSpace(c.Database.Path) == "" {
		return ErrMissingDatabasePath
	}
	return nil
}

func PathFromHome(home string) string {
	return filepath.Join(home, configDirName, appDirName, configFileName)
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return PathFromHome(home), nil
}

func LoadFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	return Decode(file)
}
