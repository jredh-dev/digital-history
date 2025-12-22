# Digital History

Digital footprint analysis tool for personal security auditing.

## Overview

Digital History helps you analyze your digital footprint across multiple platforms to identify:
- Security vulnerabilities
- OPSEC issues  
- Potentially problematic content
- PII exposure risks (emails, phone numbers, addresses)

## Features

- **Reddit Analysis**: Collect and analyze Reddit post/comment history
- **Content Analysis**: Detect problematic language, slurs, controversial statements
- **OPSEC Auditing**: Identify PII leaks, location information, identifiable patterns
- **Encrypted Storage**: AES-256-GCM encrypted SQLite database
- **Risk Scoring**: Prioritized recommendations by severity (high/medium/low)

## Installation

### From Source

```bash
git clone https://github.com/jredh-dev/digital-history.git
cd digital-history
make build
make install  # Optional: installs to /usr/local/bin
```

### Or install directly

```bash
go install github.com/jredh-dev/digital-history@latest
```

## Prerequisites

### Reddit API Setup

1. Go to https://www.reddit.com/prefs/apps
2. Click "Create App" or "Create Another App"
3. Fill in the form:
   - **Name**: digital-history
   - **App type**: Select "script"
   - **Description**: Personal security auditing tool
   - **About URL**: (leave blank)
   - **Redirect URI**: http://localhost:8080
4. Click "Create app"
5. Note your **client_id** (under the app name) and **client_secret**

## Quick Start

### 1. Configure Reddit Credentials

```bash
digital-history config set reddit.client_id YOUR_CLIENT_ID
digital-history config set reddit.client_secret YOUR_CLIENT_SECRET
digital-history config set reddit.username YOUR_REDDIT_USERNAME
digital-history config set reddit.password YOUR_REDDIT_PASSWORD
```

### 2. Set Database Passphrase (Optional but Recommended)

```bash
digital-history config set database.passphrase YOUR_SECURE_PASSPHRASE
```

If not set, you'll be prompted to enter it each time.

### 3. Scan a Reddit User

```bash
# Scan and analyze
digital-history scan reddit --username jredh

# Scan without analysis
digital-history scan reddit --username jredh --analyze=false
```

The tool will:
- Collect all posts and comments from the user
- Store them encrypted in SQLite database
- Analyze content for problematic language and PII
- Display a summary of findings

### 4. View Results

```bash
# View summary
digital-history list --username jredh

# Export to JSON
digital-history export --username jredh --format json

# Export to file
digital-history export --username jredh --output report.json
```

## Commands

### scan

Collect data from a platform and optionally analyze it.

```bash
digital-history scan <platform> --username <username> [--analyze=true]
```

Supported platforms:
- `reddit` - Reddit posts and comments

### list

Display a summary of collected data and analysis results.

```bash
digital-history list --username <username>
```

Shows:
- Total posts and comments collected
- Analysis results grouped by severity
- Top recent findings

### export

Export collected data and analysis results to JSON.

```bash
digital-history export --username <username> [--format json] [--output file.json]
```

If `--output` is omitted, exports to stdout.

### config

Manage configuration settings.

```bash
# Set a value
digital-history config set <key> <value>

# Get a value (sensitive values are redacted)
digital-history config get <key>

# Show all config
digital-history config show
```

Configuration keys:
- `reddit.client_id` - Reddit OAuth app client ID
- `reddit.client_secret` - Reddit OAuth app client secret  
- `reddit.username` - Reddit username
- `reddit.password` - Reddit password
- `database.path` - Database file location (default: ~/.digital-history/data.db)
- `database.passphrase` - Encryption passphrase
- `analyzer.enable_slur_detection` - Enable/disable slur detection (default: true)

## Privacy & Security

### Encryption

All sensitive data is encrypted at rest using AES-256-GCM:
- Reddit posts and comments (title, body, URLs)
- Analysis results (matched text, descriptions)
- Passphrase is derived using PBKDF2 with 100,000 iterations

### Local Storage

- All data stored locally in `~/.digital-history/data.db`
- No cloud uploads or external API calls (except to Reddit for collection)
- Configuration stored in `~/.digital-history.yaml`

### Credential Security

- Reddit credentials stored in plaintext config file (file permissions: 0600)
- Consider using environment variables or a secrets manager for production use
- Database passphrase encrypts all collected content

### Rate Limiting

- Respects Reddit API rate limits (2 second delay between requests)
- Uses authenticated API access for higher limits

## What Gets Detected

### Content Analysis

- **Slurs**: Racial, homophobic, transphobic, ableist language
- **PII**: Emails, phone numbers, SSNs, credit cards, addresses, IP addresses
- **Risk Levels**:
  - 🔴 **High**: Slurs, SSNs, credit cards, physical addresses
  - 🟡 **Medium**: IP addresses, less sensitive PII
  - 🟢 **Low**: General privacy concerns

### Future Analysis (Planned)

- Location information extraction
- Writing style fingerprinting
- Cross-platform username correlation
- Sentiment/political stance tracking
- Account age and activity patterns

## Development Status

✅ **MVP Complete** - Core functionality implemented:
- Reddit API integration
- Encrypted storage
- Basic content analysis (slurs, PII)
- CLI commands (scan, list, export, config)

🚧 **In Progress**:
- Additional platform support (Twitter, GitHub, Facebook)
- Advanced pattern detection
- HTML report generation
- Account deletion recommendations

## Development

```bash
# Build
make build

# Run tests (when implemented)
make test

# Install locally
make install

# Clean build artifacts
make clean

# Format and lint
make lint
```

## Contributing

Contributions welcome! This is an AGPL-3.0 project - all derivatives must remain open source.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

AGPL-3.0 License - See LICENSE file for details.

This project is licensed under the GNU Affero General Public License v3.0, which requires that any modified versions or network services using this code must also be open source.

## Disclaimer

This tool is intended for personal security auditing of your own accounts. Respect others' privacy and platform terms of service.

## Support

- Issues: https://github.com/jredh-dev/digital-history/issues
- Documentation: https://github.com/jredh-dev/digital-history
