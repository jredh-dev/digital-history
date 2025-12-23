// Digital History - Digital footprint analysis tool
// Copyright (C) 2025 jredh
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jredh-dev/digital-history/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all configuration (redacts sensitive values)",
	RunE:  runConfigShow,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Load current config
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".digital-history.yaml")

	cfg, err := config.Load(configPath)
	if err != nil {
		// Create new config if it doesn't exist
		cfg = &config.Config{}
	}

	// Set value based on key
	switch key {
	case "reddit.client_id":
		cfg.Reddit.ClientID = value
	case "reddit.client_secret":
		cfg.Reddit.ClientSecret = value
	case "reddit.username":
		cfg.Reddit.Username = value
	case "reddit.password":
		cfg.Reddit.Password = value
	case "database.path":
		cfg.Database.Path = value
	case "database.passphrase":
		cfg.Database.Passphrase = value
	case "analyzer.enable_slur_detection":
		cfg.Analyzer.EnableSlurDetection = strings.ToLower(value) == "true"
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s\n", key)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Get value based on key
	var value string
	switch key {
	case "reddit.client_id":
		value = cfg.Reddit.ClientID
	case "reddit.client_secret":
		value = redact(cfg.Reddit.ClientSecret)
	case "reddit.username":
		value = cfg.Reddit.Username
	case "reddit.password":
		value = redact(cfg.Reddit.Password)
	case "database.path":
		value = cfg.Database.Path
	case "database.passphrase":
		value = redact(cfg.Database.Passphrase)
	case "analyzer.enable_slur_detection":
		value = fmt.Sprintf("%t", cfg.Analyzer.EnableSlurDetection)
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	fmt.Println(value)
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	fmt.Println("=== Configuration ===")
	fmt.Printf("Database:\n")
	fmt.Printf("  path: %s\n", cfg.Database.Path)
	fmt.Printf("  passphrase: %s\n", redact(cfg.Database.Passphrase))
	fmt.Printf("\nReddit:\n")
	fmt.Printf("  client_id: %s\n", cfg.Reddit.ClientID)
	fmt.Printf("  client_secret: %s\n", redact(cfg.Reddit.ClientSecret))
	fmt.Printf("  username: %s\n", cfg.Reddit.Username)
	fmt.Printf("  password: %s\n", redact(cfg.Reddit.Password))
	fmt.Printf("\nAnalyzer:\n")
	fmt.Printf("  enable_slur_detection: %t\n", cfg.Analyzer.EnableSlurDetection)
	return nil
}

func redact(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}
