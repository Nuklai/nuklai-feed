#!/usr/bin/env bash
# Copyright (C) 2024, AllianceBlock. All rights reserved.
# See the file LICENSE for licensing terms.

# Check if .env file exists
if [ ! -f .env ]; then
  echo ".env file not found!"
  exit 1
fi

# Source the .env file to load environment variables
source .env

# Read the .env file and construct the --env options for docker run
env_vars=$(grep -v '^#' .env | xargs -I {} echo --env {} | xargs)

# Function to create a custom Docker network
function create_network() {
  echo "Creating custom Docker network..."
  docker network create nuklai-feed-network || true
}

# Function to start the PostgreSQL container
function start_postgres() {
  echo "Starting PostgreSQL container..."

  # Remove any existing data volume to ensure clean initialization
  docker volume rm postgres_data_feed || true

  # Run the PostgreSQL container with the constructed --env options
  docker run -d --name nuklai-feed-postgres --network nuklai-feed-network \
      --env POSTGRES_USER=${POSTGRES_USER} \
      --env POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
      --env POSTGRES_DBNAME=${POSTGRES_DBNAME} \
      -p ${POSTGRES_PORT:-5432}:5432 \
      -v postgres_data_feed:/var/lib/postgresql/data \
      -v $(pwd)/docker-entrypoint-initdb.d:/docker-entrypoint-initdb.d \
      postgres:13

  echo "Waiting for PostgreSQL to become healthy..."
  until docker exec nuklai-feed-postgres pg_isready -U $POSTGRES_USER -d $POSTGRES_DBNAME; do
    echo "PostgreSQL is unavailable - sleeping"
    sleep 1
  done

  echo "PostgreSQL is up and running"
}

# Function to start the Feed container
function start_feed() {
  echo "Starting Feed container..."

# Ensure PostgreSQL is fully ready before starting Feed
  echo "Checking if PostgreSQL is ready..."
  until docker exec nuklai-feed-postgres pg_isready -U $POSTGRES_USER -d $POSTGRES_DBNAME; do
    echo "PostgreSQL is unavailable - sleeping"
    sleep 1
  done

  # Run the Feed container with the constructed --env options
  docker run -d -p 10592:10592 --name nuklai-feed --network nuklai-feed-network $env_vars nuklai-feed

  echo "Feed container started"
}

# Function to stop and remove the containers
function stop_services() {
  echo "Stopping Feed container..."
  docker stop nuklai-feed || true
  docker rm nuklai-feed || true

  echo "Stopping PostgreSQL container..."
  docker stop nuklai-feed-postgres || true
  docker rm nuklai-feed-postgres || true

  echo "Removing custom network..."
  docker network rm nuklai-feed-network || true
}

case "$1" in
  start)
    stop_services
    create_network
    start_postgres
    start_feed
    ;;
  stop)
    stop_services
    ;;
  logs)
    docker logs -f nuklai-feed
    ;;
  *)
    echo "Usage: $0 {start|stop|logs}"
    ;;
esac
