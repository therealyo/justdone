#! /bin/bash
docker compose down
docker compose up -d
docker compose exec -it justdone /bin/sh
