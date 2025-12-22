// Digital History - Digital footprint analysis tool
// Copyright (C) 2025 jredh
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jredh-dev/digital-history/internal/storage"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export collected data and analysis results",
	RunE:  runExport,
}

var (
	exportUsername string
	exportFormat   string
	exportOutput   string
)

func init() {
	exportCmd.Flags().StringVarP(&exportUsername, "username", "u", "", "Username to export data for (required)")
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Export format (json)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportCmd.MarkFlagRequired("username")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	// Get database passphrase
	passphrase := cfg.Database.Passphrase
	if passphrase == "" {
		fmt.Fprint(os.Stderr, "Enter database passphrase: ")
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

	// Get all data
	posts, err := store.GetPostsByUsername(exportUsername)
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	comments, err := store.GetCommentsByUsername(exportUsername)
	if err != nil {
		return fmt.Errorf("failed to get comments: %w", err)
	}

	results, err := store.GetAnalysisByUsername(exportUsername)
	if err != nil {
		return fmt.Errorf("failed to get analysis results: %w", err)
	}

	// Create export structure
	export := map[string]interface{}{
		"username": exportUsername,
		"posts":    posts,
		"comments": comments,
		"analysis": results,
		"summary": map[string]int{
			"total_posts":    len(posts),
			"total_comments": len(comments),
			"total_issues":   len(results),
		},
	}

	// Marshal to JSON
	var data []byte
	switch exportFormat {
	case "json":
		data, err = json.MarshalIndent(export, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", exportFormat)
	}

	// Write output
	if exportOutput == "" {
		fmt.Println(string(data))
	} else {
		if err := os.WriteFile(exportOutput, data, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Exported to: %s\n", exportOutput)
	}

	return nil
}
