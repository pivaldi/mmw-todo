#!/usr/bin/env bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

SCRIPT="$SCRIPT_DIR/install-tools.sh"
STEPPS=$(grep -c 'st.done' "$SCRIPT")

stream-stepper --wait=2 --processor=stbash --steps="$STEPPS" "$SCRIPT"
