# Claude Code Bar

macOS menu bar app that shows your Claude Code API usage at a glance.

![macOS](https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=white)
![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)

## What it does

Sits in your menu bar and displays remaining Claude Code capacity:

- **CC: 73%** — you're good
- **CC: ⚠ 35%** — slow down
- **CC: 🔴 12%** — almost out

Click it for details:
- 5-hour window remaining % and reset time
- 7-day window remaining % and reset time

Refreshes automatically every 60 seconds.

## Installation

### Download binary

Grab the latest release from [Releases](https://github.com/enlabs-org/claude-code-bar/releases):

- **Apple Silicon (M1/M2/M3/M4):** `claude-code-bar-darwin-arm64`
- **Intel Mac:** `claude-code-bar-darwin-amd64`

```bash
# Example for Apple Silicon
chmod +x claude-code-bar-darwin-arm64
./claude-code-bar-darwin-arm64
```

### Build from source

```bash
git clone https://github.com/enlabs-org/claude-code-bar.git
cd claude-code-bar
go build -o claude-code-bar .
./claude-code-bar
```

## Prerequisites

- **Claude Code** must be installed and authenticated (the app reads the OAuth token from macOS Keychain)
- macOS only (uses native menu bar API via [menuet](https://github.com/caseymrm/menuet))

## How it works

1. Reads Claude Code OAuth credentials from macOS Keychain (`Claude Code-credentials`)
2. Calls the Anthropic usage API (`/api/oauth/usage`)
3. Displays remaining capacity in the menu bar
4. Polls every 60 seconds

## License

MIT
