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

	"github.com/jredh-dev/digital-history/internal/storage"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List collected data and analysis results",
	RunE:  runList,
}

var listUsername string

func init() {
	listCmd.Flags().StringVarP(&listUsername, "username", "u", "", "Username to list data for (required)")
	listCmd.MarkFlagRequired("username")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Get posts
	posts, err := store.GetPostsByUsername(listUsername)
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	// Get comments
	comments, err := store.GetCommentsByUsername(listUsername)
	if err != nil {
		return fmt.Errorf("failed to get comments: %w", err)
	}

	// Get analysis results
	results, err := store.GetAnalysisByUsername(listUsername)
	if err != nil {
		return fmt.Errorf("failed to get analysis results: %w", err)
	}

	// Display summary
	fmt.Printf("\n=== Digital History: %s ===\n\n", listUsername)
	fmt.Printf("Posts:              %d\n", len(posts))
	fmt.Printf("Comments:           %d\n", len(comments))
	fmt.Printf("Analysis Results:   %d\n", len(results))

	if len(results) > 0 {
		fmt.Println("\n=== Top Issues ===")

		// Group by severity
		high, medium, low := 0, 0, 0
		for _, r := range results {
			switch r.Severity {
			case "high":
				high++
			case "medium":
				medium++
			case "low":
				low++
			}
		}

		if high > 0 {
			fmt.Printf("  🔴 High severity:   %d\n", high)
		}
		if medium > 0 {
			fmt.Printf("  🟡 Medium severity: %d\n", medium)
		}
		if low > 0 {
			fmt.Printf("  🟢 Low severity:    %d\n", low)
		}

		// Show sample issues
		fmt.Println("\nRecent Findings:")
		count := 0
		for _, r := range results {
			if count >= 10 {
				break
			}
			fmt.Printf("  [%s] %s: %s\n", r.Severity, r.Category, r.Description)
			count++
		}
		if len(results) > 10 {
			fmt.Printf("\n  ... and %d more (use 'export' to see all)\n", len(results)-10)
		}
	}

	return nil
}
