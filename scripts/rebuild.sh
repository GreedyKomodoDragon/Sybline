#!/bin/bash

set -e # Exit immediately if a command fails

echo "Removing node_* directories..."
rm -rf infra/docker/node_*
rm -rf infra/docker/cache*

echo "Stopping Docker Compose..."
docker-compose -f infra/docker/docker-compose-local.yml down

echo "Building Docker image..."
docker build -t sybline:latest -f infra/docker/sybline.dockerfile . 

echo "Cleaning up unused Docker images..."
docker image prune -f

echo "Starting Docker Compose..."
docker-compose -f infra/docker/docker-compose-local.yml up