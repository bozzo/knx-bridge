#!/bin/bash
#
#    Copyright 2020 Boris Barnier <bozzo@users.noreply.github.com>
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.
#

mkdir -vp ~/.docker/cli-plugins/

# Set docker platforms for which to build.
export DOCKER_PLATFORMS='linux/amd64 linux/arm64 linux/arm/v7'
export DOCKER_BASE="$DOCKER_IMAGE"

case $TRAVIS_BRANCH in
master|main)
  export TAGS="latest $TRAVIS_BRANCH"
  ;;
*)
  export TAGS="$TRAVIS_BRANCH"
  ;;
esac

source ./scripts/multi-arch-docker-ci.sh
set -ex; multi_arch_docker::main; set +x