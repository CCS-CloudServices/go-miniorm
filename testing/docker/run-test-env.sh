#!/usr/bin/env bash

set -e
shopt -s expand_aliases

export GBS_GELF_IP=localhost
export GBS_GELF_PORT=0
export DOCKER_BUILD_IMAGE=bash

CUR_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "${CUR_SCRIPT_DIR}"

alias docker_compose='docker-compose -f docker-compose.yml -f docker-compose-test-env.yml'

function docker_cleanup() {
  docker_compose down --remove-orphans
  docker_compose rm -sfv # remove any leftovers
}
trap docker_cleanup EXIT

docker_cleanup
docker_compose up --remove-orphans --scale integration-tests=0
