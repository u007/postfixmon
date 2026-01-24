# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Postfixmon monitors Postfix mail logs to detect and prevent spam from compromised email accounts. It scans `/var/log/mail.log`, tracks email activity per sender, enforces rate limits (per-minute and per-hour), and automatically suspends accounts that exceed thresholds via WHM (cPanel) or Virtualmin APIs.

## Build and Run Commands

```bash
# Build for Linux production
make build                    # Produces bin/postfixmon and bin/pffind

# Run locally with mail.log
API_TOKEN=xxx PF_LOG=mail.log go run main.go run

# Run tests
go test -v ./...

# Run a single test
go test -v -run TestEmailDomainName

# Rerun from a specific date
make rerun date="2022 oct 16"
```

## Architecture

**Main Flow (main.go):**
- `postfixLogScanner()` - Core log scanning loop that parses mail.log line by line
- `processSession()` - Handles each mail session, counting recipients and enforcing limits
- `mailCount()`/`mailCountStore()` - Read/write per-sender email counts from `data/` directory

**Server Type Backends:**
- `whm/` - cPanel/WHM API integration (requires `API_TOKEN` env var)
- `virtualmin/` - Virtualmin CLI integration (calls `/usr/sbin/virtualmin modify-user`)

**Utilities (`tools/`):**
- `tools.ParseDate()` - Parses dates in format `"2006 Jan 2 15:04:05"`
- `tools.FileExists()`, `tools.ReadContentAsList()` - File helpers

**State Files:**
- `.config` - Stores last processed line number and prefix for log resumption
- `data/<email>/<date>/` - Per-sender email count files (hourly: `HH`, minutely: `HHMM`)
- `skip.conf` - Whitelist of emails/domains to ignore (supports `*@domain.com` patterns)

## Key Environment Variables

- `SERVERTYPE` - `virtualmin` (default) or `cpanel`
- `API_TOKEN` - WHM API token (required for cpanel mode)
- `PF_LOG` - Path to mail.log (default: `/var/log/mail.log`)
- `MAX_PER_MIN` - Rate limit per minute (default: 8)
- `MAX_PER_HOUR` - Rate limit per hour (default: 100)
- `NOTIFY_EMAIL` - Email address for suspension notifications

## Log Parsing Regex

The regex in `postfixRegLine` extracts: date, session ID, direction (from/to), and email address from Postfix log entries:
```
(?i)([a-z]*  ?\d+ \d+:\d+:\d+) [a-zA-Z0-9_]* postfix/[a-z]*\[\d*\]: ([a-z0-9]*): (from|to)=<email>,.*$
```
