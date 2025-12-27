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
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/jredh-dev/digital-history/internal/storage"
)

// NSFWAnalyzer detects NSFW content in posts/comments
type NSFWAnalyzer struct {
	storage    *storage.Storage
	subreddits map[string]bool
	keywords   []string
}

// NewNSFWAnalyzer creates a new NSFW content analyzer
func NewNSFWAnalyzer(store *storage.Storage) *NSFWAnalyzer {
	// Known NSFW subreddits (abbreviated list - should be configurable)
	subreddits := map[string]bool{
		// Adult content
		"nsfw":                 true,
		"gonewild":             true,
		"realgirls":            true,
		"nsfw_gifs":            true,
		"holdthemoan":          true,
		"celebnsfw":            true,
		"amateur":              true,
		"milf":                 true,
		"asiansgonewild":       true,
		"bigboobsgw":           true,
		"gwcouples":            true,
		"petitegonewild":       true,
		"adorableporn":         true,
		"porninfifteenseconds": true,

		// Adult discussion
		"sex":                true,
		"askredditafterdark": true,
		"dirtypenpals":       true,
		"sexstories":         true,

		// Dating/hookup
		"tinder":   true,
		"bumble":   true,
		"dirtyr4r": true,
		"r4r":      true,

		// Thirst trap adjacent
		"prettygirls":    true,
		"hot":            true,
		"sexybutnotporn": true,
		"models":         true,
		"fitgirls":       true,
		"workoutgirls":   true,
	}

	// Keywords that might indicate thirst trap content
	keywords := []string{
		"onlyfans", "of link", "dm me", "check my profile",
		"thirst trap", "attention seeking", "validation",
		"hot or not", "rate me", "am i sexy",
	}

	return &NSFWAnalyzer{
		storage:    store,
		subreddits: subreddits,
		keywords:   keywords,
	}
}

// AnalyzeUser performs NSFW content analysis on a user's content
func (na *NSFWAnalyzer) AnalyzeUser(ctx context.Context, username string) error {
	// Analyze posts
	posts, err := na.storage.GetPostsByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	for _, post := range posts {
		// Check if posted in NSFW subreddit
		if na.isNSFWSubreddit(post.Subreddit) {
			if err := na.saveResult(username, "post", post.ID, "nsfw_subreddit",
				fmt.Sprintf("Posted in NSFW subreddit: r/%s", post.Subreddit),
				post.Subreddit); err != nil {
				return err
			}
		}

		// Check post content for NSFW keywords
		if err := na.analyzeText(username, "post", post.ID, post.Title+" "+post.Body); err != nil {
			return err
		}
	}

	// Analyze comments
	comments, err := na.storage.GetCommentsByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to get comments: %w", err)
	}

	for _, comment := range comments {
		// Check if commented in NSFW subreddit
		if na.isNSFWSubreddit(comment.Subreddit) {
			if err := na.saveResult(username, "comment", comment.ID, "nsfw_subreddit",
				fmt.Sprintf("Commented in NSFW subreddit: r/%s", comment.Subreddit),
				comment.Subreddit); err != nil {
				return err
			}
		}

		// Check comment content for NSFW keywords
		if err := na.analyzeText(username, "comment", comment.ID, comment.Body); err != nil {
			return err
		}
	}

	return nil
}

func (na *NSFWAnalyzer) isNSFWSubreddit(subreddit string) bool {
	return na.subreddits[strings.ToLower(subreddit)]
}

func (na *NSFWAnalyzer) analyzeText(username, contentType, contentID, text string) error {
	textLower := strings.ToLower(text)

	// Check for NSFW keywords
	for _, keyword := range na.keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			description := fmt.Sprintf("NSFW keyword detected: %s", keyword)
			if err := na.saveResult(username, contentType, contentID, "nsfw_keyword",
				description, keyword); err != nil {
				return err
			}
		}
	}

	return nil
}

func (na *NSFWAnalyzer) saveResult(username, contentType, contentID, subcategory, description, matchedText string) error {
	// Generate unique ID
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s-%s", username, contentID, subcategory, matchedText)))
	id := fmt.Sprintf("%x", hash[:8])

	result := &storage.AnalysisResult{
		ID:          id,
		Username:    username,
		ContentType: contentType,
		ContentID:   contentID,
		Category:    "nsfw",
		Severity:    "medium", // NSFW content is medium severity for scrubbing
		Description: description,
		MatchedText: matchedText,
		AnalyzedAt:  time.Now(),
	}

	return na.storage.SaveAnalysis(result)
}
