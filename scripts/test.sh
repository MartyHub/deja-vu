#!/usr/bin/env sh

set -eux

test='./...'

if [[ -n "$@" ]]; then
  test="-run $@"
fi

gotest -coverprofile coverage.out -race -timeout 30s ${test}
