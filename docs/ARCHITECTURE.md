# Architecture

Technical architecture documentation for digital-history.

## Overview

```
┌─────────────────┐
│   CLI (Cobra)   │  User interface
└────────┬────────┘
         │
    ┌────┴────┐
    │ Config  │  Configuration management
    └────┬────┘
         │
    ┌────┴─────────────────────────────┐
    │                                   │
┌───┴──────┐                    ┌──────┴──────┐
│ Platform │ ← Data Collection  │  Analyzer   │  Content analysis
│ Clients  │                    │             │
└───┬──────┘                    └──────┬──────┘
    │                                   │
    │         ┌─────────────────────────┘
    │         │
┌───┴─────────┴───┐
│     Storage     │  Encrypted persistence
│  (SQLite+AES)   │
└─────────────────┘
```

## Components

### 1. CLI Layer (`cmd/digital-history/`)

**Purpose**: User-facing command-line interface

**Technology**: Cobra framework

**Commands**:
- `scan` - Collect data from platforms
- `list` - View summary of collected data
- `export` - Export data to JSON
- `config` - Manage configuration

**Flow**:
1. Parse command-line arguments
2. Load configuration from `~/.digital-history.yaml`
3. Validate credentials
4. Delegate to appropriate component
5. Display results to user

### 2. Configuration (`internal/config/`)

**Purpose**: Centralized configuration management

**Technology**: Viper

**Responsibilities**:
- Load config from YAML file
- Provide defaults
- Validate credentials
- Save configuration changes

**Configuration Structure**:
```go
type Config struct {
    Database DatabaseConfig  // Path and encryption passphrase
    Reddit   RedditConfig    // API credentials
    Analyzer AnalyzerConfig  // Analysis settings
}
```

**File Location**: `~/.digital-history.yaml`

### 3. Platform Clients (`internal/platform/`)

**Purpose**: Interface with external platforms (Reddit, Twitter, etc.)

**Current Implementation**: Reddit only

#### Reddit Client (`internal/platform/reddit/`)

**Technology**: `vartanbeno/go-reddit/v2`

**Responsibilities**:
- OAuth2 authentication with Reddit API
- Fetch user posts (with pagination)
- Fetch user comments (with pagination)
- Rate limiting (2 second delay between requests)
- Save to storage

**Authentication Flow**:
```
1. Create OAuth2 client with credentials
2. Reddit API validates and issues token
3. Token used for all subsequent requests
4. Token refresh handled automatically
```

**Data Collection Flow**:
```
1. Fetch batch of 100 posts/comments
2. Encrypt and store each item
3. Check for more pages (pagination)
4. Wait 2 seconds (rate limiting)
5. Repeat until all data collected
```

**Future Platforms**:
- Twitter/X client
- GitHub client
- Facebook client
- Stack Overflow client

### 4. Storage Layer (`internal/storage/`)

**Purpose**: Encrypted data persistence

**Technology**: SQLite3 with AES-256-GCM encryption

**Architecture**:

```
┌──────────────────────────────────┐
│        Application Layer         │
└────────────┬─────────────────────┘
             │
┌────────────┴─────────────────────┐
│       Encryption Layer           │
│  (AES-256-GCM + PBKDF2)         │
└────────────┬─────────────────────┘
             │
┌────────────┴─────────────────────┐
│        SQLite Database           │
│     (~/.digital-history/data.db) │
└──────────────────────────────────┘
```

#### Encryption Details

**Algorithm**: AES-256-GCM (Galois/Counter Mode)
- Provides both confidentiality and authenticity
- Detects tampering
- Each record encrypted with unique nonce

**Key Derivation**: PBKDF2
- Hash: SHA-256
- Iterations: 100,000
- Salt: `digital-history-salt-v1` (hardcoded for MVP)
- Output: 32-byte key for AES-256

**What's Encrypted**:
- Post/comment content (title, body, URL)
- Analysis results (matched text, descriptions)

**What's Not Encrypted** (for querying):
- IDs (post_id, comment_id)
- Usernames
- Timestamps
- Subreddit names
- Severity levels
- Categories

#### Database Schema

