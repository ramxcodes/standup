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

# Interactive: let user choose install location (or use default)
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
if [ -t 0 ]; then
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

# Remind to reload shell if we added something to .zshrc
case "$INSTALL_DIR" in
  /usr/local/bin|/opt/homebrew/bin) echo "Run: standup" ;;
  *) echo "Run \`source ~/.zshrc\` or open a new terminal, then run: standup" ;;
esac
