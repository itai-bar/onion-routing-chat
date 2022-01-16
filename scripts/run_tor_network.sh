#!/bin/bash

ARGC=$#
NODES_AMOUNT=$1

if [ $ARGC != 1 ]; then
    echo "Usage: ./run_tor_network [nodes amount]"
    exit 1
fi

if [ $NODES_AMOUNT -lt 3 ]; then
    echo "The network needs at least 3 nodes to start"
    exit 1
fi

echo "Cleaning old build"
docker-compose down > /dev/null 2>&1
echo "Building the services"
docker-compose build > /dev/null 2>&1


docker-compose up --scale node=$NODES_AMOUNT

echo "Saving Database"
docker cp torbasedchat_chat_server_1:/app/db.sqlite . # saving DB to host for next executions

# cleanup
echo "Cleaning the build"
docker-compose down > /dev/null 2>&1
echo "Removing unused docker data"
echo "y" | docker system prune > /dev/null 2>&1 