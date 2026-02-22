#!/usr/bin/env sh
# Install standup CLI on macOS (Apple Silicon or Intel).
# Usage: curl -fsSL https://raw.githubusercontent.com/ramxcodes/standup/master/install.sh -o install.sh && sh install.sh
# Or (non-interactive): curl -fsSL https://raw.githubusercontent.com/ramxcodes/standup/master/install.sh | sh

set -e

GITHUB_REPO="${GITHUB_REPO:-ramxcodes/standup}"
BINARY_NAME="standup"
RELEASE_BASE="https://github.com/${GITHUB_REPO}/releases/latest/download"

# Detect OS and arch (macOS only for this script)
OS=$(uname -s)
ARCH=$(uname -m)

if [ "$OS" != "Darwin" ]; then
  echo "This install script is for macOS only. OS detected: $OS"
  exit 1
fi

case "$ARCH" in
  arm64|aarch64)  SUFFIX="darwin-arm64" ;;
  x86_64)         SUFFIX="darwin-amd64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

DOWNLOAD_URL="${RELEASE_BASE}/${BINARY_NAME}-${SUFFIX}"
echo "Installing standup for ${SUFFIX}..."
echo "  from: $DOWNLOAD_URL"
echo ""

# Default install dir: first writable of /usr/local/bin, Homebrew bin, or ~/.local/bin
DEFAULT_INSTALL_DIR=""
if [ -w /usr/local/bin ] 2>/dev/null; then
  DEFAULT_INSTALL_DIR="/usr/local/bin"
elif [ -w /opt/homebrew/bin ] 2>/dev/null; then
  DEFAULT_INSTALL_DIR="/opt/homebrew/bin"
else
  DEFAULT_INSTALL_DIR="${HOME}/.local/bin"
fi

# Use default or env (STANDUP_INSTALL_DIR for non-interactive/weekly updater)
INSTALL_DIR="${STANDUP_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
# Interactive: let user choose install location (or use default)
if [ -t 0 ] && [ -z "${STANDUP_INSTALL_DIR:-}" ]; then
  printf "Where would you like to install standup?\n"
  printf "  [1] %s (default)\n" "$DEFAULT_INSTALL_DIR"
  printf "  [2] /opt/homebrew/bin\n"
  printf "  [3] %s/.local/bin\n" "$HOME"
  printf "  [4] Enter a custom path\n"
  printf "\nChoice [1]: "
  read -r choice
  choice="${choice:-1}"
  case "$choice" in
    1) INSTALL_DIR="$DEFAULT_INSTALL_DIR" ;;
    2) INSTALL_DIR="/opt/homebrew/bin" ;;
    3) INSTALL_DIR="${HOME}/.local/bin" ;;
    4)
      printf "Enter full path (e.g. %s/bin or /usr/local/bin): " "$HOME"
      read -r custom_path
      custom_path=$(echo "$custom_path" | sed "s|^~|$HOME|")
      if [ -z "$custom_path" ]; then
        echo "Using default: $DEFAULT_INSTALL_DIR"
        INSTALL_DIR="$DEFAULT_INSTALL_DIR"
      else
        INSTALL_DIR="$custom_path"
      fi
      ;;
    *)
      echo "Using default: $DEFAULT_INSTALL_DIR"
      INSTALL_DIR="$DEFAULT_INSTALL_DIR"
      ;;
  esac
  echo ""
  echo "Installing to: $INSTALL_DIR"
  echo ""
fi

mkdir -p "$INSTALL_DIR"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Error: Cannot write to $INSTALL_DIR (no permission). Try a path under your home, e.g. ${HOME}/.local/bin"
  exit 1
fi

# Remove existing standup if present so we install the latest cleanly
EXISTING="${INSTALL_DIR}/${BINARY_NAME}"
if [ -f "$EXISTING" ] || [ -L "$EXISTING" ]; then
  echo "Removing existing standup installation at $EXISTING..."
  rm -f "$EXISTING"
fi

# If installing to ~/.local/bin (or another dir that might not be on PATH), ensure .zshrc has it
SHELL_RC="${HOME}/.zshrc"
case "$INSTALL_DIR" in
  /usr/local/bin|/opt/homebrew/bin) ;;
  *)
    PATH_ADD="export PATH=\"${INSTALL_DIR}:\$PATH\""
    if [ -f "$SHELL_RC" ]; then
      if ! grep -q "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
        echo "" >> "$SHELL_RC"
        echo "# standup CLI" >> "$SHELL_RC"
        echo "$PATH_ADD" >> "$SHELL_RC"
        echo "Added ${INSTALL_DIR} to PATH in $SHELL_RC"
      fi
    else
      echo "$PATH_ADD" >> "$SHELL_RC"
      echo "Created $SHELL_RC and added ${INSTALL_DIR} to PATH"
    fi
    ;;
