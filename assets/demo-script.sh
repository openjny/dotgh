#!/bin/bash
# Demo script for asciinema recording

# Typing simulation
type_slow() {
  local text="$1"
  for ((i=0; i<${#text}; i++)); do
    printf '%s' "${text:$i:1}"
    sleep 0.08
  done
}

# Wait and show prompt
prompt() {
  sleep 0.3
  printf '\n$ '
  sleep 0.5
}

clear
printf '$ '
sleep 0.5

# 0. Show empty directory
type_slow "ls -la"
sleep 0.3
printf '\n'
printf 'total 0\n'
prompt

# 1. List templates
type_slow "dotgh list"
sleep 0.3
printf '\n'
dotgh list
prompt

# 2. Pull a template
type_slow "dotgh pull bicep-mslearn-mcp"
sleep 0.3
printf '\n'
dotgh pull bicep-mslearn-mcp
prompt

# 3. Show files created
type_slow "ls -la"
sleep 0.3
printf '\n'
echo 'total 16'
echo 'drwxr-xr-x  3 user user 4096 Dec  1 12:00 .'
echo 'drwxr-xr-x 10 user user 4096 Dec  1 12:00 ..'
echo 'drwxr-xr-x  4 user user 4096 Dec  1 12:00 .github'
echo 'drwxr-xr-x  2 user user 4096 Dec  1 12:00 .vscode'
echo '-rw-r--r--  1 user user  512 Dec  1 12:00 AGENTS.md'

# End with prompt visible for a moment
printf '\n$ '
sleep 0.5
printf ' '
sleep 0.5
printf ' '
sleep 0.5
printf ' '
sleep 0.5
printf ' '
sleep 0.5
printf ' '
sleep 0.5
printf ' '
sleep 0.5