```sql
-- Reddit posts
CREATE TABLE reddit_posts (
    id TEXT PRIMARY KEY,              -- Reddit post ID
    username TEXT NOT NULL,           -- Reddit username
    subreddit TEXT NOT NULL,          -- Subreddit name
    encrypted_data BLOB NOT NULL,     -- Encrypted JSON payload
    created_at TIMESTAMP NOT NULL,    -- Post creation time
    collected_at TIMESTAMP NOT NULL,  -- Collection timestamp
    INDEX idx_username (username),
    INDEX idx_subreddit (subreddit),
    INDEX idx_created_at (created_at)
);

-- Reddit comments
CREATE TABLE reddit_comments (
    id TEXT PRIMARY KEY,              -- Reddit comment ID
    username TEXT NOT NULL,           -- Reddit username
    post_id TEXT NOT NULL,            -- Parent post ID
    subreddit TEXT NOT NULL,          -- Subreddit name
    encrypted_data BLOB NOT NULL,     -- Encrypted JSON payload
    created_at TIMESTAMP NOT NULL,    -- Comment creation time
    collected_at TIMESTAMP NOT NULL,  -- Collection timestamp
    INDEX idx_username (username),
    INDEX idx_subreddit (subreddit),
    INDEX idx_created_at (created_at)
);

-- Analysis results
CREATE TABLE analysis_results (
    id TEXT PRIMARY KEY,              -- SHA-256 hash of content
    username TEXT NOT NULL,           -- Reddit username
    content_type TEXT NOT NULL,       -- "post" or "comment"
    content_id TEXT NOT NULL,         -- Reference to post/comment
    category TEXT NOT NULL,           -- "slur", "pii", etc.
    severity TEXT NOT NULL,           -- "high", "medium", "low"
    encrypted_data BLOB NOT NULL,     -- Encrypted JSON payload
    analyzed_at TIMESTAMP NOT NULL,   -- Analysis timestamp
    INDEX idx_username (username),
    INDEX idx_category (category),
    INDEX idx_severity (severity)
);
```

#### Encrypted Payload Structure

**Posts**:
```json
{
  "id": "abc123",
  "username": "jredh",
  "subreddit": "golang",
  "title": "How to encrypt data at rest?",
  "body": "I'm building an app...",
  "url": "https://reddit.com/...",
  "score": 42,
  "created_at": "2025-01-01T12:00:00Z",
  "collected_at": "2025-01-15T14:30:00Z"
}
```

**Comments**:
```json
{
  "id": "def456",
  "username": "jredh",
  "post_id": "abc123",
  "subreddit": "golang",
  "body": "You should use AES-256-GCM...",
  "score": 15,
  "created_at": "2025-01-01T13:00:00Z",
  "collected_at": "2025-01-15T14:30:00Z"
}
```

**Analysis Results**:
```json
{
  "id": "a1b2c3d4",
  "username": "jredh",
  "content_type": "comment",
  "content_id": "def456",
  "category": "pii",
  "severity": "high",
  "description": "email detected in content",
  "matched_text": "j***@example.com",
  "analyzed_at": "2025-01-15T14:35:00Z"
}
```

### 5. Analyzer (`internal/analyzer/`)

**Purpose**: Content analysis and risk detection

**Current Capabilities**:
- Slur detection
- PII detection (email, phone, SSN, credit card, address, IP)
- Severity scoring

#### Analysis Pipeline

```
1. Load posts/comments from storage
2. For each piece of content:
   a. Normalize text (lowercase)
   b. Run pattern matchers
   c. Check against slur list
   d. Create finding records
   e. Save encrypted findings
3. Generate summary statistics
```

#### Detection Patterns

**Slurs** (keyword matching):
- Configurable list of problematic terms
- Case-insensitive matching
- Currently hardcoded (future: external config)

**PII** (regex patterns):
```go
"email":        \b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b
"phone":        \b\d{3}[-.]?\d{3}[-.]?\d{4}\b
"ssn":          \b\d{3}-\d{2}-\d{4}\b
"credit_card":  \b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b
"address":      \b\d+\s+[A-Za-z0-9\s,]+\s+(Street|St|Avenue|Ave|...)\b
"ip_address":   \b(?:\d{1,3}\.){3}\d{1,3}\b
```

#### Severity Assignment

| Category | Type | Severity |
|----------|------|----------|
| Slur | Any | High |
| PII | SSN, Credit Card, Address, Phone, Email | High |
| PII | IP Address | Medium |

#### Future Enhancements

- Location extraction (city, state, country mentions)
- Writing style fingerprinting
- Sentiment analysis
- Political stance detection
- Cross-reference with other platforms
- Timeline anomaly detection

## Data Flow

### Scan Flow

```
User runs: digital-history scan reddit --username jredh

1. CLI validates credentials from config
2. Create encrypted storage connection
3. Initialize Reddit client
4. Reddit client:
   a. Authenticate with OAuth2
   b. Fetch posts (paginated)
   c. Encrypt and store each post
   d. Fetch comments (paginated)
   e. Encrypt and store each comment
5. Analyzer:
   a. Load all posts/comments
   b. Run detection patterns
   c. Create finding records
   d. Encrypt and store findings
6. Display summary to user
```

