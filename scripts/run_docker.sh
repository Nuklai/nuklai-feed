#!/usr/bin/env bash
# Copyright (C) 2024, AllianceBlock. All rights reserved.
# See the file LICENSE for licensing terms.

docker container rm -f nuklai-feed || true

# Check if .env file exists
if [ ! -f .env ]; then
  echo ".env file not found!"
  exit 1
fi

# Read the .env file and construct the --env options for docker run
env_vars=$(grep -v '^#' .env | xargs -I {} echo --env {} | xargs)

# Run the docker container with the constructed --env options
docker run -d -p 10592:10592 --name nuklai-feed $env_vars nuklai-feed

# Print the logs
  docker container logs -f nuklai-feed