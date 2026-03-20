#!/bin/zsh

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Next Work Project
# @raycast.mode silent

# Optional parameters:
# @raycast.icon ⏭️
# @raycast.packageName WorkDay

export PATH="$HOME/bin:$HOME/.local/bin:$HOME/go/bin:/urs/local/bin:/opt/homebrew/bin:$PATH"

RESULT=$(trak next 2>&1)
if [ $? -eq 0 ]; then
  echo "$RESULT"
else
  echo "❌ $RESULT"
  exit 1
fi
