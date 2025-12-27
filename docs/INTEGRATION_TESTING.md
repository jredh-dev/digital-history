# Integration Testing Guide

This document explains how to run integration tests for the digital-history analyzers.

## Overview

Integration tests verify that the analyzers work correctly with real Reddit data. These tests:
- Make actual API calls to Reddit
- Fetch real user data
- Run all analyzers (political, NSFW, PII, slurs)
- Generate analysis reports

## Prerequisites

### 1. Reddit API Credentials

You need a Reddit application with OAuth2 credentials:

1. Go to https://www.reddit.com/prefs/apps
2. Click "Create App" or "Create Another App"
3. Fill in the form:
   - **Name**: `digital-history-test`
   - **App type**: Select **"script"**
   - **Description**: Integration testing
   - **Redirect URI**: `http://localhost:8080`
4. Click "Create app"
5. Note your **client_id** (under the app name) and **client_secret**

### 2. Environment Variables

Export your Reddit credentials:

```bash
export REDDIT_CLIENT_ID="your_client_id_here"
export REDDIT_CLIENT_SECRET="your_client_secret_here"
export REDDIT_USERNAME="your_reddit_username"
export REDDIT_PASSWORD="your_reddit_password"
```

**Security Note**: These are temporary for testing. Never commit credentials to git.

## Running Integration Tests

### Test Against Public Account (spez)

This test analyzes Reddit CEO Steve Huffman's public account to verify analyzer functionality:

```bash
cd internal/analyzer
go test -v -run TestIntegration_AnalyzeRedditUser
```

**What it does**:
- Scans the `spez` account (Reddit's CEO)
- Collects recent posts and comments
- Runs political, NSFW, and PII analyzers
- Reports findings by category

**Expected results**:
- Should find political content (spez comments on Reddit policy, moderation, etc.)
- May find some NSFW subreddit participation (spez moderates various subs)
- Minimal PII (spez is privacy-conscious)

### Test Against Your Own Account

This test analyzes your authenticated Reddit account:

```bash
go test -v -run TestIntegration_AnalyzeCurrentUser
```

**What it does**:
- Scans your own Reddit account
- Runs all analyzers
- Generates cleanup recommendations by severity:
  - **High**: Strongly recommend deletion (PII, slurs)
  - **Medium**: Consider deletion (political, NSFW)
  - **Low**: Optional cleanup

**Use case**: This is the "self-audit" mode for finding content you may want to delete.

## Dual-Account Strategy

### Reading Account vs. Target Account

For shadowban detection and rate-limit management, you can use two accounts:

**Reading Account** (credentials in environment variables):
- Used to fetch public data via API
- Sees what everyone else sees
- Preserves your main account's rate limits

**Target Account** (the one being analyzed):
- Can be any public Reddit user
- Specified in test or CLI command
- For `spez` test: reads spez's public content
- For self-test: reads your own content

### Shadowban Detection (Future Feature)

Compare results from two perspectives:

1. **Logged out view** (reading account sees what's public)
2. **Logged in view** (your account sees what you posted)

Content visible to you but not to the reading account = shadowbanned.

## Integration Test Output

Example output:

```
=== RUN   TestIntegration_AnalyzeRedditUser
    integration_test.go:71: Scanning Reddit user: spez
    integration_test.go:82: Collected 100 posts
    integration_test.go:87: Collected 100 comments
    integration_test.go:90: Running political content analyzer...
    integration_test.go:95: Running NSFW content analyzer...
    integration_test.go:100: Running PII/slur analyzer...
    integration_test.go:107: Total analysis results: 47
    integration_test.go:115: Analysis results by category:
    integration_test.go:116:   political: 38 findings
    integration_test.go:116:   nsfw: 5 findings
    integration_test.go:116:   pii: 4 findings
    integration_test.go:121: Political findings: 38
    integration_test.go:123: Sample political findings:
    integration_test.go:128:   - [comment] Political keyword detected: government: government
    integration_test.go:128:   - [post] Political keyword detected: policy: policy
--- PASS: TestIntegration_AnalyzeRedditUser (47.23s)
PASS
ok      github.com/jredh-dev/digital-history/internal/analyzer 47.541s
```

## Interpreting Results

### Political Findings

Indicates:
- Posts/comments in political subreddits (r/politics, r/conservative, etc.)
- Political keywords (government, congress, biden, trump, etc.)
- May want to delete if trying to reduce political footprint

### NSFW Findings

Indicates:
- Posts/comments in NSFW subreddits
- Thirst trap keywords (onlyfans, dm me, etc.)
- Consider deletion for professional profile cleanup

### PII Findings

Indicates:
- Email addresses, phone numbers in content
- Physical addresses, IP addresses
- **High priority for deletion** (security risk)

### Slur Findings

Indicates:
- Offensive language detected
- **High priority for deletion** (reputation risk)

## Troubleshooting

### "Reddit credentials not provided"

Solution: Set environment variables (see Prerequisites)

### "Rate limit exceeded"

Solution: Reddit API limits requests. Wait 10 minutes and try again.

### "Failed to scan user"

Possible causes:
- Invalid credentials
- User doesn't exist
- User account is private/deleted
- Network connectivity issues

### "No findings detected"

Possible causes:
- Target account has minimal activity
- Content doesn't match detection patterns
- Recent account with few posts

## Next Steps

After running integration tests:

1. **Review findings**: Understand what content was flagged
2. **Adjust patterns**: Edit analyzer keyword lists if needed
3. **Test cleanup**: Use the `review` command to see deletion candidates
4. **Dry-run deletion**: Test the `cleanup` command with `--dry-run`

See main README.md for full CLI usage.

## CI/CD Integration

Integration tests are skipped by default (`testing.Short()`). To run in CI:

```bash
# Skip integration tests (default)
go test -short ./...

# Run all tests including integration (requires credentials)
go test ./...
```

## Privacy & Ethics

**Important**:
- Only analyze accounts you own or have permission to analyze
- Public figures (like spez) are fair game for testing
- Respect Reddit's ToS and rate limits
- Don't use findings maliciously or for harassment
- This tool is for **personal security auditing** only
