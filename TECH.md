# Technical Documentation

## Project Overview

*   **Purpose:** Monitor Postfix mail logs to detect and prevent spam from compromised accounts.
*   **Core Functionality:** Scans mail logs, tracks email activity, enforces limits, suspends accounts (via WHM/Virtualmin), sends notifications.

## Architecture and Components

*   **Language:** Go
*   **Main Program:** `main.go`
    *   Handles command-line arguments (`start`, `run`, `rerun`, `skip`, `reset`, `suspend`, `unsuspend`, `info`, `test-notify`, `help`).
    *   Manages environment variables (e.g., `API_TOKEN`, `NOTIFY_EMAIL`, `PF_LOG`, `MAX_PER_MIN`, `MAX_PER_HOUR`, `WHM_API_HOST`, `SERVERTYPE`).
*   **Log Parsing (`main.go`):**
    *   Scans the Postfix log file (typically `/var/log/mail.log`).
    *   Uses the regex: `(?i)([a-z]* \d+ \d+:\d+:\d+) [a-zA-Z0-9_]* postfix/[a-z]*\[\d*\]: ([a-z0-9]*): (from|to)=<([a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4})>,.*$` to extract date, session ID, direction (from/to), and email address.
*   **State Management (`.config` file):**
    *   Stores the size of the log file, the last processed line number, and a prefix of the last processed line to allow resumption of scanning.
    *   Location: Root directory of the application.
*   **Email Statistics Storage (`data/` directory):**
    *   Stores counts of emails sent per minute and per hour.
    *   Structure: `data/<cleaned_email_address>/<YYYY_MM_DD>/<HH>` (hourly count) and `data/<cleaned_email_address>/<YYYY_MM_DD>/<HHMM>` (minutely count).
    *   `<cleaned_email_address>` is the email with `@` and `-` replaced by `_`.
*   **WHM Integration (`whm/` directory):**
    *   Communicates with WHM API v1 for email suspension/unsubscription.
    *   Key functions: `whm.SuspendEmail`, `whm.UnSuspendEmail`.
    *   API calls used: `Email/suspend_outgoing`, `Email/unsuspend_outgoing`.
    *   Authentication: `Authorization: whm <ApiUser>:<ApiToken>`. Note: `ApiUser` (likely 'root') is not explicitly set via an environment variable in `main.go` but is part of the WHM library's internal constants or assumptions. `API_TOKEN` must be provided.
*   **Virtualmin Integration (`virtualmin/` directory):**
    *   Uses Virtualmin CLI (`/usr/sbin/virtualmin`) for email suspension/unsubscription.
    *   Key functions: `virtualmin.SuspendEmail`, `virtualmin.EnableEmail`.
    *   Commands: `virtualmin modify-user --domain <domain> --user <username> --disable` and `virtualmin modify-user --domain <domain> --user <username> --enable`.
    *   Assumption: The local part of the email address is treated as the Virtualmin username.
*   **Notification (`main.go` - `notifySuspend` function):**
    *   Uses the system `mail` command to send email notifications.
*   **Whitelist (`skip.conf`):**
    *   File containing email addresses or domain patterns (`*@domain.com`) to ignore.

## Setup and Configuration

*   Briefly mention required environment variables (refer to README for details).
*   Location of `skip.conf` and `.config`.

## Known Issues and Limitations

*   **Log Rotation:** Sensitivity to log rotation. While there's a size check, if the log resets and the last read line's prefix isn't found, it rescans from the beginning. This could be an issue if the prefix is genuinely gone after rotation, potentially leading to repeated full scans.
*   **Error Handling:** Frequent use of `panic` can halt the monitoring service. More resilient error handling would be beneficial.
*   **Virtualmin Error Reporting:** `virtualmin/email.go` currently ignores errors from `virtualmin` CLI commands. (This is planned to be fixed).
*   **WHM `ApiUser`:** The WHM API user (typically 'root') is assumed by the `whm` package and not configured via an env var in `main.go`.
*   **Hardcoded Paths:** `virtualmin` command path (`/usr/sbin/virtualmin`), `.config`, and `data/` are hardcoded.

## Dependencies

*   External Go packages: None explicitly listed in `go.mod` (relies on standard library and `go get` for any transitive dependencies).
*   System commands: `mail` (for notifications), `/usr/sbin/virtualmin` (if `SERVERTYPE=virtualmin`).

## Development Setup

*   Refer to README.md.
