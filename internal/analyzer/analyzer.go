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
	"regexp"
	"strings"
	"time"

	"github.com/jredh-dev/digital-history/internal/storage"
)

// Analyzer performs content analysis
type Analyzer struct {
	storage  *storage.Storage
	slurs    []string
	patterns map[string]*regexp.Regexp
}

// New creates a new content analyzer
func New(store *storage.Storage) *Analyzer {
	// Common slurs and problematic terms (abbreviated for safety)
	// In production, load from encrypted config file
	slurs := []string{
		// Racial slurs
		"n-word", "f-word", "r-word", // placeholders - actual words should be in config

		// Homophobic/transphobic slurs
		"f*g", "tr*nny", "d*ke",

		// Ableist slurs
		"ret*rd", "sp*z",
	}

	// Compile regex patterns for PII detection
	patterns := map[string]*regexp.Regexp{
		"email":       regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
		"phone":       regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
		"ssn":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		"credit_card": regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),
		"address":     regexp.MustCompile(`\b\d+\s+[A-Za-z0-9\s,]+\s+(Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Lane|Ln|Drive|Dr|Court|Ct)\b`),
		"ip_address":  regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
	}

	return &Analyzer{
		storage:  store,
		slurs:    slurs,
		patterns: patterns,
	}
}

// AnalyzeUser performs comprehensive analysis on a user's content
func (a *Analyzer) AnalyzeUser(ctx context.Context, username string) error {
	fmt.Printf("Analyzing content for user: %s\n", username)

	// Analyze posts
	posts, err := a.storage.GetPostsByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	fmt.Printf("Analyzing %d posts...\n", len(posts))
	for _, post := range posts {
		// Analyze title
		if err := a.analyzeText(username, "post", post.ID, post.Title); err != nil {
			return err
		}
		// Analyze body
		if err := a.analyzeText(username, "post", post.ID, post.Body); err != nil {
			return err
		}
	}

	// Analyze comments
	comments, err := a.storage.GetCommentsByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to get comments: %w", err)
	}

	fmt.Printf("Analyzing %d comments...\n", len(comments))
	for _, comment := range comments {
		if err := a.analyzeText(username, "comment", comment.ID, comment.Body); err != nil {
			return err
		}
	}

	fmt.Println("Analysis complete!")
	return nil
}

func (a *Analyzer) analyzeText(username, contentType, contentID, text string) error {
	text = strings.ToLower(text)

	// Check for slurs
	for _, slur := range a.slurs {
		if strings.Contains(text, strings.ToLower(slur)) {
			if err := a.saveResult(username, contentType, contentID, "slur", "high",
				"Potentially offensive language detected", slur); err != nil {
				return err
			}
		}
	}

	// Check for PII
	for category, pattern := range a.patterns {
		if matches := pattern.FindAllString(text, -1); len(matches) > 0 {
			severity := "high"
			if category == "ip_address" {
				severity = "medium"
			}

			for _, match := range matches {
				description := fmt.Sprintf("%s detected in content", category)
				if err := a.saveResult(username, contentType, contentID, "pii", severity,
					description, match); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (a *Analyzer) saveResult(username, contentType, contentID, category, severity, description, matchedText string) error {
	// Generate unique ID from content
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s-%s", username, contentID, category, matchedText)))
	id := fmt.Sprintf("%x", hash[:8])

	result := &storage.AnalysisResult{
		ID:          id,
		Username:    username,
		ContentType: contentType,
		ContentID:   contentID,
		Category:    category,
		Severity:    severity,
		Description: description,
		MatchedText: matchedText,
		AnalyzedAt:  time.Now(),
	}

	return a.storage.SaveAnalysis(result)
}

// GetSummary retrieves analysis summary for a user
func (a *Analyzer) GetSummary(username string) (map[string]int, error) {
	results, err := a.storage.GetAnalysisByUsername(username)
	if err != nil {
		return nil, err
	}

	summary := make(map[string]int)
	for _, result := range results {
		key := fmt.Sprintf("%s_%s", result.Category, result.Severity)
		summary[key]++
	}

	return summary, nil
}
