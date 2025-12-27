// Digital History - Digital footprint analysis tool
// Copyright (C) 2025 jredh
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/pbkdf2"
)

// Storage handles encrypted data persistence
type Storage struct {
	db         *sql.DB
	aead       cipher.AEAD
	passphrase string
}

// RedditPost represents a Reddit post
type RedditPost struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Subreddit   string    `json:"subreddit"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	URL         string    `json:"url"`
	Score       int       `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
	CollectedAt time.Time `json:"collected_at"`
}

// RedditComment represents a Reddit comment
type RedditComment struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	PostID      string    `json:"post_id"`
	Subreddit   string    `json:"subreddit"`
	Body        string    `json:"body"`
	Score       int       `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
	CollectedAt time.Time `json:"collected_at"`
}

// AnalysisResult represents content analysis findings
type AnalysisResult struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	ContentType string    `json:"content_type"` // "post" or "comment"
	ContentID   string    `json:"content_id"`
	Category    string    `json:"category"` // "slur", "pii", "location", etc.
	Severity    string    `json:"severity"` // "low", "medium", "high"
	Description string    `json:"description"`
	MatchedText string    `json:"matched_text"`
	AnalyzedAt  time.Time `json:"analyzed_at"`
}

// New creates a new encrypted storage
func New(dbPath, passphrase string) (*Storage, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Derive encryption key from passphrase
	salt := []byte("digital-history-salt-v1") // In production, store salt separately
	key := pbkdf2.Key([]byte(passphrase), salt, 100000, 32, sha256.New)

	// Create AES-GCM cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	s := &Storage{
		db:         db,
		aead:       aead,
		passphrase: passphrase,
	}

	// Initialize schema
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return s, nil
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS reddit_posts (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		subreddit TEXT NOT NULL,
		encrypted_data BLOB NOT NULL,
		created_at TIMESTAMP NOT NULL,
		collected_at TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_posts_username ON reddit_posts(username);
	CREATE INDEX IF NOT EXISTS idx_posts_subreddit ON reddit_posts(subreddit);
	CREATE INDEX IF NOT EXISTS idx_posts_created_at ON reddit_posts(created_at);

	CREATE TABLE IF NOT EXISTS reddit_comments (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		post_id TEXT NOT NULL,
		subreddit TEXT NOT NULL,
		encrypted_data BLOB NOT NULL,
		created_at TIMESTAMP NOT NULL,
		collected_at TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_comments_username ON reddit_comments(username);
	CREATE INDEX IF NOT EXISTS idx_comments_subreddit ON reddit_comments(subreddit);
	CREATE INDEX IF NOT EXISTS idx_comments_created_at ON reddit_comments(created_at);

	CREATE TABLE IF NOT EXISTS analysis_results (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		content_type TEXT NOT NULL,
		content_id TEXT NOT NULL,
		category TEXT NOT NULL,
		severity TEXT NOT NULL,
		encrypted_data BLOB NOT NULL,
		analyzed_at TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_analysis_username ON analysis_results(username);
	CREATE INDEX IF NOT EXISTS idx_analysis_category ON analysis_results(category);
	CREATE INDEX IF NOT EXISTS idx_analysis_severity ON analysis_results(severity);
	`

	_, err := s.db.Exec(schema)
	return err
}

// encrypt encrypts data using AES-GCM
func (s *Storage) encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := s.aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func (s *Storage) decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < s.aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:s.aead.NonceSize()]
	ciphertext = ciphertext[s.aead.NonceSize():]

	plaintext, err := s.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// SavePost stores a Reddit post with encryption
func (s *Storage) SavePost(post *RedditPost) error {
	// Encrypt sensitive fields
	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %w", err)
	}

	encrypted, err := s.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt post: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO reddit_posts 
		(id, username, subreddit, encrypted_data, created_at, collected_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		post.ID, post.Username, post.Subreddit, encrypted, post.CreatedAt, post.CollectedAt,
	)

	return err
}

// SaveComment stores a Reddit comment with encryption
func (s *Storage) SaveComment(comment *RedditComment) error {
	data, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	encrypted, err := s.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt comment: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO reddit_comments 
		(id, username, post_id, subreddit, encrypted_data, created_at, collected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		comment.ID, comment.Username, comment.PostID, comment.Subreddit, encrypted, comment.CreatedAt, comment.CollectedAt,
	)

	return err
}

// SaveAnalysis stores an analysis result with encryption
func (s *Storage) SaveAnalysis(result *AnalysisResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	encrypted, err := s.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt analysis: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO analysis_results 
		(id, username, content_type, content_id, category, severity, encrypted_data, analyzed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		result.ID, result.Username, result.ContentType, result.ContentID, result.Category, result.Severity, encrypted, result.AnalyzedAt,
	)

	return err
}

// GetPostsByUsername retrieves all posts for a username
func (s *Storage) GetPostsByUsername(username string) ([]*RedditPost, error) {
	rows, err := s.db.Query(`
		SELECT encrypted_data FROM reddit_posts 
		WHERE username = ? ORDER BY created_at DESC`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*RedditPost
	for rows.Next() {
		var encrypted []byte
		if err := rows.Scan(&encrypted); err != nil {
			return nil, err
		}

		decrypted, err := s.decrypt(encrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt post: %w", err)
		}

		var post RedditPost
		if err := json.Unmarshal(decrypted, &post); err != nil {
			return nil, fmt.Errorf("failed to unmarshal post: %w", err)
		}

		posts = append(posts, &post)
	}

	return posts, nil
}

// GetCommentsByUsername retrieves all comments for a username
func (s *Storage) GetCommentsByUsername(username string) ([]*RedditComment, error) {
	rows, err := s.db.Query(`
		SELECT encrypted_data FROM reddit_comments 
		WHERE username = ? ORDER BY created_at DESC`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*RedditComment
	for rows.Next() {
		var encrypted []byte
		if err := rows.Scan(&encrypted); err != nil {
			return nil, err
		}

		decrypted, err := s.decrypt(encrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt comment: %w", err)
		}

		var comment RedditComment
		if err := json.Unmarshal(decrypted, &comment); err != nil {
			return nil, fmt.Errorf("failed to unmarshal comment: %w", err)
		}

		comments = append(comments, &comment)
	}

	return comments, nil
}

// GetAnalysisByUsername retrieves all analysis results for a username
func (s *Storage) GetAnalysisByUsername(username string) ([]*AnalysisResult, error) {
	rows, err := s.db.Query(`
		SELECT encrypted_data FROM analysis_results 
		WHERE username = ? ORDER BY analyzed_at DESC`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*AnalysisResult
	for rows.Next() {
		var encrypted []byte
		if err := rows.Scan(&encrypted); err != nil {
			return nil, err
		}

		decrypted, err := s.decrypt(encrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt analysis: %w", err)
		}

		var result AnalysisResult
		if err := json.Unmarshal(decrypted, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal analysis: %w", err)
		}

		results = append(results, &result)
	}

	return results, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}
