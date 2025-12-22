// Digital History - Digital footprint analysis tool
// Copyright (C) 2025 jredh
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package reddit

import (
	"context"
	"fmt"
	"time"

	"github.com/jredh-dev/digital-history/internal/storage"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

// Client wraps the Reddit API client
type Client struct {
	reddit  *reddit.Client
	storage *storage.Storage
}

// Credentials holds Reddit API credentials
type Credentials struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

// New creates a new Reddit client
func New(creds Credentials, store *storage.Storage) (*Client, error) {
	// Create Reddit client with password authentication
	client, err := reddit.NewClient(
		reddit.Credentials{
			ID:       creds.ClientID,
			Secret:   creds.ClientSecret,
			Username: creds.Username,
			Password: creds.Password,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reddit client: %w", err)
	}

	return &Client{
		reddit:  client,
		storage: store,
	}, nil
}

// ScanUser collects all posts and comments for a Reddit user
func (c *Client) ScanUser(ctx context.Context, username string) error {
	fmt.Printf("Scanning Reddit user: %s\n", username)

	// Collect posts
	fmt.Println("Collecting posts...")
	if err := c.collectPosts(ctx, username); err != nil {
		return fmt.Errorf("failed to collect posts: %w", err)
	}

	// Collect comments
	fmt.Println("Collecting comments...")
	if err := c.collectComments(ctx, username); err != nil {
		return fmt.Errorf("failed to collect comments: %w", err)
	}

	fmt.Println("Scan complete!")
	return nil
}

func (c *Client) collectPosts(ctx context.Context, username string) error {
	var after string
	postCount := 0

	for {
		// Fetch posts (100 at a time, Reddit's max)
		posts, resp, err := c.reddit.User.PostsOf(ctx, username, &reddit.ListUserOverviewOptions{
			ListOptions: reddit.ListOptions{
				Limit: 100,
				After: after,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to fetch posts: %w", err)
		}

		// Save posts to storage
		for _, post := range posts {
			p := &storage.RedditPost{
				ID:          post.ID,
				Username:    username,
				Subreddit:   post.SubredditName,
				Title:       post.Title,
				Body:        post.Body,
				URL:         post.URL,
				Score:       post.Score,
				CreatedAt:   post.Created.Time,
				CollectedAt: time.Now(),
			}

			if err := c.storage.SavePost(p); err != nil {
				return fmt.Errorf("failed to save post %s: %w", post.ID, err)
			}
			postCount++
		}

		fmt.Printf("  Collected %d posts...\n", postCount)

		// Check if there are more posts
		if resp.After == "" {
			break
		}
		after = resp.After

		// Rate limiting - be respectful
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("Total posts collected: %d\n", postCount)
	return nil
}

func (c *Client) collectComments(ctx context.Context, username string) error {
	var after string
	commentCount := 0

	for {
		// Fetch comments (100 at a time, Reddit's max)
		comments, resp, err := c.reddit.User.CommentsOf(ctx, username, &reddit.ListUserOverviewOptions{
			ListOptions: reddit.ListOptions{
				Limit: 100,
				After: after,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to fetch comments: %w", err)
		}

		// Save comments to storage
		for _, comment := range comments {
			cm := &storage.RedditComment{
				ID:          comment.ID,
				Username:    username,
				PostID:      comment.PostID,
				Subreddit:   comment.SubredditName,
				Body:        comment.Body,
				Score:       comment.Score,
				CreatedAt:   comment.Created.Time,
				CollectedAt: time.Now(),
			}

			if err := c.storage.SaveComment(cm); err != nil {
				return fmt.Errorf("failed to save comment %s: %w", comment.ID, err)
			}
			commentCount++
		}

		fmt.Printf("  Collected %d comments...\n", commentCount)

		// Check if there are more comments
		if resp.After == "" {
			break
		}
		after = resp.After

		// Rate limiting - be respectful
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("Total comments collected: %d\n", commentCount)
	return nil
}
