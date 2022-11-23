#!/usr/bin/env sh

# Workaround for Windows Support
printf '%s' "$SLACK_SCRIPT_NOTIFY" > "notify.sh"
chmod +x "notify.sh"
./notify.sh
rm -rf "notify.sh" || { printf '%s\n' "Could not clean the working directory." }