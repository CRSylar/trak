#!/bin/zsh

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Rest
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 🛋️
# @raycast.packageName WorkDay

export PATH="$HOME/bin:$HOME/.local/bin:$HOME/go/bin:/usr/local/bin:/opt/homebrew/bin:$PATH"

RESULT=$(trak rest 2>&1)
if [ $? -eq 0 ]; then
  echo "$RESULT"
else
  echo "❌ $RESULT"
  exit 1
fi
