#!/bin/bash
set -e
docker-compose up -d db
echo "Waiting for database to be ready..."
until docker-compose exec -T db pg_isready -U postgres; do
  sleep 5
done
docker-compose up -d --build app
