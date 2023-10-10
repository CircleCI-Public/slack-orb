#!/bin/sh

# Determine the http client to use
# Returns 1 if no HTTP client is found
determine_http_client() {
  if command -v curl >/dev/null 2>&1; then
    HTTP_CLIENT=curl
  elif command -v wget >/dev/null 2>&1; then
    HTTP_CLIENT=wget
  else
    return 1
  fi
}

# Download a binary file
# $1: The path to save the file to
# $2: The URL to download the file from
# $3: The HTTP client to use (curl or wget)
download_binary() {
  if [ "$3" = "curl" ]; then
    curl --fail --retry 3 -L -o "$1" "$2"
  elif [ "$3" = "wget" ]; then
    wget --tries=3 --timeout=10 --quiet -O "$1" "$2"
  else
    return 1
  fi
}

detect_os() {
  detected_platform="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$detected_platform" in
  linux*) PLATFORM=Linux ;;
  darwin*) PLATFORM=Darwin ;;
  msys* | cygwin*) PLATFORM=Windows ;;
  *) return 1 ;;
  esac
}

detect_arch() {
  detected_arch="$(uname -m)"
  case "$detected_arch" in
  x86_64) ARCH=x86_64 ;;
  i386 | i486 | i586 | i686) ARCH=x86 ;;
  arm64 | aarch64) ARCH=arm64 ;;
  arm*) ARCH=arm ;;
  *) return 1 ;;
  esac
}

# Determine the latest version of a GitHub release.
# $1: The GitHub organization
# $2: The GitHub repository
# $3: The HTTP client to use (curl or wget)
determine_release_latest_version() {
  url="https://github.com/$1/$2/releases/latest"

  if [ "$3" = "curl" ]; then
    LATEST_VERSION="$(curl --fail --retry 3 -Ls -o /dev/null -w '%{url_effective}' "https://github.com/$1/$2/releases/latest" | sed 's:.*/::')"
  elif [ "$3" = "wget" ]; then
    effective_url="$(wget --tries=3 --max-redirect=1000 --server-response -O /dev/null "$url" 2>&1 | awk '/Location: /{print $2}' | tail -1)"
    LATEST_VERSION="$(printf '%s' "$effective_url" | sed 's:.*/::')"
  else
    printf '%s\n' "Invalid HTTP client specified."
    return 1
  fi
}

base_dir="$(printf "%s" "$CIRCLE_WORKING_DIRECTORY" | sed "s|~|$HOME|")"
repo_org="CircleCI-Public"
repo_name="slack-orb-go"

# If the tag is empty, then we are building the Slack binary
# Therefore we will manually build and execute the binary for testing purposes
# Otherwise, we will download the binary from GitHub
binary=""
if [ "$SLACK_PARAM_DEVELOPER_MODE" -eq 1 ]; then
  binary="$repo_name"
  printf '%s\n' "Building $binary binary..."
  if ! go build -o "$binary" ./src/scripts/main.go; then
    printf '%s\n' "Failed to build $binary binary."
    exit 1
  fi
else
  if ! determine_http_client; then
    printf '%s\n' "cURL or wget is required to download the Slack binary."
    printf '%s\n' "Please install cURL or wget and try again."
    exit 1
  fi
  printf '%s\n' "HTTP client: $HTTP_CLIENT."

  if ! detect_os; then
    printf '%s\n' "Unsupported operating system: $(uname -s)."
    exit 1
  fi
  printf '%s\n' "Operating system: $PLATFORM."

  if ! detect_arch; then
    printf '%s\n' "Unsupported architecture: $(uname -m)."
    exit 1
  fi
  printf '%s\n' "Architecture: $ARCH."

  if ! determine_release_latest_version "$repo_org" "$repo_name" "$HTTP_CLIENT"; then
    printf '%s\n' "Failed to determine latest version."
    exit 1
  fi
  printf '%s\n' "Release's latest version: $LATEST_VERSION."

  # TODO: Make the version configurable via command parameter
  repo_url="https://github.com/$repo_org/$repo_name/releases/download/$LATEST_VERSION/$repo_name-$PLATFORM-$ARCH"
  printf '%s\n' "Release URL: $repo_url."

  # TODO: Check the md5sum of the downloaded binary
  binary_download_dir="$(mktemp -d -t "$repo_name".XXXXXXXXXX)"
  binary="$binary_download_dir/$repo_name"
  if ! download_binary "$binary" "$repo_url" "$HTTP_CLIENT"; then
    printf '%s\n' "Failed to download $repo_name binary from GitHub."
    exit 1
  fi

  printf '%s\n' "Downloaded $repo_name binary to $binary_download_dir"
fi

printf '%s\n' "Making $binary binary executable..."
if ! chmod +x "$binary"; then
  printf '%s\n' "Failed to make $binary binary executable."
  exit 1
fi

printf '%s\n' "Executing \"$base_dir/$binary\" binary..."
set -x
"$base_dir/$binary"
exit_code=$?
set +x
if [ $exit_code -ne 0 ]; then
  printf '%s\n' "Failed to execute $binary binary or it exited with a non-zero exit code."
fi

printf '%s\n' "Removing $binary binary..."
rm -rf "$binary"
exit $exit_code