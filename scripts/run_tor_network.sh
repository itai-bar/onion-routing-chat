#!/bin/bash
docker-compose down
docker-compose up --build

# cleanup
docker-compose down
docker system prune