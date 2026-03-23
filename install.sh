#!/bin/sh
# install.sh - Install promptc on Linux or macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/Geogboe/promptc/main/install.sh | sh
set -eu

REPO="${PROMPTC_REPO:-Geogboe/promptc}"
BINARY="promptc"
API_BASE="${PROMPTC_RELEASES_API_BASE:-https://api.github.com/repos}"
RELEASE_TAG="${PROMPTC_RELEASE_TAG:-}"
USER_AGENT="promptc-install"

die() {
  echo "error: $*" >&2
  exit 1
}

require_tool() {
  command -v "$1" >/dev/null 2>&1 || die "required tool '$1' was not found"
}

require_tool curl
require_tool perl

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

fetch_release_json() {
  if [ -n "${RELEASE_TAG}" ]; then
    RELEASE_URL="${API_BASE}/${REPO}/releases/tags/${RELEASE_TAG}"
  else
    RELEASE_URL="${API_BASE}/${REPO}/releases/latest"
  fi

  echo "Fetching release metadata..." >&2
  curl -fsSL \
    -H "Accept: application/vnd.github+json" \
    -H "User-Agent: ${USER_AGENT}" \
    "${RELEASE_URL}" \
    || die "failed to fetch release metadata from ${RELEASE_URL}"
}

release_tag_from_json() {
  printf '%s' "$1" | perl -0ne 'if (/"tag_name"\s*:\s*"(.*?)"/s) { print $1; exit 0 }'
}

asset_count_from_json() {
  printf '%s' "$1" | perl -0ne 'print scalar(() = /"browser_download_url"\s*:/g), "\n"'
}

select_asset() {
  json="$1"
  regex="$2"
  description="$3"

  if ! result="$(
    printf '%s' "$json" | ASSET_REGEX="$regex" perl -0ne '
      my $pattern = $ENV{ASSET_REGEX};
      while (/"name"\s*:\s*"([^"]+)".*?"browser_download_url"\s*:\s*"([^"]+)"/sg) {
        if ($1 =~ /$pattern/) {
          print "$1\n$2\n";
          exit 0;
        }
      }
      exit 1;
    '
  )"; then
    die "could not find a ${description} asset matching '${regex}'"
  fi

  printf '%s' "$result"
}

RELEASE_JSON="$(fetch_release_json)"
TAG="$(release_tag_from_json "${RELEASE_JSON}")"
[ -n "${TAG}" ] || die "could not determine release tag"
echo "Release: ${TAG}"

ASSET_COUNT="$(asset_count_from_json "${RELEASE_JSON}")"
[ "${ASSET_COUNT}" -gt 0 ] || die "release '${TAG}' has no published assets"

# ── Work in a temp dir ────────────────────────────────────────────────────────

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

# ── Resolve release assets ────────────────────────────────────────────────────

ARCHIVE_INFO="$(select_asset "${RELEASE_JSON}" "^${BINARY}_.+_${OS}_${ARCH}\.tar\.gz$" "archive")"
ARCHIVE_NAME="$(printf '%s' "${ARCHIVE_INFO}" | sed -n '1p')"
ARCHIVE_URL="$(printf '%s' "${ARCHIVE_INFO}" | sed -n '2p')"

CHECKSUMS_INFO="$(select_asset "${RELEASE_JSON}" '^checksums\.txt$' "checksum file")"
CHECKSUMS_NAME="$(printf '%s' "${CHECKSUMS_INFO}" | sed -n '1p')"
CHECKSUMS_URL="$(printf '%s' "${CHECKSUMS_INFO}" | sed -n '2p')"

# ── Download archive + checksums ─────────────────────────────────────────────

echo "Downloading ${ARCHIVE_NAME}..."
curl -fsSL -o "${TMP}/${ARCHIVE_NAME}" "${ARCHIVE_URL}" \
  || die "failed to download ${ARCHIVE_URL}"

echo "Downloading ${CHECKSUMS_NAME}..."
curl -fsSL -o "${TMP}/${CHECKSUMS_NAME}" "${CHECKSUMS_URL}" \
  || die "failed to download ${CHECKSUMS_URL}"

# ── Verify SHA256 ─────────────────────────────────────────────────────────────

echo "Verifying checksum..."
EXPECTED_HASH="$(ASSET_NAME="${ARCHIVE_NAME}" perl -0ne '
  if (/^([A-Fa-f0-9]{64})\s+\Q$ENV{ASSET_NAME}\E$/m) {
    print "$1\n";
    exit 0;
  }
' "${TMP}/${CHECKSUMS_NAME}")"
[ -n "${EXPECTED_HASH}" ] || die "checksum entry for '${ARCHIVE_NAME}' not found in ${CHECKSUMS_NAME}"

if command -v sha256sum >/dev/null 2>&1; then
  # Linux / GNU coreutils
  ACTUAL_HASH="$(sha256sum "${TMP}/${ARCHIVE_NAME}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  # macOS / BSD
  ACTUAL_HASH="$(shasum -a 256 "${TMP}/${ARCHIVE_NAME}" | awk '{print $1}')"
else
  die "no sha256sum or shasum found; cannot verify download"
fi

[ "${ACTUAL_HASH}" = "${EXPECTED_HASH}" ] || die "checksum verification failed"
echo "Checksum OK."

# ── Extract binary ────────────────────────────────────────────────────────────

tar -xzf "${TMP}/${ARCHIVE_NAME}" -C "${TMP}" "${BINARY}" \
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
