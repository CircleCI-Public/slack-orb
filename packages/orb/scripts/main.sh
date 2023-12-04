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
    set -x
    curl --fail --retry 3 -L -o "$1" "$2"
    set +x
  elif [ "$3" = "wget" ]; then
    set -x
    wget --tries=3 --timeout=10 --quiet -O "$1" "$2"
    set +x
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
  x86_64 | amd64) ARCH=x86_64 ;;
  i386 | i486 | i586 | i686) ARCH=i386 ;;
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
  if [ "$3" = "curl" ]; then
    LATEST_VERSION="$(curl --fail --retry 3 -Ls -o /dev/null -w '%{url_effective}' "https://github.com/$1/$2/releases/latest" | sed 's:.*/::')"
  elif [ "$3" = "wget" ]; then
    LATEST_VERSION="$(wget -qO- "https://api.github.com/repos/$1/$2/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')"
  else
    printf '%s\n' "Invalid HTTP client specified."
    return 1
  fi
}

# Print a warning message
# $1: The warning message to print
print_warn() {
  yellow="\033[1;33m"
  normal="\033[0m"
  printf "${yellow}%s${normal}\n" "$1"
}

# Print a success message
# $1: The success message to print
print_success() {
  green="\033[0;32m"
  normal="\033[0m"
  printf "${green}%s${normal}\n" "$1"
}

# Print an error message
# $1: The error message to print
print_error() {
  red="\033[0;31m"
  normal="\033[0m"
  printf "${red}%s${normal}\n" "$1"
}

# Print a debug message
# $1: The debug message to print
print_debug() {
    if [ "$SLACK_PARAM_DEBUG" = 1 ]; then
        # ANSI escape code for blue color
        BLUE="\033[0;34m"
        # ANSI escape code to reset color
        RESET="\033[0m"
        
        printf "${BLUE}DEBU ${RESET} %s\n" "$1"
    fi
}

print_warn "This is an experimental version of the Slack Orb in Go."
print_warn "Thank you for trying it out and please provide feedback to us at https://github.com/CircleCI-Public/slack-orb-go/issues"

if ! detect_os; then
  printf '%s\n' "Unsupported operating system: $(uname -s)."
  exit 1
fi
print_debug "Operating system: $PLATFORM."

if ! detect_arch; then
  printf '%s\n' "Unsupported architecture: $(uname -m)."
  exit 1
fi

print_debug "Architecture: $ARCH."

base_dir="$(printf "%s" "$CIRCLE_WORKING_DIRECTORY" | sed "s|~|$HOME|")"
orb_bin_dir="$base_dir/.circleci/orbs/circleci/slack/$PLATFORM/$ARCH"
bin_name="slack-orb-go"
repo_org="CircleCI-Public"
repo_name="slack-orb-go"
binary="$orb_bin_dir/$bin_name"
input_sha256=$(circleci env subst "$SLACK_PARAM_SHA256")

if [ ! -f "$binary" ]; then
  mkdir -p "$orb_bin_dir"
  if ! determine_http_client; then
    printf '%s\n' "cURL or wget is required to download the Slack binary."
    printf '%s\n' "Please install cURL or wget and try again."
    exit 1
  fi
  printf '%s\n' "HTTP client: $HTTP_CLIENT."

  if ! determine_release_latest_version "$repo_org" "$repo_name" "$HTTP_CLIENT"; then
    printf '%s\n' "Failed to determine latest version."
    exit 1
  fi
  printf '%s\n' "Release's latest version: $LATEST_VERSION."

  # TODO: Make the version configurable via command parameter
  repo_url="https://github.com/$repo_org/$repo_name/releases/download/$LATEST_VERSION/${repo_name}_${PLATFORM}_${ARCH}"
  [ "$PLATFORM" = "Windows" ] && repo_url="$repo_url.exe"
  printf '%s\n' "Release URL: $repo_url."

  if ! download_binary "$binary" "$repo_url" "$HTTP_CLIENT"; then
    printf '%s\n' "Failed to download $repo_name binary from GitHub."
    exit 1
  fi

  print_debug "Downloaded $repo_name binary to $orb_bin_dir"
else
  print_debug "Skipping binary download since it already exists at $binary."
fi

# Validate binary
## This validates, even if the binary already existed before.
## This can help with cache integrity but was also a convenience for testing where the binary will never be downloaded.
if [ -n "$input_sha256" ]; then
  actual_sha256=""
  if [ "$PLATFORM" = "Windows" ]; then
    actual_sha256=$(powershell.exe -Command "(Get-FileHash -Path '$binary' -Algorithm SHA256).Hash.ToLower()")
  else
    actual_sha256=$(sha256sum "$binary" | cut -d' ' -f1)
  fi

  if [ "$actual_sha256" != "$input_sha256" ]; then
    print_error "SHA256 checksum does not match. Expected $input_sha256 but got $actual_sha256"
    exit 1
  else
    print_success "SHA256 checksum matches. Binary is valid."
  fi
else
  print_warn "SHA256 checksum not provided. Skipping validation."
fi

print_debug "Making $binary binary executable..."
if ! chmod +x "$binary"; then
  printf '%s\n' "Failed to make $binary binary executable."
  exit 1
fi

print_debug "Executing \"$binary\" binary..."
"$binary" notify
exit_code=$?
if [ $exit_code -ne 0 ]; then
  printf '%s\n' "Failed to execute $binary binary or it exited with a non-zero exit code."
fi

exit $exit_code
