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
	"testing"
	"time"

	"github.com/jredh-dev/digital-history/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzer_AnalyzeUser_SlurDetection(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data with slurs (using placeholders)
	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "test",
			Body:        "This contains the r-word which is offensive",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := New(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	slurResults := filterByCategory(results, "slur")
	assert.Greater(t, len(slurResults), 0, "Should detect slur")
	assert.Equal(t, "high", slurResults[0].Severity)
}

func TestAnalyzer_AnalyzeUser_EmailDetection(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data with email
	testPosts := []*storage.RedditPost{
		{
			ID:          "post1",
			Username:    "testuser",
			Subreddit:   "test",
			Title:       "Contact me",
			Body:        "Email me at john.doe@example.com for details",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, post := range testPosts {
		require.NoError(t, store.SavePost(post))
	}

	// Run analyzer
	analyzer := New(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	piiResults := filterByCategory(results, "pii")
	assert.Greater(t, len(piiResults), 0, "Should detect email")

	// Find email result
	var emailResult *storage.AnalysisResult
	for _, r := range piiResults {
		if r.MatchedText == "john.doe@example.com" {
			emailResult = r
			break
		}
	}
	require.NotNil(t, emailResult, "Should find email result")
	assert.Equal(t, "high", emailResult.Severity)
	assert.Contains(t, emailResult.Description, "email")
}

func TestAnalyzer_AnalyzeUser_PhoneDetection(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data with phone number
	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "test",
			Body:        "Call me at 555-123-4567",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := New(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	piiResults := filterByCategory(results, "pii")
	assert.Greater(t, len(piiResults), 0, "Should detect phone number")
}

func TestAnalyzer_AnalyzeUser_SSNDetection(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data with SSN
	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "test",
			Body:        "My SSN is 123-45-6789 (fake)",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := New(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	piiResults := filterByCategory(results, "pii")
	assert.Greater(t, len(piiResults), 0, "Should detect SSN")
	assert.Equal(t, "high", piiResults[0].Severity)
}

func TestAnalyzer_AnalyzeUser_IPAddressDetection(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data with IP address
	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "test",
			Body:        "Server is at 192.168.1.100",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := New(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	piiResults := filterByCategory(results, "pii")
	assert.Greater(t, len(piiResults), 0, "Should detect IP address")
	assert.Equal(t, "medium", piiResults[0].Severity, "IP addresses should be medium severity")
}

func TestAnalyzer_GetSummary(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data
	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "test",
			Body:        "Email me at test@example.com or call 555-1234",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := New(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Get summary
	summary, err := analyzer.GetSummary("testuser")
	require.NoError(t, err)

	// Should have PII findings
	totalPII := summary["pii_high"] + summary["pii_medium"]
	assert.Greater(t, totalPII, 0, "Should have PII findings in summary")
}

func TestAnalyzer_AnalyzeUser_NoContent(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Run analyzer on user with no content
	analyzer := New(store)
	err := analyzer.AnalyzeUser(context.Background(), "nonexistent")
	require.NoError(t, err, "Should handle user with no content gracefully")

	// Check results
	results, err := store.GetAnalysisByUsername("nonexistent")
	require.NoError(t, err)
	assert.Equal(t, 0, len(results), "Should have no results for nonexistent user")
}

func TestAnalyzer_AnalyzeUser_MultipleFindings(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data with multiple issues
	testPosts := []*storage.RedditPost{
		{
			ID:          "post1",
			Username:    "testuser",
			Subreddit:   "test",
			Title:       "Multiple issues",
			Body:        "Email: test@example.com, Phone: 555-123-4567, Address: 123 Main Street",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, post := range testPosts {
		require.NoError(t, store.SavePost(post))
	}

	// Run analyzer
	analyzer := New(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results - should find multiple PII items
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	piiResults := filterByCategory(results, "pii")
	assert.GreaterOrEqual(t, len(piiResults), 2, "Should detect at least email and phone")
}
