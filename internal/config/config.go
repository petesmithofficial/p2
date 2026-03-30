package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"p2/internal/powers"
)

const (
	dirName  = "p2"
	fileName = "config.json"
)

type Config struct {
	LowerBound            int  `json:"lower_bound"`
	UpperBound            int  `json:"upper_bound"`
	UseCommas             bool `json:"use_commas"`
	CopySingleToClipboard bool `json:"copy_single_to_clipboard"`
}

func Default() Config {
	return Config{
		LowerBound:            0,
		UpperBound:            16,
		UseCommas:             true,
		CopySingleToClipboard: true,
	}
}

func Path() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determine user config directory: %w", err)
	}

	return PathFromUserConfigDir(userConfigDir), nil
}

func DisplayPath() string {
	path, err := Path()
	if err == nil {
		return path
	}

	return filepath.Join("<user-config-dir>", dirName, fileName)
}

func PathFromUserConfigDir(userConfigDir string) string {
	return filepath.Join(userConfigDir, dirName, fileName)
}

func Load() (Config, string, error) {
	path, err := Path()
	if err != nil {
		return Config{}, "", err
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		return Config{}, path, err
	}

	return cfg, path, nil
}

func Save(cfg Config) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}

	if err := SaveToPath(path, cfg); err != nil {
		return "", err
	}

	return path, nil
}

func SaveToPath(path string, cfg Config) error {
	if err := validate(cfg); err != nil {
		return fmt.Errorf("validate %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory for %s: %w", path, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func LoadFromPath(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}

		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}

	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("validate %s: %w", path, err)
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.LowerBound < 0 || cfg.LowerBound > int(powers.MaxExponent) {
		return fmt.Errorf("lower_bound must be between 0 and %d", powers.MaxExponent)
	}

	if cfg.UpperBound < 0 || cfg.UpperBound > int(powers.MaxExponent) {
		return fmt.Errorf("upper_bound must be between 0 and %d", powers.MaxExponent)
	}

	if cfg.LowerBound > cfg.UpperBound {
		return fmt.Errorf("lower_bound must be less than or equal to upper_bound")
	}

	return nil
}
