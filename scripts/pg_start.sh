#!/usr/bin/env bash

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source "${script_dir}/env"

podman_id=$(podman ps --filter "name=postgresql" --filter "status=running" --quiet)

if [[ -z "$podman_id" ]]; then
  echo "${CYAN}Starting PostgreSQL...${NC}"
  if ! podman run \
    --detach \
    --env POSTGRES_PASSWORD=postgres \
    --health-cmd pg_isready \
    --health-interval 1s \
    --health-timeout 5s \
    --health-retries 5 \
    --name postgresql \
    --publish 5432:5432 \
    --rm \
    postgres:15.2 \
    >/dev/null; then
    echo "[${RED}ERROR${NC}] Failed to start PostgreSQL"
    exit 1
  fi
fi

echo "[${GREEN}OK${NC}] PostgreSQL is running"
