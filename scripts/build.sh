#!/usr/bin/env bash
# Copyright (C) 2024, AllianceBlock. All rights reserved.
# See the file LICENSE for licensing terms.

set -o errexit
set -o nounset
set -o pipefail

# Function to build the binary locally
build_source() {
    # Set the CGO flags to use the portable version of BLST
    export CGO_CFLAGS="-O -D__BLST_PORTABLE__" CGO_ENABLED=1

    # Root directory
    ROOT_PATH=$(
        cd "$(dirname "${BASH_SOURCE[0]}")"
        cd .. && pwd
    )

    # Set default binary directory location
    FEED_PATH=$ROOT_PATH/build/nuklai-feed

    echo "Building nuklai-feed in $FEED_PATH"
    mkdir -p "$(dirname "$FEED_PATH")"
    go build -o "$FEED_PATH" ./
}

# Function to build the Docker image
build_docker() {
    ROOT_PATH=$(
        cd "$(dirname "${BASH_SOURCE[0]}")"
        cd .. && pwd
    )

    echo "Building Docker image for nuklai-feed"
    docker build -t nuklai-feed "$ROOT_PATH"
}

# Check for argument and call the appropriate function
if [ $# -eq 0 ]; then
    build_source
else
    case "$1" in
        docker)
            build_docker
            ;;
        source)
            build_source
            ;;
        *)
            echo "Invalid build type specified. Usage: $0 {docker|source}"
            exit 1
            ;;
    esac
fi
