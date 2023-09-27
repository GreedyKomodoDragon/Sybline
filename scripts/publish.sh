#!/bin/bash

while getopts ":n:pl" opt; do
  case $opt in
    n)
      image_name="$OPTARG"
      ;;
    p)
      push=true
      ;;
    l)
      latest=true
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

# Build the Docker container
docker build -t "$image_name" -f ./infra/docker/sybline.dockerfile .
docker build -t "$image_name-ubi" -f ./infra/docker/sybline.dockerfile .


# Push the container to Docker Hub if the push flag is set
if [ "$push" = true ]; then
    docker push "$image_name"
    docker push "$image_name-ubi"
    
    # Tag the container as "latest" if the flag is set
    if [ "$latest" = true ]; then
        docker tag "$image_name" "${image_name%:*}:latest"
        docker tag "$image_name-ubi" "${image_name%:*}-ubi:latest"
        
        docker push "${image_name%:*}:latest"
        docker push "${image_name%:*}-ubi:latest"
    fi
fi
