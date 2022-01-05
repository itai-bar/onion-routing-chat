#!/bin/bash
docker-compose down
docker-compose up --build --scale node=$1

# cleanup
docker-compose down
docker system prune