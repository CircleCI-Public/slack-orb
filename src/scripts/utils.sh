#!/bin/false
# shellcheck shell=bash
# shellcheck disable=SC2154

detect_os() {
  detected_platform="$(uname -s | tr '[:upper:]' '[:lower:]')"

  case "$detected_platform" in
    linux*)
      PLATFORM=linux
      ;;
    darwin*)
      PLATFORM=macos
      ;;
    msys*|cygwin*)
      PLATFORM=windows
      ;;
    *)
      printf '%s\n' "Unsupported OS: \"$platform\"."
      exit 1
      ;;
  esac
  readonly PLATFORM
  export PLATFORM
}
