#!/usr/bin/env bash

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source "${script_dir}/env"

podman_id=$(podman ps --filter "name=mysql" --filter "status=running" --quiet)

if [[ -n "$podman_id" ]]; then
  echo "${CYAN}Stopping MySQL...${NC}"
  if ! podman stop mysql >/dev/null; then
    echo "[${RED}ERROR${NC}] Failed to stop MySQL"
    exit 1
  fi
fi

echo "[${GREEN}OK${NC}] MySQL is stopped"
