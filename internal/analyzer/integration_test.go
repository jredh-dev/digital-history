// Digital History - Digital footprint analysis tool
// Copyright (C) 2025 jredh
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jredh-dev/digital-history/internal/platform/reddit"
	"github.com/jredh-dev/digital-history/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_AnalyzeRedditUser is an integration test that fetches real data from Reddit
// and runs all analyzers. This test requires valid Reddit API credentials set as environment variables:
//
//	REDDIT_CLIENT_ID
//	REDDIT_CLIENT_SECRET
//	REDDIT_USERNAME
//	REDDIT_PASSWORD
//
// To run this test:
//
//	export REDDIT_CLIENT_ID="your_client_id"
//	export REDDIT_CLIENT_SECRET="your_client_secret"
//	export REDDIT_USERNAME="your_username"
//	export REDDIT_PASSWORD="your_password"
//	go test -v -tags=integration ./internal/analyzer/ -run TestIntegration_AnalyzeRedditUser
//
// Note: This test will make real API calls to Reddit and may take some time.
func TestIntegration_AnalyzeRedditUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check for required environment variables
	clientID := os.Getenv("REDDIT_CLIENT_ID")
	clientSecret := os.Getenv("REDDIT_CLIENT_SECRET")
	username := os.Getenv("REDDIT_USERNAME")
	password := os.Getenv("REDDIT_PASSWORD")

	if clientID == "" || clientSecret == "" || username == "" || password == "" {
		t.Skip("Skipping integration test: Reddit credentials not provided via environment variables")
	}

	// Use a well-known Reddit account for testing (spez is Reddit's CEO, has public posts)
	targetUsername := "spez"

	// Setup test storage
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration_test.db")
	store, err := storage.New(dbPath, "integration-test-passphrase")
	require.NoError(t, err, "Failed to create test storage")
	defer store.Close()

	// Setup Reddit client
	creds := reddit.Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
	}

	client, err := reddit.New(creds, store)
	require.NoError(t, err, "Failed to create Reddit client")

	t.Logf("Scanning Reddit user: %s", targetUsername)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Scan the user (collect posts and comments)
	err = client.ScanUser(ctx, targetUsername)
	require.NoError(t, err, "Failed to scan user")

	// Verify data was collected
	posts, err := store.GetPostsByUsername(targetUsername)
	require.NoError(t, err, "Failed to get posts")
	t.Logf("Collected %d posts", len(posts))
	assert.Greater(t, len(posts), 0, "Should have collected some posts")

	comments, err := store.GetCommentsByUsername(targetUsername)
	require.NoError(t, err, "Failed to get comments")
	t.Logf("Collected %d comments", len(comments))
	assert.Greater(t, len(comments), 0, "Should have collected some comments")

	// Run political analyzer
	t.Log("Running political content analyzer...")
	politicalAnalyzer := NewPoliticalAnalyzer(store)
	err = politicalAnalyzer.AnalyzeUser(ctx, targetUsername)
	require.NoError(t, err, "Political analyzer failed")

	// Run NSFW analyzer
	t.Log("Running NSFW content analyzer...")
	nsfwAnalyzer := NewNSFWAnalyzer(store)
	err = nsfwAnalyzer.AnalyzeUser(ctx, targetUsername)
	require.NoError(t, err, "NSFW analyzer failed")

	// Run original analyzer (PII, slurs)
	t.Log("Running PII/slur analyzer...")
	piiAnalyzer := New(store)
	err = piiAnalyzer.AnalyzeUser(ctx, targetUsername)
	require.NoError(t, err, "PII analyzer failed")

	// Get all analysis results
	results, err := store.GetAnalysisByUsername(targetUsername)
	require.NoError(t, err, "Failed to get analysis results")
	t.Logf("Total analysis results: %d", len(results))

	// Categorize results
	resultsByCategory := make(map[string]int)
	for _, result := range results {
		resultsByCategory[result.Category]++
	}

	t.Log("Analysis results by category:")
	for category, count := range resultsByCategory {
		t.Logf("  %s: %d findings", category, count)
	}

	// Since spez is Reddit's CEO, we expect to find political content
	politicalResults := filterByCategory(results, "political")
	t.Logf("Political findings: %d", len(politicalResults))

	if len(politicalResults) > 0 {
		t.Log("Sample political findings:")
		for i, result := range politicalResults {
			if i >= 5 {
				break // Only show first 5
			}
			t.Logf("  - [%s] %s: %s", result.ContentType, result.Description, result.MatchedText)
		}
	}

	// Check NSFW results
	nsfwResults := filterByCategory(results, "nsfw")
	t.Logf("NSFW findings: %d", len(nsfwResults))

	if len(nsfwResults) > 0 {
		t.Log("Sample NSFW findings:")
		for i, result := range nsfwResults {
			if i >= 5 {
				break
			}
			t.Logf("  - [%s] %s: %s", result.ContentType, result.Description, result.MatchedText)
		}
	}

	// Check PII results
	piiResults := filterByCategory(results, "pii")
	t.Logf("PII findings: %d", len(piiResults))

	// Check slur results
	slurResults := filterByCategory(results, "slur")
	t.Logf("Slur findings: %d", len(slurResults))

	// Verify we got meaningful results
	assert.Greater(t, len(results), 0, "Should have found some analysis results")
}

