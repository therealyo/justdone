#! /bin/bash
docker compose down
docker compose up -d --build
docker compose exec -it justdone /bin/sh
