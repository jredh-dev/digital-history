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
	"path/filepath"
	"testing"
	"time"

	"github.com/jredh-dev/digital-history/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStorage(t *testing.T) (*storage.Storage, string) {
	t.Helper()

	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := storage.New(dbPath, "test-passphrase")
	require.NoError(t, err, "Failed to create test storage")

	return store, dbPath
}

func TestPoliticalAnalyzer_AnalyzeUser_PoliticalKeywords(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data with political keywords
	testPosts := []*storage.RedditPost{
		{
			ID:          "post1",
			Username:    "testuser",
			Subreddit:   "askreddit",
			Title:       "What do you think about Biden's new policy?",
			Body:        "I'm curious about opinions on the healthcare legislation.",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "post2",
			Username:    "testuser",
			Subreddit:   "technology",
			Title:       "New tech gadget",
			Body:        "Check out this cool device",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "news",
			Body:        "The Trump administration did things differently.",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "comment2",
			Username:    "testuser",
			PostID:      "post2",
			Subreddit:   "technology",
			Body:        "This is awesome!",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	// Save test data
	for _, post := range testPosts {
		require.NoError(t, store.SavePost(post))
	}
	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run political analyzer
	analyzer := NewPoliticalAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	// Filter for political category
	politicalResults := filterByCategory(results, "political")

	// Should detect: "biden", "policy", "healthcare", "legislation", "trump", "administration"
	assert.GreaterOrEqual(t, len(politicalResults), 4, "Should detect multiple political keywords")

	// Verify specific detections
	keywordsFound := extractMatchedText(politicalResults)
	assert.Contains(t, keywordsFound, "biden")
	assert.Contains(t, keywordsFound, "trump")
	assert.Contains(t, keywordsFound, "healthcare")
}

func TestPoliticalAnalyzer_AnalyzeUser_PoliticalSubreddits(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data in political subreddits
	testPosts := []*storage.RedditPost{
		{
			ID:          "post1",
			Username:    "testuser",
			Subreddit:   "politics",
			Title:       "Discussion about voting",
			Body:        "Let's talk about voting rights.",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "post2",
			Username:    "testuser",
			Subreddit:   "conservative",
			Title:       "Conservative viewpoint",
			Body:        "Here's my take.",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "politicaldiscussion",
			Body:        "I agree with this point.",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	// Save test data
	for _, post := range testPosts {
		require.NoError(t, store.SavePost(post))
	}
	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := NewPoliticalAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	politicalResults := filterByCategory(results, "political")

	// Should detect: 2 posts + 1 comment in political subreddits, plus keywords
	assert.GreaterOrEqual(t, len(politicalResults), 3, "Should detect political subreddit participation")

	// Check that political subreddits were detected
	matchedTexts := extractMatchedText(politicalResults)
	assert.Contains(t, matchedTexts, "politics")
	assert.Contains(t, matchedTexts, "conservative")
	assert.Contains(t, matchedTexts, "politicaldiscussion")
}

func TestPoliticalAnalyzer_AnalyzeUser_NoFalsePositives(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create non-political content
	testPosts := []*storage.RedditPost{
		{
			ID:          "post1",
			Username:    "testuser",
			Subreddit:   "aww",
			Title:       "Look at my cute cat",
			Body:        "She's so adorable!",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "gaming",
			Body:        "This game is amazing!",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	// Save test data
	for _, post := range testPosts {
		require.NoError(t, store.SavePost(post))
	}
	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := NewPoliticalAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	politicalResults := filterByCategory(results, "political")
	assert.Equal(t, 0, len(politicalResults), "Should not detect political content in non-political posts")
}

func TestPoliticalAnalyzer_CaseInsensitive(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Test case variations
	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "news",
			Body:        "BIDEN signed the bill",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "comment2",
			Username:    "testuser",
			PostID:      "post2",
			Subreddit:   "news",
			Body:        "Trump made a statement",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "comment3",
			Username:    "testuser",
			PostID:      "post3",
			Subreddit:   "news",
			Body:        "The Government announced today",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := NewPoliticalAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	politicalResults := filterByCategory(results, "political")

	// Should detect keywords regardless of case
	assert.GreaterOrEqual(t, len(politicalResults), 3)
	keywordsFound := extractMatchedText(politicalResults)
	assert.Contains(t, keywordsFound, "biden")
	assert.Contains(t, keywordsFound, "trump")
	assert.Contains(t, keywordsFound, "government")
}

// Helper functions

func filterByCategory(results []*storage.AnalysisResult, category string) []*storage.AnalysisResult {
	var filtered []*storage.AnalysisResult
	for _, r := range results {
		if r.Category == category {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func extractMatchedText(results []*storage.AnalysisResult) []string {
	var texts []string
	for _, r := range results {
		texts = append(texts, r.MatchedText)
	}
	return texts
}
