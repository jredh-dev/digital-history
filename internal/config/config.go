// Digital History - Digital footprint analysis tool
// Copyright (C) 2025 jredh
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Reddit   RedditConfig   `mapstructure:"reddit"`
	Analyzer AnalyzerConfig `mapstructure:"analyzer"`
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Path       string `mapstructure:"path"`
	Passphrase string `mapstructure:"passphrase"`
}

// RedditConfig holds Reddit API credentials
type RedditConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
}

// AnalyzerConfig holds analyzer settings
type AnalyzerConfig struct {
	EnableSlurDetection bool `mapstructure:"enable_slur_detection"`
}

// Load reads configuration from file or creates default
func Load(configFile string) (*Config, error) {
	v := viper.New()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		v.AddConfigPath(home)
		v.SetConfigName(".digital-history")
		v.SetConfigType("yaml")
	}

	// Set defaults
	v.SetDefault("database.path", filepath.Join(mustHomeDir(), ".digital-history", "data.db"))
	v.SetDefault("analyzer.enable_slur_detection", true)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is OK, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save(path string) error {
	v := viper.New()
	v.Set("database", c.Database)
	v.Set("reddit", c.Reddit)
	v.Set("analyzer", c.Analyzer)

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, ".digital-history.yaml")
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func mustHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return home
}
