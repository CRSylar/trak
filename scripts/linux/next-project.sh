#!/bin/bash
#
# trak — next-project.sh (Linux)
#
# Cycles to the next work project and shows a notification.
# Bind this script to a hotkey in your WM/DE settings.
#
# Dependencies: trak, notify-send (libnotify)

export PATH="$HOME/bin:$HOME/.local/bin:$HOME/go/bin:/usr/local/bin:/opt/homebrew/bin:$PATH"

RESULT=$(trak next 2>&1)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
  notify-send --icon=clock --expire-time=2000 "trak" "$RESULT"
else
  notify-send --icon=dialog-error --expire-time=3000 "trak error" "$RESULT"
fi
