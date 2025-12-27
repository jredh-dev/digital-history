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

// PoliticalAnalyzer detects political content in posts/comments
type PoliticalAnalyzer struct {
	storage     *storage.Storage
	keywords    []string
	subreddits  map[string]bool
	politicians []string
}

// NewPoliticalAnalyzer creates a new political content analyzer
func NewPoliticalAnalyzer(store *storage.Storage) *PoliticalAnalyzer {
	// Political keywords to detect
	keywords := []string{
		// Government/institutions
		"government", "congress", "senate", "white house", "administration",
		"legislation", "bill", "law", "policy", "regulation",

		// Political parties
		"democrat", "republican", "liberal", "conservative", "progressive",
		"left-wing", "right-wing", "libertarian", "socialist",

		// Politicians (common names - should be configurable)
		"biden", "trump", "harris", "pence", "pelosi", "mcconnell",
		"aoc", "bernie", "sanders", "desantis", "obama",

		// Political topics
		"immigration", "border", "healthcare", "abortion", "gun control",
		"climate policy", "tax reform", "voting rights", "election",
		"filibuster", "supreme court", "impeachment",

		// Controversial terms
		"fascist", "communist", "nazi", "antifa", "deep state",
		"fake news", "mainstream media", "propaganda",
	}

	// Political subreddits
	subreddits := map[string]bool{
		"politics":              true,
		"politicaldiscussion":   true,
		"politicalhumor":        true,
		"conservative":          true,
		"liberal":               true,
		"libertarian":           true,
		"socialism":             true,
		"communism":             true,
		"progressive":           true,
		"moderatepolitics":      true,
		"neutralpolitics":       true,
		"democrat":              true,
		"republican":            true,
		"the_donald":            true,
		"sandersforpresident":   true,
		"chapotraphouse":        true,
		"worldpolitics":         true,
		"geopolitics":           true,
		"politicalcompassmemes": true,
	}

	return &PoliticalAnalyzer{
		storage:    store,
		keywords:   keywords,
		subreddits: subreddits,
	}
}

// AnalyzeUser performs political content analysis on a user's content
func (pa *PoliticalAnalyzer) AnalyzeUser(ctx context.Context, username string) error {
	// Analyze posts
	posts, err := pa.storage.GetPostsByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	for _, post := range posts {
		// Check if posted in political subreddit
		if pa.isPoliticalSubreddit(post.Subreddit) {
			if err := pa.saveResult(username, "post", post.ID, "political_subreddit",
				fmt.Sprintf("Posted in political subreddit: r/%s", post.Subreddit),
				post.Subreddit); err != nil {
				return err
			}
		}

		// Check post content for political keywords
		if err := pa.analyzeText(username, "post", post.ID, post.Title+" "+post.Body); err != nil {
			return err
		}
	}

	// Analyze comments
	comments, err := pa.storage.GetCommentsByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to get comments: %w", err)
	}

	for _, comment := range comments {
		// Check if commented in political subreddit
		if pa.isPoliticalSubreddit(comment.Subreddit) {
			if err := pa.saveResult(username, "comment", comment.ID, "political_subreddit",
				fmt.Sprintf("Commented in political subreddit: r/%s", comment.Subreddit),
				comment.Subreddit); err != nil {
				return err
			}
		}

		// Check comment content for political keywords
		if err := pa.analyzeText(username, "comment", comment.ID, comment.Body); err != nil {
			return err
		}
	}

	return nil
}

func (pa *PoliticalAnalyzer) isPoliticalSubreddit(subreddit string) bool {
	return pa.subreddits[strings.ToLower(subreddit)]
}

func (pa *PoliticalAnalyzer) analyzeText(username, contentType, contentID, text string) error {
	textLower := strings.ToLower(text)

	// Check for political keywords
	for _, keyword := range pa.keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			description := fmt.Sprintf("Political keyword detected: %s", keyword)
			if err := pa.saveResult(username, contentType, contentID, "political_keyword",
				description, keyword); err != nil {
				return err
			}
		}
	}

	return nil
}

func (pa *PoliticalAnalyzer) saveResult(username, contentType, contentID, subcategory, description, matchedText string) error {
	// Generate unique ID
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s-%s", username, contentID, subcategory, matchedText)))
	id := fmt.Sprintf("%x", hash[:8])

	result := &storage.AnalysisResult{
		ID:          id,
		Username:    username,
		ContentType: contentType,
		ContentID:   contentID,
		Category:    "political",
		Severity:    "medium", // Political content is medium severity for scrubbing purposes
		Description: description,
		MatchedText: matchedText,
		AnalyzedAt:  time.Now(),
	}

	return pa.storage.SaveAnalysis(result)
}
