#!/usr/bin/env bash
# Build a docker image using Dockerfile.optimized
# usage:
# ./build-docker.sh {amber|mainnet}

set -eux

NETWORK=${1:-"mainnet"}
if [ $NETWORK == "amber" ]; then
    docker build -f Dockerfile.optimized \
      --build-arg SUPERBCH_BUILD_TAGS='cppbtree,params_amber' \
      --build-arg CONFIG_VERSION=v0.0.4 \
      --build-arg CHAIN_ID=0x2711 -t superbch-amber:latest .
elif [ $NETWORK == "mainnet" ]; then
    docker build -f Dockerfile.optimized -t superbch:latest .
else
    echo "Invalid argument, usage: ./build-docker.sh {amber|mainnet}"
fi
