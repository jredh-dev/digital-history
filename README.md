# Digital History

Digital footprint analysis tool for personal security auditing.

## Overview

Digital History helps you analyze your digital footprint across multiple platforms to identify:
- Security vulnerabilities
- OPSEC issues
- Potentially problematic content
- PII exposure risks

## Features

- **Reddit Analysis**: Collect and analyze Reddit post/comment history
- **Content Analysis**: Detect problematic language, controversial statements
- **OPSEC Auditing**: Identify PII leaks, location information, identifiable patterns
- **Encrypted Storage**: SQLite database with encryption at rest
- **Risk Scoring**: Prioritized recommendations for account management

## Installation

```bash
go install github.com/jredh-dev/digital-history@latest
```

## Quick Start

```bash
# Configure Reddit API credentials
digital-history config set reddit.client_id YOUR_CLIENT_ID
digital-history config set reddit.client_secret YOUR_CLIENT_SECRET
digital-history config set reddit.username YOUR_USERNAME

# Scan a Reddit user
digital-history scan reddit --username jredh

# View collected data
digital-history list --username jredh

# Export results
digital-history export --username jredh --format json
```

## Privacy

- All data stored locally only
- SQLite database encrypted at rest
- No cloud uploads or external data sharing
- Data deletion commands available

## Development Status

🚧 **Early Development** - Core functionality being implemented.

## License

AGPL-3.0 License - See LICENSE file for details.

This project is licensed under the GNU Affero General Public License v3.0, which requires that any modified versions or network services using this code must also be open source.
