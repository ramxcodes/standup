# standup

A simple terminal CLI that helps you prepare quickly for standup. Generates a report from recent git commits (optionally with an AI summary via Gemini).

**macOS only** (Apple Silicon & Intel).

---

## Installation

**Prerequisites:** macOS, zsh (default on macOS).

### One-liner (interactive — choose install location)

```bash
curl -fsSL https://raw.githubusercontent.com/ramxcodes/standup/master/install.sh -o install.sh && sh install.sh
```

You’ll be prompted to pick where to install (e.g. `/opt/homebrew/bin`, `~/.local/bin`, or a custom path).

### Non-interactive (use default location)

```bash
curl -fsSL https://raw.githubusercontent.com/ramxcodes/standup/master/install.sh | sh
```

- **Existing installs:** If standup is already installed in the chosen directory, it is removed and the latest release is installed.
- **Weekly auto-update (optional):** When you run the installer interactively, you can opt in to weekly auto-update. A launchd job runs every Sunday at 3:00 AM and installs the latest release only if a new version exists. To disable: `launchctl unload ~/Library/LaunchAgents/com.standup.cli.update.plist`

If the installer says **404** when downloading the binary, a release with binaries hasn’t been published yet. See [Creating a release](#creating-a-release) below.

After install, run `standup` from any folder (open a new terminal or `source ~/.zshrc` if you used a custom path).

---

## Usage

Run from a git repo or a directory that contains git repos:

```bash
standup
```

### Options

| Flag | Short | Description |
|------|--------|-------------|
| `--days` | `-d` | Number of days to look back (default: 1) |
| `--author` | `-a` | Filter commits by author (default: current git user) |
| `--version` | `-v` | Show version |

### AI (Gemini)

| Flag | Description |
|------|-------------|
| `--set-api-key` | Set your Gemini API key |
| `--set-model-name` | Set Gemini model name |
| `--show-api-key` | Show stored API key (masked) |
| `--remove-api-key` | Remove stored API key |
| `--enable-ai` | Enable AI summary |
| `--disable-ai` | Disable AI summary |

### Examples

```bash
# Last day (default)
standup

# Last 3 days
standup -d 3

# Last 2 days, only your commits
standup -d 2 -a "Your Name"

# One-time: set API key and enable AI
standup --set-api-key YOUR_GEMINI_API_KEY
standup --enable-ai

# Then run as usual — AI summary will be included when enabled
standup -d 2
```

---

## Creating a release (maintainers)

The install script downloads binaries from GitHub Releases. To publish a new version:

1. Build the binaries:
   ```bash
   ./scripts/build-release.sh
   ```
2. Create a new [Release](https://github.com/ramxcodes/standup/releases/new) on GitHub (tag e.g. `v1.0.0`).
3. Upload the two files from `dist/`:
   - `standup-darwin-arm64`
   - `standup-darwin-amd64`

After that, the install one-liner will work for everyone.

---

## License & contact

Made with ♡ by Ram · [ramx.in](https://ramx.in)
