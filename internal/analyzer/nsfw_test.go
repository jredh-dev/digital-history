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

func TestNSFWAnalyzer_AnalyzeUser_NSFWSubreddits(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data in NSFW subreddits
	testPosts := []*storage.RedditPost{
		{
			ID:          "post1",
			Username:    "testuser",
			Subreddit:   "gonewild",
			Title:       "First post",
			Body:        "Hope you like it",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "post2",
			Username:    "testuser",
			Subreddit:   "nsfw",
			Title:       "Check this out",
			Body:        "",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "realgirls",
			Body:        "Nice!",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "comment2",
			Username:    "testuser",
			PostID:      "post2",
			Subreddit:   "askredditafterdark",
			Body:        "I agree with this",
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

	// Run NSFW analyzer
	analyzer := NewNSFWAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	nsfwResults := filterByCategory(results, "nsfw")

	// Should detect: 2 posts + 2 comments in NSFW subreddits
	assert.GreaterOrEqual(t, len(nsfwResults), 4, "Should detect NSFW subreddit participation")

	// Check that NSFW subreddits were detected
	matchedTexts := extractMatchedText(nsfwResults)
	assert.Contains(t, matchedTexts, "gonewild")
	assert.Contains(t, matchedTexts, "nsfw")
	assert.Contains(t, matchedTexts, "realgirls")
	assert.Contains(t, matchedTexts, "askredditafterdark")
}

func TestNSFWAnalyzer_AnalyzeUser_NSFWKeywords(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create test data with NSFW keywords
	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "pics",
			Body:        "Check out my OnlyFans for more content",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "comment2",
			Username:    "testuser",
			PostID:      "post2",
			Subreddit:   "selfie",
			Body:        "This is definitely a thirst trap lol",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "comment3",
			Username:    "testuser",
			PostID:      "post3",
			Subreddit:   "rateme",
			Body:        "Am I sexy? DM me",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := NewNSFWAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	nsfwResults := filterByCategory(results, "nsfw")

	// Should detect NSFW keywords
	assert.GreaterOrEqual(t, len(nsfwResults), 3, "Should detect NSFW keywords")

	keywordsFound := extractMatchedText(nsfwResults)
	assert.Contains(t, keywordsFound, "onlyfans")
	assert.Contains(t, keywordsFound, "thirst trap")
	assert.Contains(t, keywordsFound, "dm me")
}

func TestNSFWAnalyzer_AnalyzeUser_NoFalsePositives(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Create safe content
	testPosts := []*storage.RedditPost{
		{
			ID:          "post1",
			Username:    "testuser",
			Subreddit:   "cats",
			Title:       "My cute kitten",
			Body:        "She's adorable!",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "aww",
			Body:        "So cute!",
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
	analyzer := NewNSFWAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	nsfwResults := filterByCategory(results, "nsfw")
	assert.Equal(t, 0, len(nsfwResults), "Should not detect NSFW content in safe posts")
}

func TestNSFWAnalyzer_CaseInsensitive(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Test case variations
	testComments := []*storage.RedditComment{
		{
			ID:          "comment1",
			Username:    "testuser",
			PostID:      "post1",
			Subreddit:   "pics",
			Body:        "Check my ONLYFANS",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "comment2",
			Username:    "testuser",
			PostID:      "post2",
			Subreddit:   "selfie",
			Body:        "This is a Thirst Trap",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, comment := range testComments {
		require.NoError(t, store.SaveComment(comment))
	}

	// Run analyzer
	analyzer := NewNSFWAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	nsfwResults := filterByCategory(results, "nsfw")

	// Should detect keywords regardless of case
	assert.GreaterOrEqual(t, len(nsfwResults), 2)
	keywordsFound := extractMatchedText(nsfwResults)
	assert.Contains(t, keywordsFound, "onlyfans")
	assert.Contains(t, keywordsFound, "thirst trap")
}

func TestNSFWAnalyzer_MixedContent(t *testing.T) {
	store, _ := setupTestStorage(t)
	defer store.Close()

	// Mix of NSFW and safe content
	testPosts := []*storage.RedditPost{
		{
			ID:          "post1",
			Username:    "testuser",
			Subreddit:   "gonewild",
			Title:       "NSFW post",
			Body:        "",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "post2",
			Username:    "testuser",
			Subreddit:   "technology",
			Title:       "New tech",
			Body:        "Cool gadget",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
		{
			ID:          "post3",
			Username:    "testuser",
			Subreddit:   "gaming",
			Title:       "Great game",
			Body:        "Check it out",
			CreatedAt:   time.Now(),
			CollectedAt: time.Now(),
		},
	}

	for _, post := range testPosts {
		require.NoError(t, store.SavePost(post))
	}

	// Run analyzer
	analyzer := NewNSFWAnalyzer(store)
	err := analyzer.AnalyzeUser(context.Background(), "testuser")
	require.NoError(t, err)

	// Check results
	results, err := store.GetAnalysisByUsername("testuser")
	require.NoError(t, err)

	nsfwResults := filterByCategory(results, "nsfw")

	// Should only detect the NSFW post
	assert.Equal(t, 1, len(nsfwResults), "Should only detect actual NSFW content")
	assert.Equal(t, "gonewild", nsfwResults[0].MatchedText)
}