esac

TMP_FILE=$(mktemp)
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "$TMP_FILE" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$TMP_FILE" "$DOWNLOAD_URL"
else
  echo "Need curl or wget to download the binary."
  rm -f "$TMP_FILE"
  exit 1
fi

chmod +x "$TMP_FILE"
mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
echo "Installed to ${INSTALL_DIR}/${BINARY_NAME}"

# Save install dir for weekly updater (and future re-installs)
STANDUP_CONFIG_DIR="${HOME}/.config/standup"
mkdir -p "$STANDUP_CONFIG_DIR"
echo "$INSTALL_DIR" > "${STANDUP_CONFIG_DIR}/install-dir"
echo "$GITHUB_REPO" > "${STANDUP_CONFIG_DIR}/github-repo"

# Remind to reload shell if we added something to .zshrc
case "$INSTALL_DIR" in
  /usr/local/bin|/opt/homebrew/bin) echo "Run: standup" ;;
  *) echo "Run \`source ~/.zshrc\` or open a new terminal, then run: standup" ;;
esac

# Optional: enable weekly auto-update (only when interactive)
if [ -t 0 ]; then
  printf "\nEnable weekly auto-update? (checks for new release every Sunday) [y/N]: "
  read -r enable_update
  enable_update="${enable_update:-n}"
  case "$enable_update" in
    y|Y|yes|YES)
      UPDATE_SCRIPT="${HOME}/.standup-update.sh"
      cat << 'STANDUP_UPDATE_EOF' | sed "s|__INSTALL_DIR__|$INSTALL_DIR|g" | sed "s|__GITHUB_REPO__|$GITHUB_REPO|g" > "$UPDATE_SCRIPT"
#!/usr/bin/env sh
# Weekly updater for standup CLI. Only updates if a new release exists.
set -e
INSTALL_DIR="__INSTALL_DIR__"
GITHUB_REPO="__GITHUB_REPO__"
BINARY_NAME="standup"
STANDUP_BIN="${INSTALL_DIR}/${BINARY_NAME}"

ARCH=$(uname -m)
case "$ARCH" in
  arm64|aarch64)  SUFFIX="darwin-arm64" ;;
  x86_64)         SUFFIX="darwin-amd64" ;;
  *) exit 0 ;;
esac

CURRENT_VER=""
if [ -x "$STANDUP_BIN" ]; then
  CURRENT_VER=$("$STANDUP_BIN" -v 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1) || true
fi

LATEST_TAG=$(curl -sL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/') || true
LATEST_VER=$(echo "$LATEST_TAG" | sed 's/^v//')

if [ -z "$LATEST_VER" ]; then exit 0; fi
if [ -n "$CURRENT_VER" ] && [ "$CURRENT_VER" = "$LATEST_VER" ]; then exit 0; fi

DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}-${SUFFIX}"
TMP_FILE=$(mktemp)
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "$TMP_FILE" "$DOWNLOAD_URL" 2>/dev/null || { rm -f "$TMP_FILE"; exit 0; }
else
  exit 0
fi
chmod +x "$TMP_FILE"
mv "$TMP_FILE" "$STANDUP_BIN"
STANDUP_UPDATE_EOF
      chmod +x "$UPDATE_SCRIPT"

      PLIST="${HOME}/Library/LaunchAgents/com.standup.cli.update.plist"
      cat << PLIST_EOF > "$PLIST"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.standup.cli.update</string>
  <key>ProgramArguments</key>
  <array>
    <string>/bin/sh</string>
    <string>${UPDATE_SCRIPT}</string>
  </array>
  <key>StartCalendarInterval</key>
  <dict>
    <key>Weekday</key>
    <integer>0</integer>
    <key>Hour</key>
    <integer>3</integer>
    <key>Minute</key>
    <integer>0</integer>
  </dict>
  <key>StandardOutPath</key>
  <string>/dev/null</string>
  <key>StandardErrorPath</key>
  <string>/dev/null</string>
</dict>
</plist>
PLIST_EOF
      launchctl load "$PLIST" 2>/dev/null || true
      echo "Weekly auto-update enabled (runs every Sunday at 3:00 AM). To disable: launchctl unload $PLIST"
      ;;
    *) ;;
  esac
fi
