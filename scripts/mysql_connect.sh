#!/usr/bin/env bash

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source "${script_dir}/env"

podman exec \
  --interactive \
  --tty \
  mysql \
  mysql \
  --host=localhost \
  --user=root \
  --password=root \
  deja_vu \
  "$@"
