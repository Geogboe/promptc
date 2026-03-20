#!/bin/sh
# install.sh - Install promptc on Linux or macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/Geogboe/promptc/main/install.sh | sh
set -eu

REPO="Geogboe/promptc"
BINARY="promptc"

die() {
  echo "error: $*" >&2
  exit 1
}

# ── Detect OS ────────────────────────────────────────────────────────────────

OS="$(uname -s)"
case "${OS}" in
  Linux*)  OS="linux"  ;;
  Darwin*) OS="darwin" ;;
  *)       die "unsupported OS: ${OS}" ;;
esac

# ── Detect arch ──────────────────────────────────────────────────────────────

ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64)          ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *)               die "unsupported architecture: ${ARCH}" ;;
esac

# ── Fetch latest release tag ─────────────────────────────────────────────────

echo "Fetching latest release..."
TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
[ -n "${TAG}" ] || die "could not determine latest release tag"
echo "Latest release: ${TAG}"

# ── Construct URLs ────────────────────────────────────────────────────────────

ARCHIVE="${BINARY}_${TAG}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"
ARCHIVE_URL="${BASE_URL}/${ARCHIVE}"
CHECKSUMS_URL="${BASE_URL}/checksums.txt"

# ── Work in a temp dir ────────────────────────────────────────────────────────

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

# ── Download archive + checksums ─────────────────────────────────────────────

echo "Downloading ${ARCHIVE}..."
curl -fsSL -o "${TMP}/${ARCHIVE}" "${ARCHIVE_URL}" \
  || die "failed to download ${ARCHIVE_URL}"

echo "Downloading checksums..."
curl -fsSL -o "${TMP}/checksums.txt" "${CHECKSUMS_URL}" \
  || die "failed to download ${CHECKSUMS_URL}"

# ── Verify SHA256 ─────────────────────────────────────────────────────────────

echo "Verifying checksum..."
if command -v sha256sum >/dev/null 2>&1; then
  # Linux / GNU coreutils
  (cd "${TMP}" && grep "${ARCHIVE}" checksums.txt | sha256sum --check --status) \
    || die "checksum verification failed"
elif command -v shasum >/dev/null 2>&1; then
  # macOS / BSD
  (cd "${TMP}" && grep "${ARCHIVE}" checksums.txt | shasum -a 256 --check --status) \
    || die "checksum verification failed"
else
  die "no sha256sum or shasum found; cannot verify download"
fi
echo "Checksum OK."

# ── Extract binary ────────────────────────────────────────────────────────────

tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}" "${BINARY}" \
  || die "failed to extract binary from archive"

chmod +x "${TMP}/${BINARY}"

# ── Determine install directory ───────────────────────────────────────────────

if [ -n "${PROMPTC_INSTALL_DIR:-}" ]; then
  INSTALL_DIR="${PROMPTC_INSTALL_DIR}"
elif [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "${INSTALL_DIR}"
  # Warn if not in PATH
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      echo ""
      echo "NOTE: ${INSTALL_DIR} is not in your PATH."
      echo "Add this to your shell profile:"
      echo "  export PATH=\"\${HOME}/.local/bin:\${PATH}\""
      echo ""
      ;;
  esac
fi

# ── Install ───────────────────────────────────────────────────────────────────

cp "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

# ── Confirm ───────────────────────────────────────────────────────────────────

if command -v "${BINARY}" >/dev/null 2>&1; then
  "${BINARY}" --version
else
  "${INSTALL_DIR}/${BINARY}" --version
fi
