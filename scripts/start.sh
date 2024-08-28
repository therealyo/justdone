#!/bin/bash

set -e

cp .env.example .env

docker-compose -f compose.prod.yaml down

docker build -t justdone .

docker-compose -f compose.prod.yaml up -d

# Install goose for migrations
# docker-compose -f compose.prod.yaml exec justdone go install github.com/pressly/goose/v3/cmd/goose@latest
# 
# Run migrations
# docker-compose -f compose.prod.yaml exec justdone goose -dir=migrations up

# Print application logs
# docker-compose -f compose.prod.yaml logs -f justdone
docker-compose exec -it justdone /bin/sh
