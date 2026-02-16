package config

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mgierok/dbc/internal/application/port"
	"github.com/pelletier/go-toml/v2"
)

const (
	configDirName  = ".config"
	appDirName     = "dbc"
	configFileName = "config.toml"
)

var (
	ErrMissingDatabase         = errors.New("database config missing")
	ErrMissingDatabaseName     = errors.New("database name is required")
	ErrMissingDatabasePath     = errors.New("database path is required")
	ErrDatabaseIndexOutOfRange = errors.New("database index out of range")
)

type Config struct {
	Databases []DatabaseConfig `toml:"databases"`
}

type DatabaseConfig struct {
	Name string `toml:"name"`
	Path string `toml:"db_path"`
}

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
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
	if len(c.Databases) == 0 {
		return ErrMissingDatabase
	}
	for _, database := range c.Databases {
		if strings.TrimSpace(database.Name) == "" {
			return ErrMissingDatabaseName
		}
		if strings.TrimSpace(database.Path) == "" {
			return ErrMissingDatabasePath
		}
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
	return ResolvePathForOS(runtime.GOOS, home, os.Getenv("APPDATA")), nil
}

func ResolvePathForOS(goos, home, appData string) string {
	if goos == "windows" {
		base := strings.TrimSpace(appData)
		if base == "" {
			base = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(base, appDirName, configFileName)
	}
	return PathFromHome(home)
}

func LoadFile(path string) (cfg Config, err error) {
	// #nosec G304 -- path is provided by trusted caller flow (default path or explicit local user input).
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	cfg, err = Decode(file)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s *Store) List(_ context.Context) ([]port.ConfigEntry, error) {
	cfg, err := LoadFile(s.path)
	if err != nil {
		return nil, err
	}
	result := make([]port.ConfigEntry, len(cfg.Databases))
	for i, database := range cfg.Databases {
		result[i] = port.ConfigEntry{
			Name:   database.Name,
			DBPath: database.Path,
		}
	}
	return result, nil
}

func (s *Store) Create(_ context.Context, entry port.ConfigEntry) error {
	cfg, err := LoadFile(s.path)
	if err != nil {
		return err
	}
	cfg.Databases = append(cfg.Databases, DatabaseConfig{
		Name: entry.Name,
		Path: entry.DBPath,
	})
	return saveFile(s.path, cfg)
}

func (s *Store) Update(_ context.Context, index int, entry port.ConfigEntry) error {
	cfg, err := LoadFile(s.path)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(cfg.Databases) {
		return ErrDatabaseIndexOutOfRange
	}
	cfg.Databases[index] = DatabaseConfig{
		Name: entry.Name,
		Path: entry.DBPath,
	}
	return saveFile(s.path, cfg)
}

func (s *Store) Delete(_ context.Context, index int) error {
	cfg, err := LoadFile(s.path)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(cfg.Databases) {
		return ErrDatabaseIndexOutOfRange
	}
	cfg.Databases = append(cfg.Databases[:index], cfg.Databases[index+1:]...)
	return saveFile(s.path, cfg)
}

func (s *Store) ActivePath(_ context.Context) (string, error) {
	return s.path, nil
}

func saveFile(path string, cfg Config) (err error) {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".config.*.toml")
	if err != nil {
		return err
	}
	tmpPath := file.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		_ = file.Close()
		return err
	}
	if closeErr := file.Close(); closeErr != nil {
		return closeErr
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	return nil
}
