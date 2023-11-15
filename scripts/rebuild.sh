#!/bin/bash

set -e # Exit immediately if a command fails

# Default Docker Compose file
DOCKER_COMPOSE_FILE="infra/docker/docker-compose-local.yml"

# Parse command line options
while getopts "f:" opt; do
  case ${opt} in
    f )
      DOCKER_COMPOSE_FILE="$OPTARG"
      ;;
    \? )
      echo "Usage: $0 [-f docker-compose-file]" >&2
      exit 1
      ;;
  esac
done

echo "Removing node_* directories..."
rm -rf infra/docker/node_*
rm -rf infra/docker/cache*

echo "Stopping Docker Compose..."
docker-compose -f "$DOCKER_COMPOSE_FILE" down

echo "Building Docker image..."
docker build -t sybline:latest -f infra/docker/sybline.dockerfile .

echo "Cleaning up unused Docker images..."
docker image prune -f

echo "Starting Docker Compose..."
docker-compose -f "$DOCKER_COMPOSE_FILE" up