// TestIntegration_AnalyzeCurrentUser is similar but analyzes the authenticated user
// This demonstrates the "self-analysis" use case where you analyze your own account
func TestIntegration_AnalyzeCurrentUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check for required environment variables
	clientID := os.Getenv("REDDIT_CLIENT_ID")
	clientSecret := os.Getenv("REDDIT_CLIENT_SECRET")
	username := os.Getenv("REDDIT_USERNAME")
	password := os.Getenv("REDDIT_PASSWORD")

	if clientID == "" || clientSecret == "" || username == "" || password == "" {
		t.Skip("Skipping integration test: Reddit credentials not provided")
	}

	// Use the authenticated user's own account
	targetUsername := username

	// Setup test storage
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "self_analysis_test.db")
	store, err := storage.New(dbPath, "self-analysis-passphrase")
	require.NoError(t, err)
	defer store.Close()

	// Setup Reddit client
	creds := reddit.Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
	}

	client, err := reddit.New(creds, store)
	require.NoError(t, err)

	t.Logf("Analyzing your own account: %s", targetUsername)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Scan own account
	err = client.ScanUser(ctx, targetUsername)
	require.NoError(t, err)

	// Run all analyzers
	politicalAnalyzer := NewPoliticalAnalyzer(store)
	err = politicalAnalyzer.AnalyzeUser(ctx, targetUsername)
	require.NoError(t, err)

	nsfwAnalyzer := NewNSFWAnalyzer(store)
	err = nsfwAnalyzer.AnalyzeUser(ctx, targetUsername)
	require.NoError(t, err)

	piiAnalyzer := New(store)
	err = piiAnalyzer.AnalyzeUser(ctx, targetUsername)
	require.NoError(t, err)

	// Get results
	results, err := store.GetAnalysisByUsername(targetUsername)
	require.NoError(t, err)

	t.Logf("Self-analysis complete. Found %d total findings.", len(results))

	// Generate cleanup recommendations
	highSeverity := 0
	mediumSeverity := 0
	lowSeverity := 0

	for _, result := range results {
		switch result.Severity {
		case "high":
			highSeverity++
		case "medium":
			mediumSeverity++
		case "low":
			lowSeverity++
		}
	}

	t.Logf("Severity breakdown:")
	t.Logf("  High:   %d (recommended for deletion)", highSeverity)
	t.Logf("  Medium: %d (consider for deletion)", mediumSeverity)
	t.Logf("  Low:    %d (optional)", lowSeverity)

	if highSeverity > 0 {
		t.Log("\nHigh-severity findings (strongly recommended for deletion):")
		for i, result := range results {
			if result.Severity == "high" && i < 10 {
				t.Logf("  - [%s:%s] %s", result.ContentType, result.ContentID, result.Description)
			}
		}
	}
}
