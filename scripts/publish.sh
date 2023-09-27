#!/bin/bash

while getopts ":n:t:pl" opt; do
  case $opt in
    n)
      image_name="$OPTARG"
      ;;
    t)
      tag="$OPTARG"
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
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

# Check if the tag parameter is provided
if [ -z "$tag" ]; then
  echo "Tag parameter (-t) is required."
  exit 1
fi

# Build the Docker container with the specified tag
docker build -t "$image_name:$tag" -f ./infra/docker/sybline.dockerfile .
docker build -t "$image_name-ubi:$tag" -f ./infra/docker/UBI.dockerfile .

# Push the containers to Docker Hub if the push flag is set
if [ "$push" = true ]; then
    docker push "$image_name:$tag"
    docker push "$image_name-ubi:$tag"
    
    # Tag the containers as "latest" if the flag is set
    if [ "$latest" = true ]; then
        docker tag "$image_name:$tag" "$image_name:latest"
        docker tag "$image_name-ubi:$tag" "$image_name-ubi:latest"
        docker push "$image_name:latest"
        docker push "$image_name-ubi:latest"
    fi
fi
