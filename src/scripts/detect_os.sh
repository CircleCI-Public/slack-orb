#!/usr/bin/env bash
detected_platform="$(uname -s | tr '[:upper:]' '[:lower:]')"

case "$detected_platform" in
linux*)
    if grep "Alpine" /etc/issue >/dev/null 2>&1; then
        printf '%s\n' "Detected OS: Alpine Linux."
        SYS_ENV_PLATFORM=linux_alpine
    else
        printf '%s\n' "Detected OS: Linux."
        SYS_ENV_PLATFORM=linux
    fi  
    ;;
darwin*)
    printf '%s\n' "Detected OS: macOS."
    SYS_ENV_PLATFORM=macos
    ;;
msys*|cygwin*)
    printf '%s\n' "Detected OS: Windows."
    SYS_ENV_PLATFORM=windows
    ;;
*)
    printf '%s\n' "Unsupported OS: \"$detected_platform\"."
    exit 1
    ;;
esac

echo "export INTERNAL_PARAM_EXECUTOR=$SYS_ENV_PLATFORM" >> "$BASH_ENV"