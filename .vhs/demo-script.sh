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

sleep 1.5
