#!/usr/bin/env bash

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source "${script_dir}/env"

podman_id=$(podman ps --filter "name=mysql" --filter "status=running" --quiet)

if [[ -z "$podman_id" ]]; then
  echo "${CYAN}Starting MySQL...${NC}"
  if ! podman run \
    --detach \
    --env MYSQL_DATABASE=deja_vu \
    --env MYSQL_ROOT_PASSWORD=root \
    --name mysql \
    --publish 3306:3306 \
    --rm \
    mysql:8.0 \
    --character-set-server=utf8mb4 \
    --collation-server=utf8mb4_unicode_ci \
    >/dev/null; then
    echo "[${RED}ERROR${NC}] Failed to start MySQL"
    exit 1
  fi
fi

wait_cmd MySQL "${script_dir}/mysql_connect.sh --execute \s --silent"

echo "[${GREEN}OK${NC}] MySQL is running"
