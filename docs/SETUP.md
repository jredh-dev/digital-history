# Setup Guide

Complete guide to setting up and using digital-history for the first time.

## Table of Contents

1. [Installation](#installation)
2. [Reddit API Setup](#reddit-api-setup)
3. [Configuration](#configuration)
4. [First Scan](#first-scan)
5. [Understanding Results](#understanding-results)
6. [Troubleshooting](#troubleshooting)

## Installation

### Option 1: Install from Source (Recommended for development)

```bash
# Clone the repository
git clone https://github.com/jredh-dev/digital-history.git
cd digital-history

# Build the binary
make build

# Optional: Install system-wide
sudo make install
```

### Option 2: Install with Go

```bash
go install github.com/jredh-dev/digital-history@latest
```

### Verify Installation

```bash
digital-history --help
```

You should see the help menu with available commands.

## Reddit API Setup

To scan Reddit accounts, you need to create a Reddit application:

### Step 1: Create Reddit App

1. Log in to Reddit
2. Navigate to: https://www.reddit.com/prefs/apps
3. Scroll to "Developed Applications"
4. Click **"Create App"** or **"Create Another App"**

### Step 2: Fill Application Form

- **Name**: `digital-history` (or any name you prefer)
- **App type**: Select **"script"** (important!)
- **Description**: `Personal security auditing tool`
- **About URL**: Leave blank or use `https://github.com/jredh-dev/digital-history`
- **Redirect URI**: `http://localhost:8080` (required even though we don't use it)

### Step 3: Save Credentials

After clicking "Create app", note down:

- **Client ID**: The string under your app name (looks like: `abc123def456`)
- **Client Secret**: The string labeled "secret" (looks like: `XyZ789-abcDEF...`)

**Important**: Keep these credentials secret! They provide access to your Reddit account.

## Configuration

### Set Reddit Credentials

```bash
# Configure Reddit API access
digital-history config set reddit.client_id YOUR_CLIENT_ID
digital-history config set reddit.client_secret YOUR_CLIENT_SECRET
digital-history config set reddit.username YOUR_REDDIT_USERNAME
digital-history config set reddit.password YOUR_REDDIT_PASSWORD
```

Replace:
- `YOUR_CLIENT_ID` - Client ID from Step 3
- `YOUR_CLIENT_SECRET` - Client secret from Step 3
- `YOUR_REDDIT_USERNAME` - Your Reddit username
- `YOUR_REDDIT_PASSWORD` - Your Reddit password

### Set Database Encryption Passphrase

Choose a strong passphrase to encrypt your local database:

```bash
digital-history config set database.passphrase "YOUR_SECURE_PASSPHRASE"
```

**Tips for a good passphrase**:
- Use a mix of words, numbers, and symbols
- Make it at least 20 characters
- Don't reuse passwords from other services
- Consider using a passphrase generator

### Verify Configuration

```bash
digital-history config show
```

You should see your settings (sensitive values will be redacted).

### Configuration File Location

Config is stored at: `~/.digital-history.yaml`

**File permissions**: The config file is created with restrictive permissions (0600) but still contains credentials in plaintext. Consider:
- Using a secure location
- Setting up file-based encryption
- Using environment variables for CI/CD

## First Scan

### Scan Your Account

```bash
# Scan with analysis (recommended for first run)
digital-history scan reddit --username YOUR_REDDIT_USERNAME

# Or scan without analysis (faster, analyze later)
digital-history scan reddit --username YOUR_REDDIT_USERNAME --analyze=false
```

### What Happens During a Scan

1. **Authentication**: Connects to Reddit API using your credentials
2. **Data Collection**: 
   - Fetches all your posts (title, body, URL, timestamps)
   - Fetches all your comments (body, timestamps)
   - Rate limits to 2 seconds between requests (respectful to Reddit)
3. **Storage**: Encrypts and saves data to local SQLite database
4. **Analysis** (if enabled):
   - Scans for problematic language (slurs, offensive terms)
   - Detects PII (emails, phone numbers, addresses, SSNs, credit cards)
   - Identifies IP addresses
   - Assigns severity levels (high/medium/low)
5. **Summary**: Displays findings grouped by category

### Progress Indicators

```
Scanning Reddit user: jredh
Collecting posts...
  Collected 100 posts...
  Collected 200 posts...
Total posts collected: 250

Collecting comments...
  Collected 100 comments...
  Collected 500 comments...
  Collected 1000 comments...
Total comments collected: 1247

Running content analysis...
Analyzing 250 posts...
Analyzing 1247 comments...
Analysis complete!

=== Analysis Summary ===
  pii_high: 3
  slur_high: 1
  pii_medium: 7
```

## Understanding Results

### View Summary

```bash
digital-history list --username YOUR_REDDIT_USERNAME
```

Output:
```
=== Digital History: jredh ===

Posts:              250
Comments:           1247
Analysis Results:   11

=== Top Issues ===
  🔴 High severity:   4
  🟡 Medium severity: 7

Recent Findings:
  [high] pii: email detected in content
  [high] slur: Potentially offensive language detected
  [high] pii: phone detected in content
  [high] pii: address detected in content
  [medium] pii: ip_address detected in content
  ...
```

### Export Full Report

```bash
# Export to stdout
digital-history export --username YOUR_REDDIT_USERNAME

# Export to file
digital-history export --username YOUR_REDDIT_USERNAME --output report.json
```

### Understanding Severity Levels

#### 🔴 High Severity
- **Slurs**: Offensive, discriminatory language
- **PII**: SSNs, credit cards, physical addresses, phone numbers, emails
- **Action**: Review immediately, consider editing or deleting posts

#### 🟡 Medium Severity  
- **IP Addresses**: Could reveal approximate location
- **Action**: Review and assess risk based on context

#### 🟢 Low Severity
- **General Privacy**: Minor privacy concerns
- **Action**: Awareness, no immediate action needed

### What to Do with Findings

1. **Review each high-severity finding**:
   - Click through to the original post/comment
   - Assess if the content is truly problematic
   - Consider editing or deleting

2. **False positives**: The analyzer may flag legitimate content (e.g., discussing email formats). Use your judgment.

3. **Account strategy**:
   - If many concerning findings: Consider retiring the username
   - Create a new account with better OPSEC practices
   - Be more mindful of what you share publicly

## Troubleshooting

### "Failed to create reddit client"

**Cause**: Incorrect Reddit credentials

**Solution**:
1. Verify your client_id and client_secret are correct
2. Ensure app type is "script" (not "web app" or "installed app")
3. Check username/password are correct
4. Try regenerating the client secret on Reddit

### "Passphrase required for encrypted storage"

**Cause**: Database passphrase not configured

**Solution**:
```bash
digital-history config set database.passphrase "YOUR_PASSPHRASE"
```

Or enter it when prompted during each command.

### "Failed to decrypt"

**Cause**: Wrong passphrase or corrupted database

**Solution**:
1. Verify you're using the correct passphrase
2. If you forgot the passphrase, you'll need to delete the database and re-scan
3. Database location: `~/.digital-history/data.db`

### Rate Limiting

**Symptom**: Slow scanning or API errors

**Info**: The tool waits 2 seconds between requests to respect Reddit's rate limits. For large accounts (1000+ comments), this is normal and may take 30+ minutes.

**Solution**: Be patient, or run the scan overnight.

### "No posts/comments found"

**Cause**: Username doesn't exist or has no public content

**Solution**:
1. Verify the username is spelled correctly
2. Check if the account has public posts/comments
3. Ensure the account isn't shadowbanned

### Memory Issues

**Symptom**: High memory usage during large scans

**Solution**: This is normal for accounts with 10,000+ posts/comments. The tool loads batches of 100 at a time and encrypts them before moving to the next batch.

## Data Management

### Database Location

Default: `~/.digital-history/data.db`

Change with:
```bash
digital-history config set database.path /custom/path/data.db
```

### Deleting Data

To completely remove all collected data:

```bash
rm ~/.digital-history/data.db
rm ~/.digital-history.yaml  # Also removes config
```

### Backup

To backup your encrypted database:

```bash
cp ~/.digital-history/data.db ~/backups/digital-history-backup-$(date +%Y%m%d).db
```

**Note**: Backup is still encrypted. You'll need the passphrase to access it.

## Privacy Notes

### What Gets Stored

- Reddit posts (ID, username, subreddit, title, body, URL, score, timestamps)
- Reddit comments (ID, username, post_id, subreddit, body, score, timestamps)
- Analysis results (findings, matched text, severity)

### What Doesn't Get Stored

- Your Reddit password (only used for authentication, never stored)
- Reddit API tokens (temporary, regenerated each session)
- Images or media (only URLs are stored)
- Deleted posts (can't fetch what's already deleted)

### Encryption Details

- **Algorithm**: AES-256-GCM (authenticated encryption)
- **Key Derivation**: PBKDF2 with SHA-256, 100,000 iterations
- **Encrypted Fields**: All post/comment content, analysis results
- **Unencrypted**: Metadata for querying (username, timestamps, IDs)

### Network Activity

The tool makes HTTP requests to:
- `oauth.reddit.com` - Reddit API for fetching posts/comments
- No other external services

## Next Steps

After your first scan:

1. **Review findings**: Look at the analysis results
2. **Clean up**: Edit or delete problematic content
3. **Regular scans**: Run monthly to catch new issues
4. **Expand coverage**: Plan for other platforms (Twitter, GitHub, etc.)

## Getting Help

- **Issues**: https://github.com/jredh-dev/digital-history/issues
- **Documentation**: https://github.com/jredh-dev/digital-history
- **Security concerns**: Open a private security advisory on GitHub
