#!/bin/bash

echo "Cleaning up containers and volumes..."

# Stop and remove containers
docker compose down

# Remove the database volume to start fresh
docker volume rm lang_bot_dbdata 2>/dev/null || true

# Remove any existing containers
docker rm -f go-app go-db 2>/dev/null || true

echo "Starting fresh..."

# Start the services
docker compose up --build 