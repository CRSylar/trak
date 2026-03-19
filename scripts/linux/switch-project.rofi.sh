#!/bin/bash
#
# trak — switch-project.rofi.sh (Linux / rofi)
#
# Shows a rofi menu with all registered projects.
# Select one with keyboard, trak switches to it.
# Bind this script to a hotkey in your WM/DE settings.
#
# Dependencies: trak, rofi, notify-send (libnotify)

export PATH="$HOME/bin:$HOME/.local/bin:$HOME/go/bin:/urs/local/bin:/opt/homebrew/bin:$PATH"

PROJECTS_JSON=$(trak projects --names 2>/dev/null)

if [ $? -ne 0 ] || [ -z "$PROJECTS_JSON" ]; then
  notify-send --icon=dialog-error --expire-time=3000 "trak error" \
    "Daemon not running — start your workday with 'trak start'"
  exit 1
fi

PROJECTS=$(echo "$PROJECTS_JSON" | python3 -c "
import sys, json
for p in json.load(sys.stdin):
    print(p)
")

CHOSEN=$(echo "$PROJECTS" | rofi \
  -dmenu \
  -i \
  -p "switch to" \
  -theme-str 'window {width: 300px;}' \
  -no-custom)

[ -z "$CHOSEN" ] && exit 0

RESULT=$(trak switch "$CHOSEN" 2>&1)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
  notify-send --icon=clock --expire-time=2000 "trak" "$RESULT"
else
  notify-send --icon=dialog-error --expire-time=3000 "trak error" "$RESULT"
fi
