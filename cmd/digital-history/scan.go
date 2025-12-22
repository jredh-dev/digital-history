// Digital History - Digital footprint analysis tool
// Copyright (C) 2025 jredh
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package main

import (
	"context"
	"fmt"

	"github.com/jredh-dev/digital-history/internal/analyzer"
	"github.com/jredh-dev/digital-history/internal/platform/reddit"
	"github.com/jredh-dev/digital-history/internal/storage"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan <platform>",
	Short: "Scan a platform for digital footprint data",
	Long:  `Collect posts, comments, and other content from the specified platform`,
	Args:  cobra.ExactArgs(1),
	RunE:  runScan,
}

var (
	scanUsername string
	scanAnalyze  bool
)

func init() {
	scanCmd.Flags().StringVarP(&scanUsername, "username", "u", "", "Username to scan (required)")
	scanCmd.Flags().BoolVarP(&scanAnalyze, "analyze", "a", true, "Run analysis after scanning")
	scanCmd.MarkFlagRequired("username")
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	platform := args[0]

	switch platform {
	case "reddit":
		return scanReddit(cmd.Context())
	default:
		return fmt.Errorf("unsupported platform: %s (supported: reddit)", platform)
	}
}

func scanReddit(ctx context.Context) error {
	// Validate Reddit credentials
	if cfg.Reddit.ClientID == "" || cfg.Reddit.ClientSecret == "" {
		return fmt.Errorf("Reddit credentials not configured. Run:\n" +
			"  digital-history config set reddit.client_id YOUR_CLIENT_ID\n" +
			"  digital-history config set reddit.client_secret YOUR_CLIENT_SECRET\n" +
			"  digital-history config set reddit.username YOUR_USERNAME\n" +
			"  digital-history config set reddit.password YOUR_PASSWORD")
	}

	// Get database passphrase
	passphrase := cfg.Database.Passphrase
	if passphrase == "" {
		fmt.Print("Enter database passphrase: ")
		fmt.Scanln(&passphrase)
		if passphrase == "" {
			return fmt.Errorf("passphrase required for encrypted storage")
		}
	}

	// Open storage
	store, err := storage.New(cfg.Database.Path, passphrase)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer store.Close()

	// Create Reddit client
	creds := reddit.Credentials{
		ClientID:     cfg.Reddit.ClientID,
		ClientSecret: cfg.Reddit.ClientSecret,
		Username:     cfg.Reddit.Username,
		Password:     cfg.Reddit.Password,
	}

	client, err := reddit.New(creds, store)
	if err != nil {
		return fmt.Errorf("failed to create reddit client: %w", err)
	}

	// Scan user
	if err := client.ScanUser(ctx, scanUsername); err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Run analysis if requested
	if scanAnalyze {
		fmt.Println("\nRunning content analysis...")
		analyzer := analyzer.New(store)
		if err := analyzer.AnalyzeUser(ctx, scanUsername); err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		// Show summary
		summary, err := analyzer.GetSummary(scanUsername)
		if err != nil {
			return fmt.Errorf("failed to get summary: %w", err)
		}

		fmt.Println("\n=== Analysis Summary ===")
		if len(summary) == 0 {
			fmt.Println("No issues found!")
		} else {
			for key, count := range summary {
				fmt.Printf("  %s: %d\n", key, count)
			}
		}
	}

	return nil
}
