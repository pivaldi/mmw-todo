#!/usr/bin/env bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
readonly SCRIPT_DIR

source "$SCRIPT_DIR/libs/lobash.bash" || exit 1
source "$SCRIPT_DIR/libs/log.rc" || exit 1

spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        # \r moves to the start of the line, but stdout will push it down
        # This works best if your command doesn't output MANY lines per second
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

#!/bin/bash
# Note: Requires Bash 4.0+ for fractional timeouts (standard on most Linux systems).

run_with_spinner() {
    local spinstr='|/-\'
    local i=0

    # Define how far to shift the LOGS (e.g., 4 spaces)
    local indent="    "

    tput civis

    while true; do
        if IFS= read -r -t 0.1 line; then
            # LOG LINE: Add the indent here so the text shifts right
            printf "\r\e[2K%s%s\n" "$indent" "$line"

        elif [[ $? -gt 128 ]]; then
            i=$(((i + 1) % 4))
            # SPINNER: No indent here! Pinned to the left edge.
            printf "\r\e[2K[%s] Working..." "${spinstr:$i:1}"

        else
            break
        fi
    done

    # Clean up the final line
    printf "\r\e[2K"
    tput cnorm
}
# ==========================================
# Usage Example
# ==========================================

# Group your commands in a block { ... } and pipe them to the function
# {
#     echo "Starting process..."
#     sleep 1
#     echo "Running gosec G703 checks..."
#     sleep 2
#     echo "Compiling..."
#     sleep 1
#     echo "Success!"
# } 2>&1 | run_with_spinner