### Export Flow

```
User runs: digital-history export --username jredh --output report.json

1. CLI prompts for passphrase (if not in config)
2. Open encrypted storage
3. Fetch all posts for username
4. Decrypt post data
5. Fetch all comments for username
6. Decrypt comment data
7. Fetch all analysis results
8. Decrypt analysis data
9. Combine into JSON structure
10. Write to file or stdout
```

## Security Considerations

### Threat Model

**Protected Against**:
- Physical access to laptop (data encrypted at rest)
- Database theft (useless without passphrase)
- Casual inspection of storage

**Not Protected Against**:
- Malware with keylogger (can steal passphrase)
- Memory dumps while application running (data decrypted in memory)
- Compromise of config file (credentials in plaintext)
- Brute force of weak passphrase

### Security Best Practices

1. **Strong Passphrase**: Use 20+ character passphrase with high entropy
2. **File Permissions**: Config and database have restrictive permissions (0600)
3. **No Cloud Sync**: Keep data local only
4. **Rate Limiting**: Respect platform APIs to avoid account suspension
5. **Audit Logs**: Consider adding operation logging

### Potential Improvements

1. **Salt Generation**: Use random salt per database, store separately
2. **Key Rotation**: Support re-encrypting with new passphrase
3. **Secure Memory**: Use mlocked memory for passphrase
4. **Credential Storage**: Integrate with OS keychain
5. **2FA Support**: Add OAuth2 device flow for platforms that support it

## Performance

### Scalability

**Current Limits**:
- Tested with accounts up to 10,000 posts/comments
- Memory usage: ~200MB for large accounts
- Scan time: ~2 seconds per 100 items (Reddit rate limiting)

**Optimization Opportunities**:
- Batch encryption (encrypt multiple records at once)
- Streaming analysis (analyze while collecting)
- Parallel platform scans
- Incremental scans (only fetch new content)

### Database Performance

**Indexes**:
- Username (for filtering)
- Timestamps (for date range queries)
- Severity (for filtering findings)

**Query Optimization**:
- Use prepared statements
- Limit result sets
- Avoid SELECT *

## Testing Strategy

**Unit Tests** (future):
- Encryption/decryption
- Pattern matching
- Configuration loading

**Integration Tests** (future):
- Reddit API client (with mock server)
- Storage layer
- End-to-end scan

**Coverage Target**: 95% per project policy

## Dependencies

### Core Dependencies

```
github.com/spf13/cobra          - CLI framework
github.com/spf13/viper          - Configuration management
github.com/mattn/go-sqlite3     - SQLite driver
github.com/vartanbeno/go-reddit - Reddit API client
golang.org/x/crypto             - Encryption primitives
```

### Why These Choices

- **Cobra**: Industry standard for Go CLIs, excellent UX
- **Viper**: Flexible config, multiple formats, good defaults
- **SQLite**: Embedded, no setup, perfect for local storage
- **go-reddit**: Maintained, supports OAuth2, good documentation
- **crypto/x**: Official Go crypto, well-audited

## Deployment

### Build

```bash
make build          # Builds bin/digital-history
make install        # Installs to /usr/local/bin
```

### Distribution

**Current**: Build from source

**Future**:
- GitHub releases with binaries
- Homebrew formula
- Apt/Yum packages
- Docker image

### Cross-Platform

**Supported**:
- macOS (tested)
- Linux (should work)
- Windows (should work, may need adjustments)

**Platform-Specific**:
- Config path: Uses `os.UserHomeDir()`
- File paths: Uses `filepath.Join()`
- Permissions: Uses `0600` (may need Windows adjustment)

## Future Architecture

### Plugin System

```
digital-history/
├── core/                    # Core functionality
├── plugins/
│   ├── reddit/             # Reddit plugin
│   ├── twitter/            # Twitter plugin
│   └── github/             # GitHub plugin
└── analyzers/
    ├── pii/                # PII analyzer
    ├── sentiment/          # Sentiment analyzer
    └── opsec/              # OPSEC analyzer
```

### Web Interface

```
Browser → Web UI (React) → REST API → Core Engine → Storage
```

### Multi-User Support

```
digital-history-server/
├── users/
│   ├── user1/data.db
│   ├── user2/data.db
└── auth/
```

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development setup and guidelines.

## License

AGPL-3.0 - All modifications must remain open source.
