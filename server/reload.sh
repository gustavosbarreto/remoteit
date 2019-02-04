#!/bin/sh

docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d --force-recreate --no-deps --build $1
