#!/bin/bash

# Build the Docker image (this will run tests as part of the build process)
echo "Building Docker image with tests..."
docker build -f Dockerfile -t docker-gs-ping-test --progress plain --no-cache --target run-test-stage .

# Build the production image
echo "Building production Docker image..."
docker build -f Dockerfile -t kvrpc-server --progress plain --no-cache .

# Run the server container
echo "Starting kvrpc server container..."
docker run --rm -p 8080:8080 --name kvrpc-server kvrpc-server

